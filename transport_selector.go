package clustertransport

import (
	"math/rand"
	"time"
)

type RandomSelector struct {
	conns []*Conn
}

func (rs *RandomSelector) Select() *Conn {
	rand.Seed(time.Now().Unix())
	return rs.conns[rand.Intn(len(rs.conns))]
}

type RoundRobinSelector struct {
	conns   []*Conn
	current int
}

func (rr *RoundRobinSelector) Select() *Conn {
	conn := rr.conns[rr.current]

	rr.current++
	if rr.current >= len(rr.conns) {
		rr.current = 0
	}

	return conn
}
