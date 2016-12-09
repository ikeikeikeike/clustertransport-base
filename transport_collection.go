package clustertransport

import (
	"errors"
	"sort"
)

// Conns handles cluster system connection as collection.
type Conns struct {
	cfg      *Config
	cc       []*Conn
	selector SelectorBase
}

func (cs *Conns) uris() []string {
	uris := []string{}
	for _, c := range cs.all() {
		uris = append(uris, c.URI)
	}

	return uris
}

func (cs *Conns) alives() []*Conn {
	conns := make([]*Conn, 0)
	for _, c := range cs.all() {
		if c.Dead {
			continue
		}

		conns = append(conns, c)
	}

	return conns
}

func (cs *Conns) deads() []*Conn {
	conns := make([]*Conn, 0)
	for _, c := range cs.all() {
		if !c.Dead {
			continue
		}

		conns = append(conns, c)
	}

	return conns
}

func (cs *Conns) all() []*Conn {
	return cs.cc
}

func (cs *Conns) conn() (*Conn, error) {
	if len(cs.alives()) <= 0 {
		deads := cs.deads()
		if len(deads) <= 0 {
			return nil, errors.New("There's no connection already")
		}

		sort.Sort(sort.Reverse(connsSort(deads)))
		deads[0].alive()

		cs.cfg.Logger("Resurrect a connection via %s (failures:%d deadSince:%v)",
			deads[0].URI, deads[0].Failures, deads[0].deadSince)
	}

	return cs.selector.Select(cs.alives()), nil
}

type connsSort []*Conn

func (f connsSort) Len() int           { return len(f) }
func (f connsSort) Less(i, j int) bool { return f[i].Failures > f[j].Failures }
func (f connsSort) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
