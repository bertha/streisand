package streisand

import (
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/Jille/convreq"
	"github.com/Jille/errchain"
	"github.com/bertha/streisand/diskstore"
)

type Server interface {
	http.Handler
	io.Closer
}

type PeersFunc func() ([]*url.URL, error)

type ServerConfig struct {
	DataDir, CacheDir string
	WithFsync         bool
	Debug             bool
	GetPeers          PeersFunc
}

func NewServer(conf ServerConfig) (Server, error) {
	s := server{
		conf: conf,
		store: &diskstore.Store{
			Path:          conf.DataDir,
			BitsPerFolder: []uint8{8, 8},
			Fsync:         conf.WithFsync,
		},
		xors: &XorStore{
			LayerCount: 6,
			LayerDepth: 4,
			Path:       conf.CacheDir,
		},
		hmux: http.NewServeMux(),
	}

	s.store.Initialize()

	if err := s.xors.Initialize(); err != nil {
		return nil, err
	}

	s.hmux.HandleFunc("/blob/", convreq.Wrap(func(r *http.Request) convreq.HttpResponse {
		return s.handleGetBlob(r, true)
	}))
	s.hmux.HandleFunc("/internal/blob/", convreq.Wrap(func(r *http.Request) convreq.HttpResponse {
		return s.handleGetBlob(r, false)
	}))
	s.hmux.HandleFunc("/upload", convreq.Wrap(s.handlePostBlob))
	s.hmux.HandleFunc("/internal/upload", convreq.Wrap(s.handleInternalPostBlob))
	s.hmux.HandleFunc("/list", convreq.Wrap(s.handleGetList))
	s.hmux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
	})

	if s.conf.Debug {
		s.hmux.HandleFunc("/debug/add-xor",
			convreq.Wrap(s.handleDebugAddXor))
	}

	return &s, nil
}

type server struct {
	conf ServerConfig

	store *diskstore.Store
	xors  *XorStore
	mutex sync.RWMutex
	hmux  *http.ServeMux
}

func (s *server) Close() (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	errchain.Call(&err, s.xors.Close)

	return err
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.hmux.ServeHTTP(w, r)
}
