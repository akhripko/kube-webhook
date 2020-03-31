package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"git.syneforge.com/gin/k8s-acl-sv/options"
	"git.syneforge.com/gin/k8s-acl-sv/server/healthcheck"
	"git.syneforge.com/gin/k8s-acl-sv/server/httpsrv"
	"git.syneforge.com/gin/k8s-acl-sv/service"
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
	svc, err := service.New()
	if err != nil {
		log.Error("service init error:", err.Error())
		os.Exit(1)
	}

	// build http server
	httpSrv, err := httpsrv.New(config.HTTPPort, svc)
	if err != nil {
		log.Error("http service init error:", err.Error())
		os.Exit(1)
	}

	// build healthcheck server
	healthSrv := healthcheck.New(
		config.HealthCheckPort,
		httpSrv.HealthCheck,
	)

	// run servers
	httpSrv.Run(ctx, wg)
	healthSrv.Run(ctx, wg)

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
