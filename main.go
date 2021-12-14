package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
)

func main() {

	port := flag.Int("port", 8080, "HTTP port to serve on")
	dataDir := flag.String("datadir", "", "Path to data directory")
	cacheDir := flag.String("cachedir", "", "Path to the cache directory")
	withFsync := flag.Bool("with-fsync", false,
		"Whether to fsync newly written blobs")
	debug := flag.Bool("debug", false, "Enables some debugging features")
	var peers Peers

	flag.Var(&peers, "peer", "Provide url of peer")
	flag.Parse()

	s, err := NewServer(ServerConfig{
		Addr:      fmt.Sprintf(":%d", *port),
		DataDir:   *dataDir,
		CacheDir:  *cacheDir,
		WithFsync: *withFsync,
		Debug:     *debug,
		Peers:     peers,
	})

	if err != nil {
		log.Fatal(err)
	}

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}

type Peers []*url.URL

func (p *Peers) Set(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	*p = append(*p, u)
	return nil
}

func (p *Peers) String() string {
	strs := make([]string, 0, len(*p))

	for _, u := range *p {
		strs = append(strs, u.String())
	}
	return strings.Join(strs, " ")
}
