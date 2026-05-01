// benchmark/main.go
// Benchmark automático: Secuencial vs Concurrente — Algoritmos de Recomendación Go
//
// Experimento 1: 100 ejecuciones × 2 paradigmas → Boxplot + Violin (tiempo, CPU, RAM)
// Experimento 2: Workers variables → Line plot (tiempo y speedup)
//
// Uso:
//   go run . [--runs N] [--workers-runs M] [--skip-compile]
//
// Salida:
//   results/exp1_raw.csv
//   results/exp2_workers.csv
//   results/summary.txt
//   results/charts.html   ← gráficos interactivos en el navegador

package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ─────────────────────────── Configuración ────────────────────────────────────

const (
	trimProportion = 0.10 // recortar 10% en cada extremo
)

var (
	nRuns        = flag.Int("runs", 100, "ejecuciones por paradigma (Exp 1)")
	workersRuns  = flag.Int("workers-runs", 10, "ejecuciones por config de workers (Exp 2)")
	skipCompile  = flag.Bool("skip-compile", false, "omite la compilación de binarios")
	datasetIn    = flag.String("dataset", "", "ruta al dataset (default: ../dataset_1m.csv)")
)

// ─────────────────────────── Rutas ────────────────────────────────────────────

var (
	scriptDir   string
	projectRoot string
	recoDir     string
	binDir      string
	resultsDir  string
	seqBin      string
	concBin     string
	datasetPath string
)

func initPaths() {
	// El ejecutable siempre corre desde benchmark/
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	// Durante `go run .` el ejecutable está en /tmp; usamos la cwd
	cwd, _ := os.Getwd()
	_ = ex

	scriptDir = cwd
	projectRoot = filepath.Dir(scriptDir)
	recoDir = filepath.Join(projectRoot, "Paradigmas", "recomendation")
	binDir = filepath.Join(scriptDir, "bin")
	resultsDir = filepath.Join(scriptDir, "results")
	seqBin = filepath.Join(binDir, "secuencial")
	concBin = filepath.Join(binDir, "concurrente")

	if *datasetIn != "" {
		datasetPath = *datasetIn
	} else {
		datasetPath = filepath.Join(projectRoot, "dataset_1m.csv")
	}
}

// ─────────────────────────── Estadísticas ─────────────────────────────────────

func trimmedMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)
	k := int(math.Round(float64(len(sorted)) * trimProportion))
	if k*2 >= len(sorted) {
		k = 0
	}
	slice := sorted[k : len(sorted)-k]
	sum := 0.0
	for _, v := range slice {
		sum += v
	}
	return sum / float64(len(slice))
}

func stddev(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	mean := trimmedMean(data)
	sum := 0.0
	for _, v := range data {
		d := v - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(data)))
}

func minMax(data []float64) (float64, float64) {
	if len(data) == 0 {
		return 0, 0
	}
	mn, mx := data[0], data[0]
	for _, v := range data[1:] {
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
	}
	return mn, mx
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := p / 100.0 * float64(len(sorted)-1)
	lo := int(idx)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

// ─────────────────────────── Medición de proceso ──────────────────────────────

// RunMetrics almacena las métricas de una sola ejecución.
type RunMetrics struct {
	TimeS      float64
	CPUPct     float64 // promedio de muestras
	MemPeakMB  float64 // pico de RSS
}

// measureProc lanza el binario y muestrea CPU+Mem cada 50ms vía /proc o ps.
// En macOS usamos `ps` para obtener RSS; para CPU usamos tiempo de usuario del proceso.
func measureProc(binary string, args []string) (RunMetrics, error) {
	cmd := exec.Command(binary, args...)
	cmd.Dir = recoDir // los binarios usan rutas relativas a su directorio fuente (../../data/)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr

	t0 := time.Now()
	if err := cmd.Start(); err != nil {
		return RunMetrics{}, fmt.Errorf("iniciar %s: %w", binary, err)
	}

	pid := cmd.Process.Pid

	var (
		mu       sync.Mutex
		memSamples []float64
		done     = make(chan struct{})
	)

	// Goroutine de muestreo
	go func() {
		defer close(done)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rss := sampleRSSMB(pid)
				if rss > 0 {
					mu.Lock()
					memSamples = append(memSamples, rss)
					mu.Unlock()
				}
			}
			// Verificar si el proceso terminó
			if cmd.ProcessState != nil {
				return
			}
		}
	}()

	err := cmd.Wait()
	elapsed := time.Since(t0).Seconds()
	<-done // asegurar que el goroutine terminó

	if err != nil {
		return RunMetrics{}, fmt.Errorf("ejecutar %s: %w", binary, err)
	}

	// CPU: usar UserTime del proceso (más preciso que muestreo en macOS)
	cpuPct := 0.0
	if cmd.ProcessState != nil && elapsed > 0 {
		userSec := cmd.ProcessState.UserTime().Seconds()
		sysSec := cmd.ProcessState.SystemTime().Seconds()
		cpuPct = (userSec+sysSec) / elapsed * 100.0
	}

	// Memoria peak
	memPeak := 0.0
	mu.Lock()
	if len(memSamples) > 0 {
		for _, v := range memSamples {
			if v > memPeak {
				memPeak = v
			}
		}
	}
	mu.Unlock()

	// Fallback: si no obtuvimos muestras (proceso muy rápido), usar /usr/bin/time
	if memPeak == 0 {
		memPeak = peakRSSFallback(pid)
	}

	return RunMetrics{
		TimeS:     elapsed,
		CPUPct:    cpuPct,
		MemPeakMB: memPeak,
	}, nil
}

