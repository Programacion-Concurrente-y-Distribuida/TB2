// Package coordinator implements the distributed training orchestrator.
// It divides the design matrix into partitions, ships each to a worker node
// via TCP (gob-encoded), collects the partial XTX/XTy matrices, aggregates
// them, and then delegates solving to aqsml.SolveNormalEquations.
package coordinator

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"aqsml/internal/protocol"
)

// NodeResult holds the outcome of a single worker node's computation.
type NodeResult struct {
	Addr     string
	Duration time.Duration
	Err      error
}

// ProgressFunc is called by DistributedFit when a node starts or finishes.
// done=false → node received its task; done=true → node returned its result.
type ProgressFunc func(addr string, done bool)

// DistributedFit sends row partitions of (trainX, trainY) to worker nodes in
// parallel, waits for their partial XTX/XTy results, and returns the
// aggregated matrices ready for SolveNormalEquations.
//
// trainX is row-major with shape rows×cols (includes intercept column).
// trainY has length rows.
func DistributedFit(
	nodeAddrs []string,
	trainX []float64,
	trainY []float64,
	rows, cols int,
	progress ProgressFunc,
) (xtxTotal []float64, xtyTotal []float64, nodeResults []NodeResult, err error) {
	if len(nodeAddrs) == 0 {
		return nil, nil, nil, fmt.Errorf("ningún nodo ML configurado")
	}

	n := len(nodeAddrs)
	chunk := (rows + n - 1) / n

	nodeResults = make([]NodeResult, n)
	partialXTX := make([][]float64, n)
	partialXTy := make([][]float64, n)

	var wg sync.WaitGroup

	for i, addr := range nodeAddrs {
		startRow := i * chunk
		endRow := startRow + chunk
		if endRow > rows {
			endRow = rows
		}
		if startRow >= rows {
			nodeResults[i] = NodeResult{Addr: addr}
			continue
		}

		task := protocol.TrainTask{
			X:    trainX[startRow*cols : endRow*cols],
			Y:    trainY[startRow:endRow],
			Rows: endRow - startRow,
			Cols: cols,
		}

		wg.Add(1)
		go func(idx int, nodeAddr string, t protocol.TrainTask) {
			defer wg.Done()
			start := time.Now()

			if progress != nil {
				progress(nodeAddr, false)
			}

			result, sendErr := sendTask(nodeAddr, t)
			dur := time.Since(start)

			nodeResults[idx] = NodeResult{Addr: nodeAddr, Duration: dur, Err: sendErr}

			if sendErr == nil && result.Err != "" {
				nodeResults[idx].Err = fmt.Errorf("node error: %s", result.Err)
			}

			if nodeResults[idx].Err == nil {
				partialXTX[idx] = result.XTX
				partialXTy[idx] = result.XTy
			}

			if progress != nil {
				progress(nodeAddr, true)
			}

			log.Printf("coordinator: node %s finished in %s (err=%v)", nodeAddr, dur.Round(time.Millisecond), nodeResults[idx].Err)
		}(i, addr, task)
	}

	wg.Wait()

	// Check for errors
	for _, nr := range nodeResults {
		if nr.Err != nil {
			return nil, nil, nodeResults, fmt.Errorf("nodo %s falló: %w", nr.Addr, nr.Err)
		}
	}

	// Aggregate partial matrices
	xtxTotal = make([]float64, cols*cols)
	xtyTotal = make([]float64, cols)
	for i := range nodeAddrs {
		if partialXTX[i] == nil {
			continue
		}
		for j, v := range partialXTX[i] {
			xtxTotal[j] += v
		}
		for j, v := range partialXTy[i] {
			xtyTotal[j] += v
		}
	}

	return xtxTotal, xtyTotal, nodeResults, nil
}

// CheckNodes pings each node with a trivial 1×1 task and returns latency info.
func CheckNodes(nodeAddrs []string) []NodeResult {
	results := make([]NodeResult, len(nodeAddrs))
	var wg sync.WaitGroup
	for i, addr := range nodeAddrs {
		wg.Add(1)
		go func(idx int, a string) {
			defer wg.Done()
			start := time.Now()
			conn, err := net.DialTimeout("tcp", a, 3*time.Second)
			if err != nil {
				results[idx] = NodeResult{Addr: a, Duration: time.Since(start), Err: err}
				return
			}
			conn.Close()
			results[idx] = NodeResult{Addr: a, Duration: time.Since(start)}
		}(i, addr)
	}
	wg.Wait()
	return results
}

// sendTask dials a worker node, sends a TrainTask, and decodes the TrainResult.
func sendTask(addr string, task protocol.TrainTask) (protocol.TrainResult, error) {
	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return protocol.TrainResult{}, fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	deadline := time.Now().Add(20 * time.Minute)
	if err := conn.SetDeadline(deadline); err != nil {
		return protocol.TrainResult{}, err
	}

	if err := gob.NewEncoder(conn).Encode(task); err != nil {
		return protocol.TrainResult{}, fmt.Errorf("encode task: %w", err)
	}

	var result protocol.TrainResult
	if err := gob.NewDecoder(conn).Decode(&result); err != nil {
		return protocol.TrainResult{}, fmt.Errorf("decode result: %w", err)
	}

	return result, nil
}
