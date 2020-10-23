package config

import (
	"github.com/hashicorp/hcl/v2"
)

type hclStage struct {
	Body hcl.Body `hcl:",remain"`
}

type hclBuild struct {
	Registry *hclStage `hcl:"registry,block"`
	Body     hcl.Body  `hcl:",remain"`
}

// Build are the build settings.
type Build struct {
	Labels map[string]string `hcl:"labels,optional"`
	Hooks  []*Hook           `hcl:"hook,block"`
	Use    *Use              `hcl:"use,block"`

	// This should not be used directly. This is here for validation.
	// Instead, use App.Registry().
	Registry *Registry `hcl:"registry,block"`
}

// Registry are the registry settings.
type Registry struct {
	Labels map[string]string `hcl:"labels,optional"`
	Hooks  []*Hook           `hcl:"hook,block"`
	Use    *Use              `hcl:"use,block"`
}

// Deploy are the deploy settings.
type Deploy struct {
	Labels map[string]string `hcl:"labels,optional"`
	Hooks  []*Hook           `hcl:"hook,block"`
	Use    *Use              `hcl:"use,block"`
}

// Release are the release settings.
type Release struct {
	Labels map[string]string `hcl:"labels,optional"`
	Hooks  []*Hook           `hcl:"hook,block"`
	Use    *Use              `hcl:"use,block"`
}

// Hook is the configuration for a hook that runs at specified times.
type Hook struct {
	When      string   `hcl:"when,attr"`
	Command   []string `hcl:"command,attr"`
	OnFailure string   `hcl:"on_failure,optional"`
}

func (h *Hook) ContinueOnFailure() bool {
	return h.OnFailure == "continue"
}