// sampleRSSMB obtiene el RSS actual del PID en MB (macOS: usa ps).
func sampleRSSMB(pid int) float64 {
	out, err := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return 0
	}
	kb, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0
	}
	return kb / 1024.0
}

// peakRSSFallback retorna 0 si no se pudo medir (sólo para el log).
func peakRSSFallback(_ int) float64 {
	return 0
}

// ─────────────────────────── Compilación ──────────────────────────────────────

func compileBinaries() {
	banner("COMPILACIÓN DE BINARIOS GO")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		log.Fatal(err)
	}

	targets := []struct {
		name    string
		sources []string
	}{
		{"secuencial", []string{"secuencial.go", "helpers.go"}},
		{"concurrente", []string{"concurrente.go", "helpers.go"}},
	}

	for _, t := range targets {
		out := filepath.Join(binDir, t.name)
		args := append([]string{"build", "-o", out}, t.sources...)
		fmt.Printf("  $ go build -o bin/%s %s\n", t.name, strings.Join(t.sources, " "))
		cmd := exec.Command("go", args...)
		cmd.Dir = recoDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("compilación de '%s' falló: %v", t.name, err)
		}
		fmt.Printf("  ✓ %s\n", t.name)
	}
}

// ─────────────────────────── Experimento 1 ────────────────────────────────────

type Exp1Row struct {
	Paradigm  string
	Run       int
	TimeS     float64
	CPUPct    float64
	MemPeakMB float64
}

func runExperiment1() []Exp1Row {
	banner(fmt.Sprintf("EXPERIMENTO 1 — Comparativa (%d ejecuciones × 2 paradigmas)", *nRuns))

	paradigms := []struct {
		name   string
		binary string
		args   []string
	}{
		{"Secuencial", seqBin, []string{datasetPath, "/dev/null"}},
		{"Concurrente", concBin, []string{datasetPath, "/dev/null"}},
	}

	var rows []Exp1Row

	for _, p := range paradigms {
		fmt.Printf("\n  [%s]\n", p.name)
		for i := 1; i <= *nRuns; i++ {
			m, err := measureProc(p.binary, p.args)
			if err != nil {
				log.Printf("    [WARN] run %d falló: %v — omitiendo", i, err)
				continue
			}
			rows = append(rows, Exp1Row{
				Paradigm:  p.name,
				Run:       i,
				TimeS:     m.TimeS,
				CPUPct:    m.CPUPct,
				MemPeakMB: m.MemPeakMB,
			})
			if i%10 == 0 || i == 1 {
				fmt.Printf("    [%3d/%d]  t=%.2fs  cpu=%.1f%%  mem=%.0fMB\n",
					i, *nRuns, m.TimeS, m.CPUPct, m.MemPeakMB)
			}
		}
	}

	// Guardar CSV
	saveExp1CSV(rows)
	return rows
}

