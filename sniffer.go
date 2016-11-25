package clustertransport

import "time"

// SnifferMethod si
type SnifferMethod interface {
	CarryOut() []string
}

// NewSniffer is
func NewSniffer(method SnifferMethod) *Sniffer {
	s := &Sniffer{
		method:  method,
		receive: make(chan *container),
		exit:    make(chan struct{}),
		lost:    make(chan struct{}),
	}

	go s.run()
	return s
}

// Sniffer is
type Sniffer struct {
	method  SnifferMethod
	receive chan *container
	exit    chan struct{}
	lost    chan struct{}
	sniffed []string
}

// Sniffed is
func (s *Sniffer) Sniffed() ([]string, error) {
	c := containers.Get()
	defer containers.Put(c)

	s.receive <- c
	baggage := <-c.baggage

	return baggage.item.([]string), nil
}

// Exit is
func (s *Sniffer) Exit() {
	s.exit <- struct{}{}
}

func (s *Sniffer) sniffering() {
	s.sniffed = s.method.CarryOut()
}

func (s *Sniffer) run() {
	snifferTick := time.NewTicker(60 * time.Second)
	defer snifferTick.Stop()

	for {
		select {
		case container := <-s.receive:
			b := baggages.Get(s.sniffed, nil)
			container.baggage <- b
		case <-snifferTick.C:
			s.sniffering()
		case <-s.lost:
			s.sniffed = make([]string, 0)
		case <-s.exit:
			s.sniffed = make([]string, 0)
			break
		}
	}
}
