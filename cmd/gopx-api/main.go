package main

import (
	golog "log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"gopx.io/gopx-api/pkg/config"
	"gopx.io/gopx-api/pkg/route"
	"gopx.io/gopx-common/log"
)

var serverLogger = golog.New(os.Stdout, "", golog.Ldate|golog.Ltime|golog.Lshortfile)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	startServer()
}

func startServer() {
	switch {
	case config.Service.UseHTTP && config.Service.UseHTTPS:
		go startHTTP()
		startHTTPS()
	case config.Service.UseHTTP:
		startHTTP()
	case config.Service.UseHTTPS:
		startHTTPS()
	default:
		log.Fatal("Error: no listener is specified in service config file")
	}
}

func startHTTP() {
	addr := httpAddr()
	r := route.Router()
	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadTimeout:       config.Service.ReadTimeout * time.Second,
		ReadHeaderTimeout: config.Service.ReadTimeout * time.Second,
		WriteTimeout:      config.Service.WriteTimeout * time.Second,
		IdleTimeout:       config.Service.IdleTimeout * time.Second,
		ErrorLog:          serverLogger,
	}

	log.Info("GoPx API service is running on: %s [HTTP]", addr)
	err := server.ListenAndServe()
	log.Fatal("Error: %s", err) // err is always non-nill
}

func startHTTPS() {
	addr := httpsAddr()
	r := route.Router()
	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadTimeout:       config.Service.ReadTimeout * time.Second,
		ReadHeaderTimeout: config.Service.ReadTimeout * time.Second,
		WriteTimeout:      config.Service.WriteTimeout * time.Second,
		IdleTimeout:       config.Service.IdleTimeout * time.Second,
		ErrorLog:          serverLogger,
	}

	log.Info("GoPx API service is running on: %s [HTTPS]", addr)
	err := server.ListenAndServeTLS(config.Service.CertFile, config.Service.KeyFile)
	log.Fatal("Error: %s", err) // err is always non-nill
}

func httpAddr() string {
	return net.JoinHostPort(config.Service.Host, strconv.Itoa(config.Service.HTTPPort))
}

func httpsAddr() string {
	return net.JoinHostPort(config.Service.Host, strconv.Itoa(config.Service.HTTPSPort))
}