func saveExp1CSV(rows []Exp1Row) {
	path := filepath.Join(resultsDir, "exp1_raw.csv")
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"paradigm", "run", "time_s", "cpu_pct", "mem_peak_mb"})
	for _, r := range rows {
		_ = w.Write([]string{
			r.Paradigm,
			strconv.Itoa(r.Run),
			fmt.Sprintf("%.4f", r.TimeS),
			fmt.Sprintf("%.2f", r.CPUPct),
			fmt.Sprintf("%.2f", r.MemPeakMB),
		})
	}
	w.Flush()
	fmt.Printf("\n  ✓ CSV guardado: %s\n", path)
}

// ─────────────────────────── Experimento 2 ────────────────────────────────────

type Exp2Row struct {
	Workers     int
	TimeTrimS   float64
	TimeMeanS   float64
	TimeStdS    float64
	TimeMinS    float64
	TimeMaxS    float64
}

func runExperiment2() []Exp2Row {
	numCPU := runtime.NumCPU()
	workerConfigs := uniqueSorted([]int{1, 2, 4, 8, 16, 32, numCPU})

	banner(fmt.Sprintf("EXPERIMENTO 2 — Escalabilidad (%d runs/config)", *workersRuns))
	fmt.Printf("  Workers a probar: %v  |  CPUs lógicos: %d\n", workerConfigs, numCPU)

	var rows []Exp2Row

	for _, nw := range workerConfigs {
		var times []float64
		for i := 0; i < *workersRuns; i++ {
			m, err := measureProc(concBin, []string{datasetPath, "/dev/null", strconv.Itoa(nw)})
			if err != nil {
				log.Printf("  [WARN] workers=%d run %d falló: %v", nw, i+1, err)
				continue
			}
			times = append(times, m.TimeS)
		}
		if len(times) == 0 {
			continue
		}
		mn, mx := minMax(times)
		row := Exp2Row{
			Workers:   nw,
			TimeTrimS: trimmedMean(times),
			TimeMeanS: func() float64 {
				s := 0.0
				for _, v := range times {
					s += v
				}
				return s / float64(len(times))
			}(),
			TimeStdS:  stddev(times),
			TimeMinS:  mn,
			TimeMaxS:  mx,
		}
		rows = append(rows, row)
		fmt.Printf("  workers=%2d  →  t_trim=%.3fs  σ=%.3fs  [%.3f–%.3f]\n",
			nw, row.TimeTrimS, row.TimeStdS, mn, mx)
	}

	saveExp2CSV(rows)
	return rows
}

func saveExp2CSV(rows []Exp2Row) {
	path := filepath.Join(resultsDir, "exp2_workers.csv")
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"workers", "time_trim_s", "time_mean_s", "time_std_s", "time_min_s", "time_max_s"})
	for _, r := range rows {
		_ = w.Write([]string{
			strconv.Itoa(r.Workers),
			fmt.Sprintf("%.4f", r.TimeTrimS),
			fmt.Sprintf("%.4f", r.TimeMeanS),
			fmt.Sprintf("%.4f", r.TimeStdS),
			fmt.Sprintf("%.4f", r.TimeMinS),
			fmt.Sprintf("%.4f", r.TimeMaxS),
		})
	}
	w.Flush()
	fmt.Printf("  ✓ CSV guardado: %s\n", path)
}

// ─────────────────────────── Resumen estadístico ──────────────────────────────

