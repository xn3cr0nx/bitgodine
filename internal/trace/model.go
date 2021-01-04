package trace

// Cluster struct to classify the address in a certain cluster
type Cluster struct {
	Type     string `json:"type,omitempty"`
	Message  string `json:"message,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Verified bool   `json:"verified,omitempty"`
}

// Trace between ouput and spending tx for tracing
type Trace struct {
	TxID string `json:"txid"`
	Next []Next `json:"next"`
}

// Next spending tx info
type Next struct {
	TxID     string    `json:"txid"`
	Receiver string    `json:"receiver"`
	Vout     uint32    `json:"vout"`
	Amount   float64   `json:"amount"`
	Weight   float64   `json:"weight"`
	Analysis string    `json:"analysis"`
	Clusters []Cluster `json:"clusters"`
}

// Flow list of maps creating monetary flow
type Flow struct {
	Traces     []map[string]Trace `json:"traces"`
	Occurences []string           `json:"occurences"`
}
