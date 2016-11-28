package clustertransport

import "time"

func newSniffer(cfg *Config) *Sniffer {
	s := &Sniffer{
		cluster: cfg.Cluster,
		receive: make(chan *container),
		exit:    make(chan struct{}),
		lost:    make(chan struct{}),
	}

	go s.run()
	return s
}

// Sniffer is
type Sniffer struct {
	cfg     *Config
	cluster ClusterBase
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

func (s *Sniffer) sniff() {
	s.sniffed = s.cluster.Sniff()
}

func (s *Sniffer) run() {
	sniffTick := time.NewTicker(60 * time.Second)
	defer sniffTick.Stop()

	for {
		select {
		case c := <-s.receive:
			if len(s.sniffed) <= 0 {
				s.sniff()
			}
			b := baggages.Get(s.sniffed, nil)
			c.baggage <- b
		case <-sniffTick.C:
			s.sniff()
		case <-s.lost:
			s.sniffed = make([]string, 0)
		case <-s.exit:
			s.sniffed = make([]string, 0)
			break
		}
	}
}
