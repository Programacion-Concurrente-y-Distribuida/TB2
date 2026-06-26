// Comando migrate-mongo: migra de forma concurrente el dataset CSV de AQS
// (por defecto aqs_final_3M.csv) hacia una coleccion de MongoDB.
//
// Diseno concurrente (pipeline productor/consumidor):
//
//	reader (1 goroutine) --lee CSV, arma lotes--> canal batches
//	                                                  |
//	                          N workers --InsertMany bulk--> MongoDB
//
// El UNICO parametro obligatorio es -mongo-url. El resto tiene valores por
// defecto razonables.
//
// Uso minimo:
//
//	go run ./cmd/migrate-mongo -mongo-url "mongodb://localhost:27017"
package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// columnKind clasifica como debe convertirse cada columna al insertar en Mongo.
type columnKind int

const (
	kindString columnKind = iota // texto (incluye codigos con ceros a la izquierda)
	kindInt                      // entero
	kindFloat                    // flotante (acepta nan/vacio -> null)
	kindBool                     // booleano (True/False)
)

// columnKinds mapea el nombre de cada columna del CSV a su tipo destino.
// Los codigos (State/County/Site) se guardan como string para no perder los
// ceros a la izquierda (ej. "01", "003", "0010").
var columnKinds = map[string]columnKind{
	"State Code":         kindString,
	"County Code":        kindString,
	"Site Num":           kindString,
	"POC":                kindInt,
	"Parameter Code":     kindInt,
	"Year":               kindInt,
	"Latitude":           kindFloat,
	"Longitude":          kindFloat,
	"Parameter Name":     kindString,
	"Sample Duration":    kindString,
	"Pollutant Standard": kindString,
	"Units of Measure":   kindString,
	"Event Type":         kindString,
	"State Name":         kindString,
	"County Name":        kindString,
	"City Name":          kindString,
	"CBSA Name":          kindString,
	"pollutant":          kindString,
	"is_synthetic":       kindBool,
	// El resto de columnas (conteos, medias y percentiles) son kindFloat por
	// defecto via kindFor().
}

// kindFor devuelve el tipo destino de una columna; las no listadas se asumen
// numericas flotantes (todos los conteos/medias/percentiles del dataset).
func kindFor(col string) columnKind {
	if k, ok := columnKinds[col]; ok {
		return k
	}
	return kindFloat
}

type config struct {
	mongoURL   string
	inputPath  string
	database   string
	collection string
	batchSize  int
	workers    int
	chanBuffer int
	dropFirst  bool
	limit      int
}

func main() {
	cfg := config{}

	flag.StringVar(&cfg.mongoURL, "mongo-url", "", "(OBLIGATORIO) URL de conexion a MongoDB, ej. mongodb://localhost:27017")
	flag.StringVar(&cfg.inputPath, "input", "aqs_final_3M.csv", "Ruta del CSV a migrar")
	flag.StringVar(&cfg.database, "db", "aqs", "Nombre de la base de datos destino")
	flag.StringVar(&cfg.collection, "collection", "measurements", "Nombre de la coleccion destino")
	flag.IntVar(&cfg.batchSize, "batch-size", 2000, "Filas por lote de InsertMany")
	flag.IntVar(&cfg.workers, "workers", runtime.NumCPU(), "Numero de workers concurrentes de insercion")
	flag.IntVar(&cfg.chanBuffer, "chan-buffer", 0, "Buffer del canal de lotes (0 = 2x workers)")
	flag.BoolVar(&cfg.dropFirst, "drop", false, "Eliminar la coleccion antes de migrar")
	flag.IntVar(&cfg.limit, "limit", 0, "Maximo de filas a migrar (0 = todas)")
	flag.Parse()

	if cfg.mongoURL == "" {
		fmt.Fprintln(os.Stderr, "error: -mongo-url es obligatorio")
		flag.Usage()
		os.Exit(2)
	}
	if cfg.workers < 1 {
		cfg.workers = 1
	}
	if cfg.chanBuffer <= 0 {
		cfg.chanBuffer = cfg.workers * 2
	}

	if err := run(cfg); err != nil {
		log.Fatalf("migracion fallida: %v", err)
	}
}

