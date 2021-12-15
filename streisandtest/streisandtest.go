package streisandtest

import (
	"github.com/Jille/errchain"
	"github.com/bertha/streisand"
	"net/http/httptest"
	"net/url"
	"os"
)

type Server struct {
	Streisand         streisand.Server
	Http              *httptest.Server
	DataDir, CacheDir string
}

type Servers struct {
	Servers []*Server
}

func NewServer(getPeers func() ([]*url.URL, error)) (*Server, error) {
	var err error
	conf := streisand.ServerConfig{
		WithFsync: true,
		Debug:     true,
		GetPeers:  getPeers,
	}

	conf.DataDir, err = os.MkdirTemp("", "streisandtest")
	if err != nil {
		return nil, err
	}
	conf.CacheDir, err = os.MkdirTemp("", "streisandtest")
	if err != nil {
		return nil, err
	}

	ss, err := streisand.NewServer(conf)
	if err != nil {
		return nil, err
	}
	s := &Server{
		Streisand: ss,
		Http:      httptest.NewServer(ss.Handler()),
		DataDir:   conf.DataDir,
		CacheDir:  conf.CacheDir,
	}
	return s, nil
}

func (s *Server) Close() error {
	var err error
	s.Http.Close()
	errchain.Append(&err, s.Streisand.Close())
	errchain.Append(&err, os.RemoveAll(s.DataDir))
	errchain.Append(&err, os.RemoveAll(s.CacheDir))
	return err
}

func NewServers(serverCount uint) (*Servers, error) {
	var err error
	s := &Servers{
		Servers: make([]*Server, serverCount),
	}
	for i := uint(0); i < serverCount; i++ {
		s.Servers[i], err = NewServer(s.getPeers)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Servers) Close() (err error) {
	for i := 0; i < len(s.Servers); i++ {
		errchain.Append(&err, s.Servers[i].Close())
	}
	return
}

func (s *Servers) getPeers() ([]*url.URL, error) {
	var err error
	result := make([]*url.URL, len(s.Servers))
	for i := 0; i < len(s.Servers); i++ {
		result[i], err = url.Parse(s.Servers[i].Http.URL)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
