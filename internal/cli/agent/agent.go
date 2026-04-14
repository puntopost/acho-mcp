package agent

import "fmt"

// Setup defines the interface for registering acho with an AI agent.
type Setup interface {
	Name() string
	Description() string
	Setup() error
	FormatContext(body string) string
	FormatRemember(body string) string
}

var registry []Setup

// Register adds an agent setup to the registry.
func Register(a Setup) {
	registry = append(registry, a)
}

// All returns all registered agent setups.
func All() []Setup {
	return registry
}

// Get returns the agent setup with the given name, or an error if not found.
func Get(name string) (Setup, error) {
	for _, a := range registry {
		if a.Name() == name {
			return a, nil
		}
	}
	return nil, fmt.Errorf("unknown agent: %s (available: %s)", name, Names())
}

// Names returns a comma-separated list of available agent names.
func Names() string {
	names := ""
	for i, a := range registry {
		if i > 0 {
			names += ", "
		}
		names += a.Name()
	}
	return names
}
