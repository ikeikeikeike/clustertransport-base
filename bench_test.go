package clustertransport

import (
	"log"
	"runtime"
	"testing"

	elastic "gopkg.in/olivere/elastic.v3"

	"github.com/k0kubun/pp"
)

func BenchmarkRace(b *testing.B) {
	cfg := NewConfig()
	cfg.DiscoverAfter = 1000000
	cfg.Cluster = &ElasticsearchCluster{}
	cfg.Logger = log.Printf
	cfg.Debug = true

	ts := NewTransport(cfg, "http://127.0.0.1:9200")

	runtime.GOMAXPROCS(runtime.NumCPU())

	fn := func() (interface{}, error) {
		item, err := ts.Req(func(conn *Conn) (interface{}, error) {
			client := conn.Client.(*elastic.Client)

			res, _, err := client.Ping(conn.Uri).Do()
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
	for 1000000000 > x {
		_, err := fn()

		if err != nil {
			pp.Println("couldnt request: ", err)
		}

		if x%97891 == 0 {
			funcy(x)
		}
		x++
		// log.Println(x)
		// time.Sleep(700 * time.Millisecond)
	}
}
