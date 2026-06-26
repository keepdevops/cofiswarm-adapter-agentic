package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keepdevops/cofiswarm-adapter-agentic/internal/httpapi"
)

const adapterName = "agentic"

func main() {
	cfgPath := flag.String("config", "", "adapter yaml")
	flag.Parse()
	if *cfgPath == "" {
		*cfgPath = "/etc/cofiswarm/adapter-agentic/adapter-agentic.yaml"
		if v := os.Getenv("COFISWARM_ADAPTER_CONFIG"); v != "" {
			*cfgPath = v
		}
	}
	cfg, err := httpapi.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	srv := httpapi.New(cfg)

	// Optional observer-bus presence (NATS + broker-free carrier), default-off via env.
	pres := startPresence("adapter-" + adapterName)

	httpSrv := &http.Server{
		Addr:              srv.Addr(),
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	go func() {
		log.Printf("adapter-%s on %s", adapterName, srv.Addr())
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("adapter-%s: server error: %v", adapterName, err)
		}
	}()

	// On SIGINT/SIGTERM: say goodbye (flip offline now, not after the TTL) then drain HTTP.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Printf("adapter-%s: shutting down", adapterName)
	pres.Shutdown()
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		log.Printf("adapter-%s: graceful shutdown: %v", adapterName, err)
	}
}
