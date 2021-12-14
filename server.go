package main

import (
	"net/http"
	"net/url"
	"sync"

	"github.com/Jille/convreq"
	"github.com/bertha/streisand/diskstore"
)

type Server interface {
	Run() error
}

type ServerConfig struct {
	DataDir, CacheDir string
	WithFsync         bool
	Addr              string
	Debug             bool
	Peers             []*url.URL
}

func NewServer(conf ServerConfig) (Server, error) {
	hmux := http.NewServeMux()

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
		httpServer: &http.Server{
			Addr:    conf.Addr,
			Handler: hmux,
		},
	}

	s.store.Initialize()

	if err := s.xors.Initialize(); err != nil {
		return nil, err
	}

	hmux.HandleFunc("/blob/", convreq.Wrap(func(r *http.Request) convreq.HttpResponse {
		return s.handleGetBlob(r, true)
	}))
	hmux.HandleFunc("/internal/blob/", convreq.Wrap(func(r *http.Request) convreq.HttpResponse {
		return s.handleGetBlob(r, false)
	}))
	hmux.HandleFunc("/upload", convreq.Wrap(s.handlePostBlob))
	http.HandleFunc("/internal/upload", convreq.Wrap(s.handleInternalPostBlob))
	hmux.HandleFunc("/list", convreq.Wrap(s.handleGetList))
	hmux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
	})

	if s.conf.Debug {
		hmux.HandleFunc("/debug/add-xor",
			convreq.Wrap(s.handleDebugAddXor))
	}

	return &s, nil
}

type server struct {
	conf ServerConfig

	store      *diskstore.Store
	xors       *XorStore
	mutex      sync.RWMutex
	httpServer *http.Server
}

func (s *server) Run() error {
	return s.httpServer.ListenAndServe()
}