func run(cfg config) error {
	// Contexto cancelable por Ctrl+C / SIGTERM para un cierre ordenado.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// --- Conexion a MongoDB ---
	clientOpts := options.Client().
		ApplyURI(cfg.mongoURL).
		SetRetryWrites(true)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return fmt.Errorf("conectando a mongo: %w", err)
	}
	defer func() {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = client.Disconnect(disconnectCtx)
	}()

	pingCtx, cancelPing := context.WithTimeout(ctx, 10*time.Second)
	defer cancelPing()
	if err := client.Ping(pingCtx, nil); err != nil {
		return fmt.Errorf("ping a mongo: %w", err)
	}
	log.Printf("conectado a MongoDB; destino %s.%s", cfg.database, cfg.collection)

	coll := client.Database(cfg.database).Collection(cfg.collection)

	if cfg.dropFirst {
		if err := coll.Drop(ctx); err != nil {
			return fmt.Errorf("drop coleccion: %w", err)
		}
		log.Printf("coleccion %s.%s eliminada antes de migrar", cfg.database, cfg.collection)
	}

	// --- Apertura del CSV ---
	f, err := os.Open(cfg.inputPath)
	if err != nil {
		return fmt.Errorf("abriendo CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.ReuseRecord = true  // reutiliza el buffer interno: menos allocs por fila
	reader.FieldsPerRecord = 0 // fija el numero de campos segun el header

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("leyendo header: %w", err)
	}
	// Copia del header: ReuseRecord invalida el slice en la siguiente lectura.
	cols := make([]string, len(header))
	copy(cols, header)
	kinds := make([]columnKind, len(cols))
	for i, c := range cols {
		kinds[i] = kindFor(c)
	}
	log.Printf("CSV abierto: %d columnas detectadas", len(cols))

	// --- Pipeline concurrente ---
	batches := make(chan []any, cfg.chanBuffer)

	var (
		inserted    atomic.Int64
		failedRows  atomic.Int64
		skippedRows atomic.Int64
		wg          sync.WaitGroup
		workerErr   atomic.Pointer[error]
	)

	start := time.Now()

	// Workers consumidores: cada uno hace InsertMany no-ordenado (mas rapido y
	// resiliente: un documento malo no aborta el lote completo).
	insertOpts := options.InsertMany().SetOrdered(false)
	for w := 0; w < cfg.workers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for batch := range batches {
				if ctx.Err() != nil {
					return
				}
				res, err := coll.InsertMany(ctx, batch, insertOpts)
				if res != nil {
					inserted.Add(int64(len(res.InsertedIDs)))
				}
				if err != nil {
					// Con escrituras no-ordenadas, un BulkWriteException reporta
					// los documentos que si entraron; contamos los fallidos.
					var bwe mongo.BulkWriteException
					if errors.As(err, &bwe) {
						failedRows.Add(int64(len(bwe.WriteErrors)))
					} else {
						// Error de transporte/conexion: lo guardamos y seguimos.
						e := fmt.Errorf("worker %d: InsertMany: %w", id, err)
						workerErr.CompareAndSwap(nil, &e)
						log.Printf("%v", e)
					}
				}
			}
		}(w)
	}

	// Reporte de progreso periodico.
	progressDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-progressDone:
				return
			case <-ticker.C:
				n := inserted.Load()
				elapsed := time.Since(start).Seconds()
				rate := float64(n) / elapsed
				log.Printf("progreso: %d insertados (%.0f docs/s)", n, rate)
			}
		}
	}()

	// Productor: lee el CSV y arma lotes de bson.D.
	readErr := func() error {
		batch := make([]any, 0, cfg.batchSize)
		var rowNum int64
		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				// Fila malformada: la saltamos y continuamos.
				skippedRows.Add(1)
				continue
			}
			rowNum++
			if cfg.limit > 0 && rowNum > int64(cfg.limit) {
				break
			}

			doc := buildDoc(cols, kinds, rec)
			batch = append(batch, doc)
			if len(batch) >= cfg.batchSize {
				select {
				case batches <- batch:
				case <-ctx.Done():
					return ctx.Err()
				}
				batch = make([]any, 0, cfg.batchSize)
			}
		}
		if len(batch) > 0 {
			select {
			case batches <- batch:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}()

	close(batches)
	wg.Wait()
	close(progressDone)

	elapsed := time.Since(start)
	log.Printf("migracion terminada en %s", elapsed.Round(time.Millisecond))
	log.Printf("  insertados:        %d", inserted.Load())
	if s := skippedRows.Load(); s > 0 {
		log.Printf("  filas saltadas:    %d (malformadas en CSV)", s)
	}
	if fr := failedRows.Load(); fr > 0 {
		log.Printf("  filas rechazadas:  %d (errores de escritura en Mongo)", fr)
	}
	if elapsed.Seconds() > 0 {
		log.Printf("  throughput:        %.0f docs/s", float64(inserted.Load())/elapsed.Seconds())
	}

	if readErr != nil && ctx.Err() != nil {
		return fmt.Errorf("migracion interrumpida: %w", readErr)
	}
	if readErr != nil {
		return fmt.Errorf("leyendo CSV: %w", readErr)
	}
	if ep := workerErr.Load(); ep != nil {
		return *ep
	}
	return nil
}

// buildDoc construye un documento BSON ordenado a partir de una fila del CSV,
// convirtiendo cada campo segun el tipo de su columna. Los valores numericos
// vacios o nan se almacenan como null.
func buildDoc(cols []string, kinds []columnKind, rec []string) bson.D {
	doc := make(bson.D, 0, len(cols))
	for i, col := range cols {
		var raw string
		if i < len(rec) {
			raw = rec[i]
		}
		doc = append(doc, bson.E{Key: col, Value: convert(kinds[i], raw)})
	}
	return doc
}

// convert transforma el texto crudo de una celda al tipo destino. Devuelve nil
// (null en Mongo) cuando el valor falta o no es parseable como numero.
func convert(kind columnKind, raw string) any {
	s := strings.TrimSpace(raw)
	switch kind {
	case kindString:
		return raw
	case kindBool:
		switch strings.ToLower(s) {
		case "true", "1":
			return true
		case "false", "0":
			return false
		default:
			return nil
		}
	case kindInt:
		if isMissing(s) {
			return nil
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
		// Algunos enteros vienen como "2000.0"; intentamos via float.
		if fv, err := strconv.ParseFloat(s, 64); err == nil && !math.IsNaN(fv) {
			return int64(fv)
		}
		return nil
	case kindFloat:
		if isMissing(s) {
			return nil
		}
		if fv, err := strconv.ParseFloat(s, 64); err == nil && !math.IsNaN(fv) && !math.IsInf(fv, 0) {
			return fv
		}
		return nil
	default:
		return raw
	}
}

// isMissing detecta marcadores de dato ausente en el dataset (CSV).
func isMissing(s string) bool {
	if s == "" {
		return true
	}
	switch strings.ToLower(s) {
	case "nan", "na", "null", "none":
		return true
	}
	return false
}
