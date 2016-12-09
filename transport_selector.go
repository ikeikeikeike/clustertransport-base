package clustertransport

import (
	"math/rand"
	"time"
)

// RandomSelector is
type RandomSelector struct{}

// Select is
func (rs *RandomSelector) Select(conns []*Conn) *Conn {
	rand.Seed(time.Now().UTC().UnixNano())
	return conns[rand.Intn(len(conns))]
}

// RoundRobinSelector is
type RoundRobinSelector struct {
	current int
}

// Select is
func (rr *RoundRobinSelector) Select(conns []*Conn) *Conn {
	if rr.current >= len(conns) {
		rr.current = 0
	}

	conn := conns[rr.current]
	rr.current++

	return conn
}
