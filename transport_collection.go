package clustertransport

import (
	"errors"
	"reflect"
	"sort"
)

// Conns is
type Conns struct {
	cc       []*Conn
	selector SelectorBase
}

func (cs *Conns) uris() []string {
	uris := []string{}
	for _, c := range cs.cc {
		uris = append(uris, c.Uri)
	}

	return uris
}

func (cs *Conns) alives() []*Conn {
	var conns []*Conn
	for _, c := range cs.cc {
		if c.dead {
			continue
		}

		conns = append(conns, c)
	}

	return conns
}

func (cs *Conns) deads() []*Conn {
	var conns []*Conn
	for _, c := range cs.cc {
		if !c.dead {
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
		if len(deads) > 0 {
			return nil, errors.New("There's no connection already.")
		}

		sort.Sort(sort.Reverse(connsSort(deads)))
		deads[0].alive()
	}

	return cs.selector.Select(), nil
}

func (cs *Conns) add(conns ...*Conn) {
	cs.cc = append(cs.cc, conns...)
}

func (cs *Conns) remove(conns ...*Conn) {
	var still []*Conn

	for _, top := range cs.cc {
		keep := true

		for _, bottom := range conns {
			if reflect.DeepEqual(top, bottom) {
				keep = false
			}
		}

		if keep {
			still = append(still, top)
		}
	}

	cs.cc = still
}

type connsSort []*Conn

func (f connsSort) Len() int           { return len(f) }
func (f connsSort) Less(i, j int) bool { return f[i].failures < f[j].failures }
func (f connsSort) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
