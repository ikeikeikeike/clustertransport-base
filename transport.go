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
		cfg:           cfg,
		request:       make(chan *container),
		configure:     make(chan struct{ fun func(*Config) *Config }),
		exit:          make(chan struct{}),
		lastRequestAt: time.Now(),
	}

	t.sniffer = newSniffer(cfg, t.buildConns(uris))
	t.reloadConns()

	go t.run()
	return t
}

// Transport struct is
type Transport struct {
	cfg           *Config
	conns         *Conns
	sniffer       *Sniffer
	request       chan *container
	configure     chan struct{ fun func(*Config) *Config }
	exit          chan struct{}
	counter       int64
	lastRequestAt time.Time
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

// Configure is
func (t *Transport) Configure(fun func(cfg *Config) *Config) {
	t.configure <- struct{ fun func(*Config) *Config }{fun: fun}
}

func (t *Transport) run() {
	tick := time.NewTicker(time.Duration(t.cfg.DiscoverTick) * time.Second)
	defer tick.Stop()

	// For debug
	// debugTick := time.NewTicker(1 * time.Second)
	// defer debugTick.Stop()
	// debugTraceTick := time.NewTicker(30 * time.Second)
	// defer debugTraceTick.Stop()

	for {
		select {
		case c := <-t.request:
			b := baggages.Get(t.req(c, 0))
			c.baggage <- b
		case c := <-t.configure:
			t.cfg = c.fun(t.cfg)
		case <-tick.C:
			if t.cfg.Discover {
				t.cfg.Logger("Discover clusters by `discoverTick`: "+
					"next time after %d secs", t.cfg.DiscoverTick)
				t.reloadConns()
			}
		// For debug
		// case <-debugTick.C:
		// pretty.Println(
		// "counter:", t.counter,
		// "alives:", len(t.conns.alives()),
		// "deads:", len(t.conns.deads()),
		// )
		// case <-debugTraceTick.C:
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
			if tries <= t.cfg.MaxRetries {
				t.cfg.Logger("Request retries %d/%d", tries, t.cfg.MaxRetries)
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
				if tries <= t.cfg.MaxRetries {
					t.cfg.Logger("Request retries %d/%d", tries, t.cfg.MaxRetries)
					item, err = t.req(c, tries)
				}

				return item, err
			}

			if len(t.conns.alives()) > 1 {
				t.cfg.Logger("Close connection to cluster via %s", conn.Uri)
				conn.terminate()
			}

			if t.cfg.RetryOnFailure && tries <= t.cfg.MaxRetries {
				t.cfg.Logger("Do retryOnFailure %d/%d", tries, t.cfg.MaxRetries)
				item, err = t.req(c, tries)
			}

			return item, err
		}
	}

	if conn.Failures > 0 {
		conn.healthy()
	}

	t.lastRequestAt = time.Now()
	return item, err
}

func (t *Transport) buildConns(uris []string) *Conns {
	var conns []*Conn

	for _, uri := range uris {
		conn, err := t.cfg.Cluster.Conn(uri, t)

		if err != nil {
			t.cfg.Logger("Failed to connection establishment via %s: %s",
				uri, err.Error())
			continue
		}

		conn.Uri = uri
		conns = append(conns, conn)
	}

	return &Conns{cc: conns, selector: t.cfg.Selector}
}

func (t *Transport) conn() (*Conn, error) {
	if time.Now().Unix() > t.lastRequestAt.Unix()+t.cfg.ResurrectAfter {
		t.cfg.Logger("Resurrect connection")
		t.resurrectDeads()
	}

	t.counter++

	if t.cfg.Discover && t.counter%t.cfg.DiscoverAfter == 0 {
		t.cfg.Logger("Discover clusters by `discoverAfter`: "+
			"next time after %d requests", t.cfg.DiscoverAfter)
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
		t.counter = 0
		t.conns = conns
		t.sniffer = newSniffer(t.cfg, t.conns)
	}
}
