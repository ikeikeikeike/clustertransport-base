package clustertransport

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/pkg/errors"
)

// ElasticacheCluster implements for ClusterBase interface.
type ElasticacheCluster struct{}

// Sniff method returns node connection strings.
func (m *ElasticacheCluster) Sniff(connection *Conn) []string {
	in, errIn := make(chan []string), make(chan error)

	go func() {
		conn, err := net.Dial("tcp", connection.Uri)
		if err != nil {
			errIn <- err
		}
		defer conn.Close()
		fmt.Fprintf(conn, "config get cluster\r\n\r\n")

		text := []string{}
		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			t := string(scanner.Text())
			text = append(text, t)
			if t == "END" {
				break
			}
		}
		if err := scanner.Err(); err != nil {
			errIn <- err
		}
		if len(text) < 3 {
			errIn <- errors.New("too few a telnet resp")
		}

		var uris []string
		for _, info := range strings.Split(text[2], " ") {
			i := strings.Split(info, "|")
			host, _, port := i[0], i[1], i[2]

			uris = append(uris, fmt.Sprintf("%s:%s", host, port))
		}

		in <- uris
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case uris := <-in:
			return uris
		case _ = <-errIn:
			// pp.Println(err)
			return []string{}
		case <-ctx.Done():
			// pp.Println(ctx.Err().Error())
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
