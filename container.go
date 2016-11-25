package clustertransport

import "sync"

type container struct {
	baggage chan *baggage
	fun     func(conn *Conn) (interface{}, error)
}

var defcontainer = &container{
	baggage: make(chan *baggage),
}

func (c *container) reset() {
	c.baggage = defcontainer.baggage
}

type containerPool struct {
	sync.Pool
}

// Get is
func (cp *containerPool) Get() *container {
	return cp.Pool.Get().(*container)
}

// Put is
func (cp *containerPool) Put(c *container) {
	c.reset()
	cp.Pool.Put(c)
}

var containers = &containerPool{
	Pool: sync.Pool{New: func() interface{} {
		return &container{baggage: make(chan *baggage)}
	}},
}


type baggage struct {
	item interface{}
	err  error
}

var defbaggage = &baggage{}

func (b *baggage) reset() {
	b.item = defbaggage.item
	b.err = defbaggage.err
}

type baggagePool struct {
	sync.Pool
}

// Get is
func (bp *baggagePool) Get(item interface{}, err error) *baggage {
	b := bp.Pool.Get().(*baggage)
	b.item = item
	b.err = err

	return b
}

// Put is
func (bp *baggagePool) Put(b *baggage) {
	b.reset()
	bp.Pool.Put(b)
}

var baggages = &baggagePool{
	Pool: sync.Pool{New: func() interface{} {
		return &baggage{}
	}},
}
