package analysis

import "errors"

// ErrUnfeasibleTx transaction not feasible for analysis error
var ErrUnfeasibleTx = errors.New("Transaction not feasible for analysis")

// ErrUnfeasibleAnalysis analysis not feasible error
var ErrUnfeasibleAnalysis = errors.New("Analysis not feasible")
