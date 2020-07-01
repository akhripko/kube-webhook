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

func New(port int, service Service) *HTTPSrv {
	// build http server
	httpSrv := http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	// build HTTPSrv
	var srv HTTPSrv
	srv.setupHTTP(&httpSrv)
	srv.service = service

	return &srv
}

type HTTPSrv struct {
	http      *http.Server
	runErr    error
	readiness bool
	service   Service
	tls       *tlsConf
}

type tlsConf struct {
	certFile string
	keyFile  string
}

func (s *HTTPSrv) setupHTTP(srv *http.Server) {
	srv.Handler = s.buildHandler()
	s.http = srv
}

func (s *HTTPSrv) buildHandler() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/validate", s.validateHandleFunc)
	r.HandleFunc("/mutate", s.mutateHandleFunc)
	r.HandleFunc("/mutate/add-owners", s.addOwnersHandleFunc)
	r.HandleFunc("/add-owners", s.addOwnersHandleFunc)

	return r
}

func (s *HTTPSrv) SetupTLS(cerfFile string, keyFile string) {
	s.tls = &tlsConf{
		certFile: cerfFile,
		keyFile:  keyFile,
	}
}

func (s *HTTPSrv) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Info("http server: begin run")

	go func() {
		defer wg.Done()
		log.Info("http server: addr=", s.http.Addr)
		if s.tls != nil {
			log.Debug("http server: tls certFile=", s.tls.certFile, " keyFile=", s.tls.keyFile)
			s.runErr = s.http.ListenAndServeTLS(s.tls.certFile, s.tls.keyFile)
		} else {
			s.runErr = s.http.ListenAndServe()
		}
		log.Info("http server: end run > ", s.runErr)
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
