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
	DiscoverTick   int   // Default: Discovers nodes per 600 sec
	DiscoverAfter  int64 // Default: Discovers nodes after passed 100,000 times
	RetryOnFailure bool  // Default: Retrying asap when one of connection failed
	ResurrectAfter int64 // Default: Kicking recovers after a second when disconnected all of connections
	MaxRetries     int   // Default: Tries to retry's number for http request
	Debug          bool
}

// PrintNothing is
func PrintNothing(format string, v ...interface{}) {}

// NewConfig is
func NewConfig() *Config {
	return &Config{
		Selector:       &RoundRobinSelector{},
		Logger:         PrintNothing,
		Discover:       true,
		DiscoverTick:   600,
		DiscoverAfter:  100000,
		RetryOnFailure: false,
		ResurrectAfter: 1,
		MaxRetries:     5,
	}
}
