package clustertransport

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/ikeikeikeike/memdtest"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	memd, _ := memdtest.NewServer(true, memdtest.Config{"port": "34567"})
	code := m.Run()
	memd.Stop()

	os.Exit(code)
}

var getter = func(conn *Conn, keys ...interface{}) (interface{}, error) {
	client := conn.Client.(*memcache.Client)
	return client.Get(keys[0].(string))
}

var setter = func(conn *Conn, items ...interface{}) (interface{}, error) {
	client := conn.Client.(*memcache.Client)
	return nil, client.Set(items[0].(*memcache.Item))
}

func TestConfigure(t *testing.T) {
	cfg := NewConfig()
	cfg.Cluster = &esCluster{}
	cfg.Logger = log.Printf
	cfg.Debug = true

	ts := NewTransport(cfg, "127.0.0.1:34567")
	s := &nproxy{ts: ts, get: ts.Args(getter), set: ts.Args(setter)}

	configure := func(x int) {
		ts.Configure(func(cfg *Config) *Config {
			cfg.DiscoverTick = x
			return cfg
		})
	}

	x := 1
	for 100000 > x {
		item := memcache.Item{Key: "configure", Value: []byte(fmt.Sprint(s))}
		if err := s.Set(&item); err != nil {
			assert.NoError(t, err, "Error happened")
		}

		if _, err := s.Get("configure"); err != nil {
			assert.NoError(t, err, "Error happened")
		}

		if x%100 == 0 {
			configure(x)
		}

		x++
	}
}

func TestUnnested(t *testing.T) {
	cfg := NewConfig()
	cfg.Cluster = &esCluster{}
	cfg.Logger = log.Printf
	cfg.Debug = true

	ts := NewTransport(cfg, "127.0.0.1:34567")
	s := &nproxy{ts: ts, get: ts.Args(getter), set: ts.Args(setter)}

	x := 1
	for 100000 > x {
		item := memcache.Item{Key: "unnested", Value: []byte(fmt.Sprint(s))}
		if err := s.Set(&item); err != nil {
			assert.NoError(t, err, "Error happened")
		}

		if _, err := s.Get("unnested"); err != nil {
			assert.NoError(t, err, "Error happened")
		}

		x++
	}
}

type nproxy struct {
	ts  *Transport
	get func(...interface{}) (interface{}, error)
	set func(...interface{}) (interface{}, error)
}

func (s *nproxy) Get(key string) (*memcache.Item, error) {
	item, err := s.get(key)
	return item.(*memcache.Item), err
}

func (s *nproxy) Set(item *memcache.Item) error {
	_, err := s.set(item)
	return err
}
