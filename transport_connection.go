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
	Failures  int64 // Counter
	Dead      bool
	rebirth   int64 // Timeout Seconds
	deadSince time.Time
}

func (c *Conn) terminate() {
	c.Dead = true
	c.Failures++
	c.deadSince = time.Now()
}

func (c *Conn) isDead() bool {
	return c.Dead
}

func (c *Conn) alive() {
	c.Dead = false
}

func (c *Conn) healthy() {
	c.Dead = false
	c.Failures = 0
}

func (c *Conn) resurrect() {
	if c.isResurrectable() {
		c.alive()
	}
}

func (c *Conn) isResurrectable() bool {
	left := c.deadSince.Unix()
	right := c.rebirth * int64(math.Pow(float64(2), float64(c.Failures-1)))
	return time.Now().Unix() > (left + right)
}
