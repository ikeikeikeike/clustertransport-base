package clustertransport

func newSniffer(cfg *Config, conns *Conns) *Sniffer {
	s := &Sniffer{
		cfg:     cfg,
		conns:   conns,
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
	conns   *Conns
	receive chan *container
	exit    chan struct{}
	lost    chan struct{}
	sniffed []string
}

// Sniffed is
func (s *Sniffer) Sniffed() ([]string, error) {
	c := &container{baggage: make(chan *baggage)}

	s.receive <- c
	b := <-c.baggage

	return b.item.([]string), nil
}

// Exit is
func (s *Sniffer) Exit() {
	s.exit <- struct{}{}
}

func (s *Sniffer) sniff() {
	conn, err := s.conns.conn()
	if err != nil {
		return
	}

	s.sniffed = s.cfg.Cluster.Sniff(conn)
}

func (s *Sniffer) run() {
	for {
		select {
		case c := <-s.receive:
			if len(s.sniffed) <= 0 {
				s.sniff()
			}
			b := baggages.Get(s.sniffed, nil)
			c.baggage <- b
		case <-s.lost:
			s.sniffed = make([]string, 0)
		case <-s.exit:
			s.sniffed = make([]string, 0)
			break
		}
	}
}
