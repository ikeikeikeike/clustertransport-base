
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
	resp, err := http.Get(conn.Uri + "/_nodes/http")
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

[Full example](https://github.com/ikeikeikeike/clustertransport-base/blob/master/_cluster_elasticsearch.go)


## Configuration

...Later

## Usage

```go
package main

import (
	elastic "gopkg.in/olivere/elastic.v3"

	ct "github.com/ikeikeikeike/clustertransport-base"
	"github.com/kr/pretty"
)

var ts *ct.Transport

func init() {
	cfg := &ct.Config{Cluster: &ElasticsearchCluster{}}
	ts = ct.NewTransport(cfg, "http://127.0.0.1:9200")
}

func main() {
	item, err := ts.Req(func(conn *ct.Conn) (interface{}, error) {
		client := conn.Client.(*elastic.Client)

		res, _, err := client.Ping(conn.Uri).Do()
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

## Request retries and dead connections handling

Cluster Transport is able to handle dead connections. Therefore, for handling it returns `*os.SyscallError` error which wraps `syscall.ECONNREFUSED`. Commonly those are wrapped from `*url.Error` and `*net.OpError`, or otherwise it's able to return `*clustertransport.Econnrefused` in explicitly.

```go
item, err := ts.Req(func(conn *ct.Conn) (interface{}, error) {
    client := conn.Client.(*any.Client)

    res, err := client.Get("users")
    if err != nil {
        return res, &ct.Econnrefused{"node econnrefused"}
    }

    return res, err
})
```

## Plugabble connection selection strategies (round-robin, random, custom)

...Later

## Node discovering (based on cluster state) on errors or on demand

...Later

## Pluggable logging and tracing

...Later
