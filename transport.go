package clustertransport

import (
	"net"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/k0kubun/pp"
)

// NewTransport is
func NewTransport(cfg *Config, uris ...string) *Transport {
	t := &Transport{
		cfg:             cfg,
		cluster:         cfg.Cluster,
		uris:            uris,
		request:         make(chan *container),
		exit:            make(chan struct{}),
		lastRequestAt:   time.Now(),
		reload:          true,
		reloadAfter:     20000,
		retryOnFailure:  true,
		resurrectAfter:  60,
		maxRetries:      3,
	}

	t.conns = t.buildConns()
	t.sniffer = newSniffer(cfg, t.conns)

	go t.run()
	return t
}

// Transport struct is
type Transport struct {
	cfg     *Config
	cluster ClusterBase
	sniffer *Sniffer
	uris    []string
	conns   *Conns
	request chan *container
	exit    chan struct{}

	lastRequestAt   time.Time
	resurrectAfter  int64
	retryOnFailure  bool
	reloadAfter     int
	reload          bool
	maxRetries      int
	counter         int
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
	// For debug
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	tick2 := time.NewTicker(30 * time.Second)
	defer tick2.Stop()

	for {
		select {
		case c := <-t.request:
			b := baggages.Get(t.req(c, 0))
			c.baggage <- b
		// For debug
		case <-tick.C:
			pp.Println("alives:", len(t.conns.alives()), " deads:", len(t.conns.deads()))
		case <-tick2.C:
			pp.Println(t.conns.all())
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

			conn.terminate()

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

func (t *Transport) buildConns() *Conns {
	var conns []*Conn
	for _, uri := range t.uris {
		conn := t.cluster.Conn(uri, t)
		conn.Uri = uri

		conns = append(conns, conn)
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

	if t.reload && t.counter%t.reloadAfter == 0 {
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
	t.uris = uris

	// TODO
	// t.cluster.CloseConns()

	t.conns = t.buildConns()
	t.sniffer = newSniffer(t.cfg, t.conns)
}
