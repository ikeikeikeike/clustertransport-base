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

	ts := NewTransport(cfg, "http://127.0.0.1:9200")

	get := ts.Args(func(conn *Conn, keys ...interface{}) (interface{}, error) {
		client := conn.Client.(*elastic.Client)
		res, _, err := client.Ping(conn.Uri).Do()

		return res, err
	})

	s := &esStorage{ts: ts, get: get}

	x := 1
	for 100000 > x {
		if _, err := s.Get(); err != nil {
			assert.NoError(t, err, "Error happened")
		}

		x++
	}
}

type esStorage struct {
	ts  *Transport
	get func(...interface{}) (interface{}, error)
}

func (s *esStorage) Get() (*elastic.PingResult, error) {
	result, err := s.get()
	return result.(*elastic.PingResult), err
}
