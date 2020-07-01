package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/akhripko/kube-webhook/k8s-acl-sv/options"
	"github.com/akhripko/kube-webhook/k8s-acl-sv/provider/kube"
	"github.com/akhripko/kube-webhook/k8s-acl-sv/server/httpsrv"
	"github.com/akhripko/kube-webhook/k8s-acl-sv/server/infosrv"
	"github.com/akhripko/kube-webhook/k8s-acl-sv/service"
)

func main() {
	// read service config from os env
	config := options.ReadEnv()

	// init logger
	initLogger(config)

	log.Info("begin...")

	// prepare main context
	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)
	var wg = &sync.WaitGroup{}

	// build main service
	svc := service.New(config.SystemUsers, config.AdminUsers, kube.New())
	log.Debug("system users:", config.SystemUsers)
	log.Debug("admin users:", config.AdminUsers)

	// build http server
	httpSrv := httpsrv.New(config.HTTPPort, svc)
	// add tls conf
	if len(config.TLSCertFile) > 0 {
		log.Debug("set tls for http service")
		httpSrv.SetupTLS(config.TLSCertFile, config.TLSKeyFile)
	}

	// build info server
	infoSrv := infosrv.New(
		config.HealthCheckPort,
		httpSrv.HealthCheck,
	)

	// run servers
	httpSrv.Run(ctx, wg)
	infoSrv.Run(ctx, wg)

	wg.Wait()
	log.Info("end")
}

func initLogger(config *options.Config) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)

	switch strings.ToLower(config.LogLevel) {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}

func setupGracefulShutdown(stop func()) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		log.Error("Got Interrupt signal")
		stop()
	}()
}
