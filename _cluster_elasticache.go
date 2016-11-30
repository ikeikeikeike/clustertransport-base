package clustertransport

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
)

// ElasticacheCluster implements for ClusterBase interface.
type ElasticacheCluster struct{}

// Sniff method returns node connection strings.
func (m *ElasticacheCluster) Sniff(connection *Conn) []string {
	conn, err := net.Dial("tcp", connection.Uri)
	if err != nil {
		return []string{}
	}

	in, errIn := make(chan []string), make(chan error)

	go func() {
		connReader := bufio.NewReader(conn)
		fmt.Fprintf(conn, "config get cluster\r\n\r\n")

		i, _ := connReader.ReadString('\n')
		if err != nil {
			errIn <- err
			return
		}

		pp.Println(i)
		in <- []string{}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case uris := <-in:
			return uris
		case err := <-errIn:
			_ = err
			return []string{}
		case <-ctx.Done():
			_ = ctx.Err()
			return []string{}
		}
	}
}

// Conn method returns one of cluster system connection.
func (m *ElasticacheCluster) Conn(uri string, st *Transport) (*Conn, error) {
	conn, err := net.Dial("tcp", uri)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to launch memcached")
	}

	fmt.Fprintf(conn, "version\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')

	if err != nil || !strings.HasPrefix(status, "VERSION") {
		return nil, errors.Wrap(err, "Failed to launch memcached")
	}

	return &Conn{Client: memcache.New(uri)}, nil
}
