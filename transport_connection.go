package clustertransport

import (
	"math"
	"time"
)

// NewConn is
func NewConn() *Conn {
	return &Conn{
		rebirth: 60,
	}
}

// Conn struct is
type Conn struct {
	Client    interface{}
	Uri       string
	rebirth   int64 // Timeout Seconds
	failures  int64 // Counter
	dead      bool
	deadSince time.Time
}

func (c *Conn) terminate() {
	c.dead = true
	c.failures++
	c.deadSince = time.Now()
}

func (c *Conn) isDead() bool {
	return c.dead
}

func (c *Conn) alive() {
	c.dead = false
}

func (c *Conn) healthy() {
	c.dead = false
	c.failures = 0
}

func (c *Conn) resurrect() {
	if c.isResurrectable() {
		c.alive()
	}
}

func (c *Conn) isResurrectable() bool {
	left := c.deadSince.Unix()
	right := c.rebirth * int64(math.Pow(float64(2), float64(c.failures-1)))
	return time.Now().Unix() > (left + right)
}
