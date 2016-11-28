package clustertransport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	elastic "gopkg.in/olivere/elastic.v3"
)

// ElasticsearchCluster is
type ElasticsearchCluster struct{}

// Sniff is
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

// Conn is
func (m *ElasticsearchCluster) Conn(uri string, st *Transport) *Conn {
	var options []elastic.ClientOptionFunc
	options = append(options, elastic.SetHttpClient(&http.Client{Timeout: 5 * time.Second}))
	options = append(options, elastic.SetURL(uri))
	options = append(options, elastic.SetSniff(false))

	client, _ := elastic.NewClient(options...)
	return &Conn{Client: client}
}
