package clustertransport

import (
	"log"
	"testing"

	"github.com/kr/pretty"

	elastic "gopkg.in/olivere/elastic.v3"
)

func BenchmarkRace(b *testing.B) {
	cfg := NewConfig()
	cfg.Cluster = &esCluster{}
	cfg.Logger = log.Printf
	cfg.Debug = true

	ts := NewTransport(cfg, "http://127.0.0.1:9200")

	fn := func() (interface{}, error) {
		item, err := ts.Req(func(conn *Conn) (interface{}, error) {
			client := conn.Client.(*elastic.Client)

			res, _, err := client.Ping(conn.URI).Do()
			return res, err
		})

		return item, err
	}

	funcy := func(x int) {
		ts.Configure(func(cfg *Config) *Config {
			// cfg.DiscoverTick = x
			return cfg
		})
	}

	x := 1
	for 500000 > x {
		_, err := fn()

		if err != nil {
			pretty.Println("couldnt request: ", err)
		}

		if x%97891 == 0 {
			funcy(x)
		}
		x++
	}
}
