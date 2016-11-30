package clustertransport

// ClusterBase interface is
type ClusterBase interface {
	Sniff(conn *Conn) []string
	Conn(uri string, st *Transport) (*Conn, error)
}

// SelectorBase is
type SelectorBase interface {
	Select(conns []*Conn) *Conn
}

// Econnrefused is
type Econnrefused struct {
	s string
}

// Error is
func (e *Econnrefused) Error() string {
	return e.s
}

// Config is
type Config struct {
	Cluster  ClusterBase
	Selector SelectorBase

	Logger func(format string, params ...interface{})

	Discover       bool  // Default: true,
	DiscoverTick   int   // Default: Discovers nodes per 1800 sec
	DiscoverAfter  int64 // Default: Discovers nodes after passed 300,000 times
	RetryOnFailure bool  // Default: Retrying asap when one of connection failed
	ResurrectAfter int64 // Default: Kicking recovers after disconnected that all of http connection
	MaxRetries     int   // Default: Tries to retry's number for http request
}

// PrintNothing is
func PrintNothing(format string, v ...interface{}) {}

// NewConfig is
func NewConfig() *Config {
	return &Config{
		Selector:       &RoundRobinSelector{},
		Logger:         PrintNothing,
		Discover:       true,
		DiscoverTick:   1800,
		DiscoverAfter:  300000,
		RetryOnFailure: false,
		ResurrectAfter: 5,
		MaxRetries:     5,
	}
}
