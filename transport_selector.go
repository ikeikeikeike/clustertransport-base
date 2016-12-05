package clustertransport

import (
	"math/rand"
	"time"

	"github.com/k0kubun/pp"
)

// RandomSelector is
type RandomSelector struct{}

// Select is
func (rs *RandomSelector) Select(conns []*Conn) *Conn {
	rand.Seed(time.Now().Unix())
	return conns[rand.Intn(len(conns))]
}

// RoundRobinSelector is
type RoundRobinSelector struct {
	current int
}

// Select is
func (rr *RoundRobinSelector) Select(conns []*Conn) *Conn {
	defer func() {
		err := recover()
		if err != nil {
			pp.Println(rr.current, len(conns), err, conns)
		}
	}()

	if rr.current >= len(conns) {
		rr.current = 0
	}

	conn := conns[rr.current]
	rr.current++

	return conn
}
