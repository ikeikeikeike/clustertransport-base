
# Cluster Transport
### Message passing based client(transport) on cluster systems.

It handles connecting to multiple nodes in the cluster which there's management interface.

Features overview:

- [Pluggable transport implementation, customizable and extendable](#pluggable-transport-implementation-customizable-and-extendable)
- [Plugabble connection selection strategies (round-robin, random, custom)](#plugabble-connection-selection-strategies-round-robin-random-custom)
- [Pluggable logging and tracing](#pluggable-logging-and-tracing)
- [Request retries and dead connections handling](#request-retries-and-dead-connections-handling)
- [Node discovering (based on cluster state) on errors or on demand](#node-discovering-based-on-cluster-state-on-errors-or-on-demand)

## Installation

```bash
$ go get github.com/ikeikeikeike/clustertransport-base
```

## Pluggable transport implementation, customizable and extendable

Needs to implement `Sniff` and `Conn` methods which connects Cluster System.
A example that's connecting Elasticsearch instead of elastic client's sniffer.

```go
package mypkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	elastic "gopkg.in/olivere/elastic.v3"
)

// ElasticsearchCluster implements for ClusterBase interface.
type ElasticsearchCluster struct{}

// Sniff method returns node connection strings.
func (m *ElasticsearchCluster) Sniff(conn *Conn) []string {
	resp, err := http.Get(conn.URI + "/_nodes/http")
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	var uris []string
	var info *elastic.NodesInfoResponse

	if err := json.NewDecoder(resp.Body).Decode(&info); err == nil {
		if len(info.Nodes) > 0 {
			for _, node := range info.Nodes {
				if node.HTTPAddress != "" {
					uris = append(uris, fmt.Sprintf("http://%s", node.HTTPAddress))
				}
			}
		}
	}

	return uris
}

// Conn method returns one of cluster system connection.
func (m *ElasticsearchCluster) Conn(uri string, st *Transport) (*Conn, error) {
	var options []elastic.ClientOptionFunc
	options = append(options, elastic.SetHttpClient(&http.Client{Timeout: 5 * time.Second}))
	options = append(options, elastic.SetURL(uri))
	options = append(options, elastic.SetSniff(false))

	client, err := elastic.NewClient(options...)
	return &Conn{Client: client}, err
}
```

- [Elasticache example](https://github.com/ikeikeikeike/clustertransport-base/blob/master/_cluster_elasticache.go)
- [Elasticsearch example](https://github.com/ikeikeikeike/clustertransport-base/blob/master/_cluster_elasticsearch.go)

## Configuration

```go
cfg := ctbase.NewConfig()
cfg.Cluster = &ctbase.ElasticsearchCluster{}
cfg.Logger = log.Printf
...
```

#### Default configuration

```go
func NewConfig() *Config {
	return &Config{
		Selector:       &RoundRobinSelector{},
		Logger:         PrintNothing,
		Discover:       true,
		DiscoverTick:   120,    // Discovers nodes per 120 sec
		DiscoverAfter:  100000, // Discovers nodes after passed 100,000 requests
		RetryOnFailure: false,  // Retrying asap when one of connection failed
		ResurrectAfter: 30,     // Tries to resurrect some of connections when Cluster Transport hasn't request to cluster system until it passed 30 sec.
		MaxRetries:     5,      // Tries to retry's number for http request
	}
}
```

## Usage

...Later

```go
package main

import (
	elastic "gopkg.in/olivere/elastic.v3"

	ctbase "github.com/ikeikeikeike/clustertransport-base"
	"github.com/kr/pretty"
)

var ts *ctbase.Transport

func init() {
    cfg := ctbase.NewConfig()
    cfg.Cluster = &ctbase.ElasticsearchCluster{}
    cfg.Logger = log.Printf

	ts = ctbase.NewTransport(cfg, "http://127.0.0.1:9200")
}

func main() {
	item, err := ts.Req(func(conn *ctbase.Conn) (interface{}, error) {
		client := conn.Client.(*elastic.Client)

		res, _, err := client.Ping(conn.URI).Do()
		return res, err
	})

	pretty.Println(item.(*elastic.PingResult), err)
}
```
Output:

```
&elastic.PingResult{
    Name:        "Sunstreak",
    ClusterName: "elasticsearch",
    Version:     struct { Number string "json:\"number\""; BuildHash string "json:\"build_hash\""; BuildTimestamp string "json:\"build_timestamp\""; BuildSnapshot bool "json:\"build_snapshot\""; LuceneVersion string "json:\"lucene_version\"" }{Number:"2.4.2", BuildHash:"161c65a337d4b422ac0c805f284565cf2014bb84", BuildTimestamp:"2016-11-17T11:51:03Z", BuildSnapshot:false, LuceneVersion:"5.5.2"},
    TagLine:     "You Know, for Search",
} nil
```

#### Hates the nested

...Later

```go
import "github.com/bradfitz/gomemcache/memcache"

type Storage struct {
	ts  *ctbase.Transport
	get func(interface{}) (interface{}, error)
	set func(...interface{}) (interface{}, error)
}

func (s *Storage) Get(key string) (*memcache.Item, error) {
	item, err := s.get(key)
	return item.(*memcache.Item), err
}

func (s *Storage) Set(key, value string) error {
	_, err := s.set(key, value)
	return err
}

func NewStorage() *Storage {
	cfg := ctbase.NewConfig()
	cfg.Cluster = &ctbase.ElasticacheCluster{}
	cfg.Logger = log.Printf

	ts := ctbase.NewTransport(cfg, "cluster-host:11211")

	get := ts.Arg(func(conn *ctbase.Conn, arg interface{}) (interface{}, error) {
		client := conn.Client.(*memcache.Client)
		return client.Get(arg.(string))
	})

	set := ts.Args(func(conn *ctbase.Conn, args ...interface{}) (interface{}, error) {
		client := conn.Client.(*memcache.Client)
        key, value := args[0], args[1]

		return nil, client.Set(&memcache.Item{Key: key, Value: []byte(value)})
	})

	return &Storage{ts: ts, get: get, set: set}
}
```

...Later

```go
storage := NewStorage()

storage.Set("egg", "by array")
storage.Get("egg")
```

## Request retries and dead connections handling

Cluster Transport is able to handle dead connections. Therefore, for handling it returns `*os.SyscallError`, `*url.Error` and `*net.OpError`, or otherwise it's able to return `*clustertransport.Econnrefused` in explicitly.

```go
item, err := ts.Req(func(conn *ctbase.Conn) (interface{}, error) {
    client := conn.Client.(*memcache.Client)

    res, err := client.Get("somekey")
    if err != nil && err == memcached.ErrNoServers {
        return res, &ctbase.Econnrefused{"node econnrefused"}
    }

    return res, err
})
```

## Plugabble connection selection strategies (round-robin, random, custom)

...Later

## Node discovering (based on cluster state) on errors or on demand

...Later

## Pluggable logging and tracing

```go
cfg := ctbase.NewConfig()
cfg.Logger = log.Printf
ts := ctbase.NewTransport(cfg, "http://127.0.0.1:9200")
```

```go
ts := ctbase.NewTransport(ctbase.NewConfig(), "http://127.0.0.1:9200")
ts.Configure(func(cfg *ctbase.Config) *ctbase.Config {
    cfg.Logger = log.Printf
    return cfg
})
```
