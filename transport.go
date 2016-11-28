package clustertransport

import (
	"net"
	"net/url"
	"os"
	"syscall"
	"time"
)

// NewTransport is
func NewTransport(cfg *Config, uris ...string) *Transport {
	t := &Transport{
		cfg:            cfg,
		cluster:        cfg.Cluster,
		request:        make(chan *container),
		exit:           make(chan struct{}),
		lastRequestAt:  time.Now(),
		discover:       true,
		discoverTick:   1800,   // Discovers nodes per 1800 sec.
		discoverAfter:  300000, // Discovers nodes after passed 300,000 times.
		retryOnFailure: true,
		resurrectAfter: 5, // Kicking after disconnected that all of http connection
		maxRetries:     5, // Tries to retry's number for http request
	}

	t.sniffer = newSniffer(cfg, t.buildConns(uris))
	t.reloadConns()

	go t.run()
	return t
}

// Transport struct is
type Transport struct {
	cfg     *Config
	cluster ClusterBase
	conns   *Conns
	sniffer *Sniffer
	request chan *container
	exit    chan struct{}

	lastRequestAt  time.Time
	resurrectAfter int64
	retryOnFailure bool
	discover       bool
	discoverTick   int
	discoverAfter  int
	maxRetries     int
	counter        int
}

// Req is
func (t *Transport) Req(fun func(conn *Conn) (interface{}, error)) (interface{}, error) {
	c := containers.Get()
	defer containers.Put(c)

	c.fun = fun
	t.request <- c

	b := <-c.baggage
	defer baggages.Put(b)

	item, err := b.item, b.err
	return item, err
}

func (t *Transport) run() {
	tick := time.NewTicker(time.Duration(t.discoverTick) * time.Second)
	defer tick.Stop()
	// For debug
	// dtick := time.NewTicker(1 * time.Second)
	// defer dtick.Stop()
	// dtick2 := time.NewTicker(30 * time.Second)
	// defer dtick2.Stop()

	for {
		select {
		case c := <-t.request:
			b := baggages.Get(t.req(c, 0))
			c.baggage <- b
		case <-tick.C:
			if t.discover {
				t.reloadConns()
			}
		// For debug
		// case <-dtick.C:
		// pretty.Println("counter:", t.counter, "alives:", len(t.conns.alives()), " deads:", len(t.conns.deads()))
		// case <-dtick2.C:
		// pretty.Println(t.conns.all())
		case <-t.exit:
			break
		}
	}
}

func (t *Transport) req(c *container, tries int) (interface{}, error) {
	conn, err := t.conn()
	if err != nil {
		return nil, err
	}

	tries++

	item, err := c.fun(conn)

	if err != nil {
		switch err.(type) {
		default:
			if tries <= t.maxRetries {
				item, err = t.req(c, tries)
			}

			return item, err

		case *url.Error, *net.OpError, *os.SyscallError, *Econnrefused:
			syserr := err
			if uerr, ok := syserr.(*url.Error); ok {
				syserr = uerr.Err
			}
			if oerr, ok := syserr.(*net.OpError); ok {
				syserr = oerr.Err
			}
			if serr, ok := syserr.(*os.SyscallError); ok && serr.Err != syscall.ECONNREFUSED {
				if tries <= t.maxRetries {
					item, err = t.req(c, tries)
				}

				return item, err
			}

			if len(t.conns.alives()) > 1 {
				conn.terminate()
			}

			if t.retryOnFailure && tries <= t.maxRetries {
				item, err = t.req(c, tries)
			}

			return item, err
		}
	}

	if conn.failures > 0 {
		conn.healthy()
	}

	t.lastRequestAt = time.Now()
	return item, err
}

func (t *Transport) buildConns(uris []string) *Conns {
	var conns []*Conn
	for _, uri := range uris {
		if conn, err := t.cluster.Conn(uri, t); err == nil {
			conn.Uri = uri
			conns = append(conns, conn)
		}
	}

	// TODO: able to choose selector
	selector := &RoundRobinSelector{}

	return &Conns{cc: conns, selector: selector}
}

func (t *Transport) conn() (*Conn, error) {
	if time.Now().Unix() > t.lastRequestAt.Unix()+t.resurrectAfter {
		t.resurrectDeads()
	}

	t.counter++

	if t.discover && t.counter%t.discoverAfter == 0 {
		t.reloadConns()
	}

	return t.conns.conn()
}

func (t *Transport) reloadConns() {
	uris, _ := t.sniffer.Sniffed()
	t.rebuildConns(uris)
}

func (t *Transport) resurrectDeads() {
	for _, dead := range t.conns.deads() {
		dead.resurrect()
	}
}

func (t *Transport) rebuildConns(uris []string) {

	// TODO
	// t.cluster.CloseConns()

	if conns := t.buildConns(uris); len(conns.alives()) > 0 {
		t.conns = conns
		t.sniffer = newSniffer(t.cfg, t.conns)
	}
}
