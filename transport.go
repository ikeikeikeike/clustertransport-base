package clustertransport

// Conn struct is
type Conn struct {
	client interface{}
}

// TransportConn interface is
type TransportConn interface {
	BuildConn(uri string, st *Transport) *Conn
}

// NewTransport is
func NewTransport(tconn TransportConn) *Transport {
	t := &Transport{
		tconn:   tconn,
		request: make(chan *container),
		exit:    make(chan struct{}),
	}

	t.buildConns()

	go t.run()
	return t
}

// Transport struct is
type Transport struct {
	uris    []string
	tconn   TransportConn
	conns   []*Conn
	request chan *container
	exit    chan struct{}
}

func (t *Transport) buildConns() {
	for _, uri := range t.uris {
		t.conns = append(t.conns, t.tconn.BuildConn(uri, t))
	}
}

func (t *Transport) getConn() *Conn {
	return t.conns[0]
}

// PerformRequest is
func (t *Transport) PerformRequest(fun func(conn *Conn) (interface{}, error)) (interface{}, error) {
	c := containers.Get()
	defer containers.Put(c)

	c.fun = fun
	t.request <- c

	b := <-c.baggage
	defer baggages.Put(b)

	item, err := b.item, b.err
	return item, err
}

func (t *Transport) doReq(c *container) (interface{}, error) {
	conn := t.getConn()

	item, err := c.fun(conn)

	return item, err
}

func (t *Transport) run() {
	for {
		select {
		case container := <-t.request:
			b := baggages.Get(t.doReq(container))
			container.baggage <- b
		case <-t.exit:
			break
		}
	}
}
