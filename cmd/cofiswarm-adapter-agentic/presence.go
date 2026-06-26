package main

import (
	"log"
	"os"

	"github.com/keepdevops/cofiswarm-adapter-agentic/internal/bus"
	"github.com/keepdevops/cofiswarm-observer-sdk/pkg/buspresence"
	"github.com/keepdevops/cofiswarm-observer-sdk/pkg/servicecomponent"
)

// presence bundles the adapter's optional observer-bus integrations so main() can start
// and stop them as a unit. Both are default-off and enabled by their own env var:
//   - COFISWARM_NATS_URL  -> announce presence + serve info/health via the NATS service component
//   - COFISWARM_BRIDGE_URL -> broker-free carrier presence in the observer live roster
type presence struct {
	comp      *servicecomponent.Component
	stopBus   func() // carrier presence stop
	closeConn func() // nats connection close
}

// startPresence wires whichever integrations are enabled by the environment. Connection
// failures are logged and skipped so the adapter still serves its HTTP API without presence.
func startPresence(id string) *presence {
	p := &presence{}

	if url := os.Getenv("COFISWARM_NATS_URL"); url != "" {
		nc, err := servicecomponent.Connect(url, "cofiswarm-adapter-"+adapterName)
		if err != nil {
			log.Printf("bus connect %s: %v (running without presence)", url, err)
		} else {
			comp := servicecomponent.New(nc, id, id, bus.Routes(adapterName))
			if err := comp.Start(); err != nil {
				log.Printf("bus start: %v (running without presence)", err)
				nc.Close()
			} else {
				log.Printf("adapter-%s announcing presence via %s", adapterName, url)
				p.comp = comp
				p.closeConn = nc.Close
			}
		}
	}

	if bridge := os.Getenv("COFISWARM_BRIDGE_URL"); bridge != "" {
		p.stopBus = buspresence.StartPresence(bridge, id, map[string]any{"name": id})
	}

	return p
}

// Shutdown says goodbye (flipping offline now, not after the TTL) and releases resources.
func (p *presence) Shutdown() {
	if p.comp != nil {
		p.comp.Shutdown() // goodbye -> offline
	}
	if p.stopBus != nil {
		p.stopBus()
	}
	if p.closeConn != nil {
		p.closeConn()
	}
}
