// ml-node: worker TCP server for distributed linear regression.
//
// Each node listens for TrainTask messages (gob-encoded), computes the partial
// XTX and XTy matrices for its data partition, and returns a TrainResult.
// The coordinator (API) aggregates all partials and solves for β.
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"

	"aqsml/internal/protocol"
)

func main() {
	port := os.Getenv("NODE_PORT")
	if port == "" {
		port = "9000"
	}

	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("ml-node: listen :%s: %v", port, err)
	}
	log.Printf("ml-node ready on :%s", port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("ml-node: accept: %v", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()

	dec := gob.NewDecoder(conn)
	enc := gob.NewEncoder(conn)

	var task protocol.TrainTask
	if err := dec.Decode(&task); err != nil {
		log.Printf("ml-node: decode from %s: %v", addr, err)
		enc.Encode(protocol.TrainResult{Err: err.Error()}) //nolint:errcheck
		return
	}

	log.Printf("ml-node: received task from %s — rows=%d cols=%d", addr, task.Rows, task.Cols)
	result := computePartial(task)

	if err := enc.Encode(result); err != nil {
		log.Printf("ml-node: encode to %s: %v", addr, err)
	}
	if result.Err != "" {
		log.Printf("ml-node: task error: %s", result.Err)
	} else {
		log.Printf("ml-node: task done for %s", addr)
	}
}

// computePartial accumulates the partial XTX (lower triangle) and XTy for
// the data partition. The result is the same computation the coordinator's
// goroutines did before distribution — now running in a separate process.
func computePartial(task protocol.TrainTask) protocol.TrainResult {
	n := task.Rows
	c := task.Cols

	if n == 0 || c == 0 {
		return protocol.TrainResult{Err: fmt.Sprintf("partición vacía: rows=%d cols=%d", n, c)}
	}
	if len(task.X) != n*c {
		return protocol.TrainResult{Err: fmt.Sprintf("X length mismatch: got %d, want %d", len(task.X), n*c)}
	}
	if len(task.Y) != n {
		return protocol.TrainResult{Err: fmt.Sprintf("Y length mismatch: got %d, want %d", len(task.Y), n)}
	}

	localXTX := make([]float64, c*c)
	localXTy := make([]float64, c)

	for i := 0; i < n; i++ {
		yi := task.Y[i]
		rowBase := i * c
		for j := 0; j < c; j++ {
			xj := task.X[rowBase+j]
			localXTy[j] += xj * yi
			base := j * c
			for k := 0; k <= j; k++ {
				localXTX[base+k] += xj * task.X[rowBase+k]
			}
		}
	}

	// Symmetrize
	for j := 0; j < c; j++ {
		for k := 0; k < j; k++ {
			localXTX[k*c+j] = localXTX[j*c+k]
		}
	}

	return protocol.TrainResult{XTX: localXTX, XTy: localXTy}
}
