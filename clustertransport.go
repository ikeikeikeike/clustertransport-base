package clustertransport

// ClusterBase interface is
type ClusterBase interface {
	Sniff() []string
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
