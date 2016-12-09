package clustertransport

// ClusterBase has interfaces which connects Cluster System.
type ClusterBase interface {
	Sniff(conn *Conn) []string
	Conn(uri string) (*Conn, error)
}

// SelectorBase has a interface which selects cluster connections.
type SelectorBase interface {
	Select(conns []*Conn) *Conn
}

// Econnrefused notices dead connection to Cluseter Transport.
type Econnrefused struct {
	s string
}

// Error returns Econnrefused's error message.
func (e *Econnrefused) Error() string {
	return e.s
}

// Config is
type Config struct {
	Cluster  ClusterBase
	Selector SelectorBase

	Logger func(format string, params ...interface{})

	Discover       bool  // Default: true,
	DiscoverTick   int   // Default: Discovers nodes per 120 sec
	DiscoverAfter  int64 // Default: Discovers nodes after passed 10,000 requests
	RetryOnFailure bool  // Default: Retrying asap when one of connection failed
	ResurrectAfter int64 // Default: Tries to resurrect some of connections when Cluster Transport hasn't request to cluster system until it passed 30 sec.
	MaxRetries     int   // Default: Tries to retry's number for http request
	Debug          bool
}

// PrintNothing does nothing.
func PrintNothing(format string, v ...interface{}) {}

// NewConfig returns a Config struct which has some of field for handling Cluster Transport.
func NewConfig() *Config {
	return &Config{
		Selector:       &RoundRobinSelector{},
		Logger:         PrintNothing,
		Discover:       true,
		DiscoverTick:   120,
		DiscoverAfter:  100000,
		RetryOnFailure: false,
		ResurrectAfter: 30,
		MaxRetries:     5,
	}
}
