package clustertransport

import (
	"net/http"
	"time"

	elastic "gopkg.in/olivere/elastic.v3"
)

// ElasticsearchCluster is
type ElasticsearchCluster struct{}

// Sniff is
func (m *ElasticsearchCluster) Sniff() []string {
	return []string{"1", "2"}
}

// Conn is
func (m *ElasticsearchCluster) Conn(uri string, st *Transport) *Conn {
	var options []elastic.ClientOptionFunc
	options = append(options, elastic.SetHttpClient(&http.Client{Timeout: 5 * time.Second}))
	options = append(options, elastic.SetURL(uri))
	options = append(options, elastic.SetSniff(false))

	client, _ := elastic.NewClient(options...)
	return &Conn{client: client}
}
