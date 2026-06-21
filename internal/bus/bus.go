// Package bus wires the adapter onto the NATS observer bus via the shared cofiswarm-observer-sdk
// service component: it announces presence and serves .adapter.<name>.{info,health} alongside
// the adapter's HTTP API.
package bus

import (
	"github.com/keepdevops/cofiswarm-observer-sdk/pkg/servicecomponent"
)

// Routes wires the adapter's capability subjects. Reply field names carry schema_version for
// the major-version gate, mirroring the other service components.
func Routes(name string) map[string]servicecomponent.Handler {
	return map[string]servicecomponent.Handler{
		servicecomponent.Prefix + ".adapter." + name + ".info":   infoHandler(name),
		servicecomponent.Prefix + ".adapter." + name + ".health": healthHandler(),
	}
}

func infoHandler(name string) servicecomponent.Handler {
	return func([]byte) (any, error) {
		return infoReply{SchemaVersion: servicecomponent.SchemaVersion, OK: true, Adapter: name}, nil
	}
}

func healthHandler() servicecomponent.Handler {
	return func([]byte) (any, error) {
		return healthReply{SchemaVersion: servicecomponent.SchemaVersion, OK: true, Status: "ok"}, nil
	}
}

type infoReply struct {
	SchemaVersion string `json:"schema_version"`
	OK            bool   `json:"ok"`
	Error         string `json:"error,omitempty"`
	Adapter       string `json:"adapter"`
}

type healthReply struct {
	SchemaVersion string `json:"schema_version"`
	OK            bool   `json:"ok"`
	Error         string `json:"error,omitempty"`
	Status        string `json:"status"`
}
