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

	"github.com/keepdevops/cofiswarm-adapter-agentic/internal/bus"
	"github.com/keepdevops/cofiswarm-adapter-agentic/internal/httpapi"
	"github.com/keepdevops/cofiswarm-observer-sdk/pkg/buspresence"
	"github.com/keepdevops/cofiswarm-observer-sdk/pkg/servicecomponent"
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

	ID := "adapter-" + adapterName

	// Optional: announce presence on the observer bus alongside the HTTP API (default-off).
	// COFISWARM_NATS_URL=nats://host:4222 enables it.
	var comp *servicecomponent.Component
	if url := os.Getenv("COFISWARM_NATS_URL"); url != "" {
		nc, cErr := servicecomponent.Connect(url, "cofiswarm-adapter-"+adapterName)
		if cErr != nil {
			log.Printf("bus connect %s: %v (running without presence)", url, cErr)
		} else {
			defer nc.Close()
			comp = servicecomponent.New(nc, ID, ID, bus.Routes(adapterName))
			if sErr := comp.Start(); sErr != nil {
				log.Printf("bus start: %v (running without presence)", sErr)
				comp = nil
			} else {
				log.Printf("adapter-%s announcing presence via %s", adapterName, url)
			}
		}
	}

	// Carrier presence (broker-free, default-off via COFISWARM_BRIDGE_URL): appear in the
	// observer live roster over the zmq-bridge without needing a NATS broker.
	var stopPresence func()
	if bridge := os.Getenv("COFISWARM_BRIDGE_URL"); bridge != "" {
		stopPresence = buspresence.StartPresence(bridge, ID, map[string]any{"name": ID})
	}

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
	if comp != nil {
		comp.Shutdown() // goodbye -> offline
	}
	if stopPresence != nil {
		stopPresence()
	}
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutCtx); err != nil {
		log.Printf("adapter-%s: graceful shutdown: %v", adapterName, err)
	}
}
