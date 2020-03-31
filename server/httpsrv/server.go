package httpsrv

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func New(port int, service Service) (*HTTPSrv, error) {
	// build http server
	httpSrv := http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	// build HTTPSrv
	var srv HTTPSrv
	srv.setupHTTP(&httpSrv)
	srv.service = service

	return &srv, nil
}

type HTTPSrv struct {
	http      *http.Server
	runErr    error
	readiness bool
	service   Service
}

func (s *HTTPSrv) setupHTTP(srv *http.Server) {
	srv.Handler = s.buildHandler()
	s.http = srv
}

func (s *HTTPSrv) buildHandler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/mutate", s.mutateHandleFunc).Methods("GET")

	return r
}

func (s *HTTPSrv) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Info("http server: begin run")

	go func() {
		defer wg.Done()
		log.Debug("http server: addr=", s.http.Addr)
		err := s.http.ListenAndServe()
		s.runErr = err
		log.Info("http server: end run > ", err)
	}()

	go func() {
		<-ctx.Done()
		sdCtx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint
		err := s.http.Shutdown(sdCtx)
		if err != nil {
			log.Info("http server shutdown (", err, ")")
		}
	}()

	s.readiness = true
}

func (s *HTTPSrv) HealthCheck() error {
	if !s.readiness {
		return errors.New("http server is't ready yet")
	}
	if s.runErr != nil {
		return errors.New("http server: run error")
	}
	return nil
}
