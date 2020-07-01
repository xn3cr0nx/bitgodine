package cltz

// Clusterizer generates clusters generating them fetched blocks
type Clusterizer interface {
	Clusterize()
	Done() (uint32, error)
}