func printSummary(rows1 []Exp1Row, rows2 []Exp2Row) float64 {
	var seqTimes, concTimes []float64
	var seqCPU, concCPU []float64
	var seqMem, concMem []float64

	for _, r := range rows1 {
		switch r.Paradigm {
		case "Secuencial":
			seqTimes = append(seqTimes, r.TimeS)
			seqCPU = append(seqCPU, r.CPUPct)
			seqMem = append(seqMem, r.MemPeakMB)
		case "Concurrente":
			concTimes = append(concTimes, r.TimeS)
			concCPU = append(concCPU, r.CPUPct)
			concMem = append(concMem, r.MemPeakMB)
		}
	}

	tSeq := trimmedMean(seqTimes)
	tConc := trimmedMean(concTimes)
	speedup := 0.0
	if tConc > 0 {
		speedup = tSeq / tConc
	}

	// Mejor config de workers
	var bestRow Exp2Row
	if len(rows2) > 0 {
		bestRow = rows2[0]
		for _, r := range rows2 {
			if r.TimeTrimS < bestRow.TimeTrimS {
				bestRow = r
			}
		}
	}
	bestSpeedup := 0.0
	if bestRow.TimeTrimS > 0 {
		bestSpeedup = tSeq / bestRow.TimeTrimS
	}

	sep := strings.Repeat("═", 62)
	lines := []string{
		"",
		sep,
		"  RESUMEN EJECUTIVO",
		sep,
		"",
		fmt.Sprintf("  Experimento 1 — Media recortada (%.0f%% en cada extremo)", trimProportion*100),
		fmt.Sprintf("  %-25s  %12s  %12s", "Métrica", "Secuencial", "Concurrente"),
		"  " + strings.Repeat("─", 52),
		fmt.Sprintf("  %-25s  %12.3f  %12.3f", "Tiempo (s)", tSeq, tConc),
		fmt.Sprintf("  %-25s  %12s  %11.2f×", "Speedup", "—", speedup),
		fmt.Sprintf("  %-25s  %12.1f  %12.1f", "CPU promedio (%)", trimmedMean(seqCPU), trimmedMean(concCPU)),
		fmt.Sprintf("  %-25s  %12.1f  %12.1f", "RAM pico (MB)", trimmedMean(seqMem), trimmedMean(concMem)),
		"",
		"  Experimento 2 — Workers óptimo",
		func() string {
			if len(rows2) == 0 {
				return "    (sin datos — todos los runs del Exp 2 fallaron)"
			}
			return fmt.Sprintf("    Workers: %d  |  Tiempo: %.3fs  |  Speedup: %.2f×",
				bestRow.Workers, bestRow.TimeTrimS, bestSpeedup)
		}(),
		"",
		sep,
	}

	text := strings.Join(lines, "\n")
	fmt.Println(text)

	summaryPath := filepath.Join(resultsDir, "summary.txt")
	_ = os.WriteFile(summaryPath, []byte(text), 0o644)
	fmt.Printf("\n  ✓ Resumen guardado: %s\n", summaryPath)

	return tSeq
}

// ─────────────────────────── Helpers ──────────────────────────────────────────

func banner(msg string) {
	line := strings.Repeat("─", 62)
	fmt.Printf("\n%s\n  %s\n%s\n", line, msg, line)
}

func uniqueSorted(in []int) []int {
	seen := make(map[int]bool)
	var out []int
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Ints(out)
	return out
}

// ─────────────────────────── Main ─────────────────────────────────────────────

func main() {
	flag.Parse()
	initPaths()

	// Verificaciones previas
	if _, err := os.Stat(datasetPath); os.IsNotExist(err) {
		fmt.Printf("\n[ERROR] Dataset no encontrado: %s\n", datasetPath)
		fmt.Println("\n  Genera el dataset primero ejecutando en orden:")
		fmt.Printf("    cd %q && go run descargar.go\n", filepath.Join(projectRoot, "0. Descargar"))
		fmt.Printf("    cd %q && go run union_y_limpieza.go\n", filepath.Join(projectRoot, "1. Limpieza"))
		fmt.Printf("    cd %q && go run expandir_dataset.go\n", filepath.Join(projectRoot, "2. Expandir"))
		os.Exit(1)
	}
	if _, err := os.Stat(recoDir); os.IsNotExist(err) {
		log.Fatalf("[ERROR] Directorio de algoritmos no encontrado: %s", recoDir)
	}
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n  Dataset   : %s\n", datasetPath)
	fmt.Printf("  Resultados: %s\n", resultsDir)
	fmt.Printf("  CPUs host : %d\n", runtime.NumCPU())

	// Pipeline
	if !*skipCompile {
		compileBinaries()
	} else {
		banner("COMPILACIÓN OMITIDA (--skip-compile)")
	}

	t0 := time.Now()

	rows1 := runExperiment1()
	rows2 := runExperiment2()
	tSeqTrim := printSummary(rows1, rows2)
	generateCharts(rows1, rows2, tSeqTrim)

	fmt.Printf("\n  ✅ Benchmark completado en %.1f min\n", time.Since(t0).Minutes())
	fmt.Printf("  📊 Abre: %s/charts.html\n\n", resultsDir)
}
