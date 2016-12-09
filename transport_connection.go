package clustertransport

import (
	"math"
	"time"
)

// Conn struct is
type Conn struct {
	Client    interface{}
	URI       string
	Failures  int64 // Counter
	Dead      bool
	rebirth   int64
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
	right := int64(math.Pow(float64(2), float64(c.Failures-1)))
	return time.Now().Unix() > (20 + left + right)
}
