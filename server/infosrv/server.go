package infosrv

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Service struct {
	http *http.Server
}

func New(port int, healthChecks ...func() error) *Service {
	return &Service{
		http: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: buildHandler(healthChecks),
		},
	}
}

func (s *Service) Run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	log.Info("info Service: begin run")

	go func() {
		defer wg.Done()
		log.Debug("info Service addr:", s.http.Addr)
		err := s.http.ListenAndServe()
		log.Info("info service: end run > ", err)
	}()

	go func() {
		<-ctx.Done()
		sdCtx, _ := context.WithTimeout(context.Background(), 5*time.Second) // nolint
		err := s.http.Shutdown(sdCtx)
		if err != nil {
			log.Info("info service shutdown (", err, ")")
		}
	}()
}

func buildHandler(healthChecks []func() error) http.Handler {
	handler := http.NewServeMux()
	handler.HandleFunc("/version", serveVersion)
	var checks = func(w http.ResponseWriter, _ *http.Request) { serveCheck(w, healthChecks) }
	handler.HandleFunc("/", checks)
	handler.HandleFunc("/health", checks)
	handler.HandleFunc("/ready", checks)
	return handler
}

func writeFile(file string, w http.ResponseWriter) {
	if data, err := ioutil.ReadFile(file); err == nil { // nolint
		w.WriteHeader(http.StatusOK)
		w.Write(data) // nolint
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func serveCheck(w http.ResponseWriter, checks []func() error) {
	var wh bool
	writeHeaderOnce := func(statusCode int) {
		if !wh {
			w.WriteHeader(statusCode)
			wh = true
		}
	}

	for _, check := range checks {
		if err := check(); err != nil {
			writeHeaderOnce(http.StatusInternalServerError)
			w.Write([]byte(err.Error() + "\n\n")) // nolint
		}
	}
	writeHeaderOnce(http.StatusNoContent)
}
