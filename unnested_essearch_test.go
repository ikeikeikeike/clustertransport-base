package clustertransport

import (
	"log"
	"testing"

	elastic "gopkg.in/olivere/elastic.v3"

	"github.com/stretchr/testify/assert"
)

func TestUnnestedESS(t *testing.T) {
	cfg := NewConfig()
	cfg.Cluster = &ElasticsearchCluster{}
	cfg.Logger = log.Printf
	cfg.Debug = true

	ts := NewTransport(cfg, "127.0.0.1:9201")

	get := ts.Args(func(conn *Conn, keys ...interface{}) (interface{}, error) {
		client := conn.Client.(*elastic.Client)
		res, _, err := client.Ping(conn.Uri).Do()
		return res, err
	})

	s := &esStorage{ts: ts, get: get}

	x := 1
	for 1000000 > x {
		if _, err := s.Get("unknownaaaaaaaa"); err != nil {
			assert.NoError(t, err, "Error happened")
		}

		x++
	}
}

type esStorage struct {
	ts  *Transport
	get func(...interface{}) (interface{}, error)
}

func (s *esStorage) Get(key string) (*elastic.PingResult, error) {
	result, err := s.get(key)
	return result.(*elastic.PingResult), err
}
