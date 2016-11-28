package clustertransport

import (
	"reflect"
	"time"
)

// NewTransport is
func NewTransport(cfg *Config, uris ...string) *Transport {
	t := &Transport{
		cfg:     cfg,
		cluster: cfg.Cluster,
		sniffer: newSniffer(cfg),
		uris:    uris,
		request: make(chan *container, 200000),
		exit:    make(chan struct{}),

		lastRequestAt:  time.Now(),
		reload:         true,
		reloadAfter:    10000,
		resurrectAfter: 60,
	}

	t.conns = t.buildConns()

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

	lastRequestAt  time.Time
	resurrectAfter int64
	reloadAfter    int64
	reload         bool
	counter        int64
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
	for {
		select {
		case c := <-t.request:
			b := baggages.Get(t.doRequest(c))
			c.baggage <- b
		case <-t.exit:
			break
		}
	}
}

func (t *Transport) doRequest(c *container) (interface{}, error) {
	conn := t.conn()

	item, err := c.fun(conn)

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
	selector := &RoundRobinSelector{conns: conns}

	return &Conns{cc: conns, selector: selector}
}

func (t *Transport) conn() *Conn {
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

	var staleConns []*Conn
	conns := t.buildConns()
	oldConns := t.conns.all()

	for _, old := range oldConns {
		for _, new := range conns.cc {
			if reflect.DeepEqual(old, new) {
				staleConns = append(staleConns, old)
			}
		}
	}

	var newConns []*Conn
	for _, new := range conns.cc {
		keep := true
		for _, old := range oldConns {
			if reflect.DeepEqual(new, old) {
				keep = false
			}
		}

		if keep {
			newConns = append(newConns, new)
		}
	}

	conns.remove(staleConns...)
	conns.add(newConns...)

	t.conns = conns
}
