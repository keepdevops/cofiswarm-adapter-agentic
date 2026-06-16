package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/keepdevops/cofiswarm-adapter-agentic/internal/httpapi"
)

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
	log.Printf("adapter-agentic on %s", srv.Addr())
	log.Fatal(http.ListenAndServe(srv.Addr(), srv.Handler()))
}
