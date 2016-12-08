package clustertransport

import (
	"log"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
)

func TestUnnestedESC(t *testing.T) {
	cfg := NewConfig()
	cfg.Cluster = &ElasticacheCluster{}
	cfg.Logger = log.Printf
	cfg.Debug = true

	ts := NewTransport(cfg, "127.0.0.1:11211")

	get := ts.Args(func(conn *Conn, keys ...interface{}) (interface{}, error) {
		client := conn.Client.(*memcache.Client)
		return client.Get(keys[0].(string))
	})

	set := ts.Args(func(conn *Conn, items ...interface{}) (interface{}, error) {
		client := conn.Client.(*memcache.Client)
		return nil, client.Set(items[0].(*memcache.Item))
	})

	s := &escStorage{ts: ts, get: get, set: set}

	item := memcache.Item{Key: "unknownaaaaaaaa", Value: []byte("byte array")}

	x := 1
	for 1000000 > x {
		if _, err := s.Get("unknownaaaaaaaa"); err != nil {
			assert.NoError(t, err, "Error happened")
		}

		if err := s.Set(&item); err != nil {
			assert.NoError(t, err, "Error happened")
		}

		x++
	}
}

type escStorage struct {
	ts  *Transport
	get func(...interface{}) (interface{}, error)
	set func(...interface{}) (interface{}, error)
}

func (s *escStorage) Get(key string) (*memcache.Item, error) {
	item, err := s.get(key)
	return item.(*memcache.Item), err
}

func (s *escStorage) Set(item *memcache.Item) error {
	_, err := s.set(item)
	return err
}
