// Package protocol defines the TCP message types exchanged between the API
// coordinator and the ML worker nodes.
package protocol

// TrainTask is sent by the coordinator to a worker node.
// X is a row-major float64 slice of shape Rows × Cols (design matrix with intercept).
// Y is a float64 slice of length Rows (target values).
type TrainTask struct {
	X    []float64
	Y    []float64
	Rows int
	Cols int
}

// TrainResult is returned by the worker node to the coordinator.
// XTX is a row-major float64 slice of shape Cols × Cols.
// XTy is a float64 slice of length Cols.
// Err is non-empty when the node failed to compute the partial result.
type TrainResult struct {
	XTX []float64
	XTy []float64
	Err string
}
