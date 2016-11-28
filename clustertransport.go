package clustertransport

// ClusterBase interface is
type ClusterBase interface {
	Sniff(conn *Conn) []string
	Conn(uri string, st *Transport) *Conn
}

// SelectorBase is
type SelectorBase interface {
	Select() *Conn
}

// Config should be Context
type Config struct {
	Cluster ClusterBase
}

// Econnrefused is
type Econnrefused struct {
	s string
}

// Error is
func (e *Econnrefused) Error() string {
	return e.s
}
