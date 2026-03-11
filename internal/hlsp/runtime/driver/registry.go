package driver

import (
	"github.com/hephbuild/heph/internal/engine"
	"github.com/hephbuild/heph/lib/pluginsdk"
)

// Registry wraps the heph engine's driver map for use by the LSP runtime.
type Registry struct {
	engine *engine.Engine
}

func NewRegistry(e *engine.Engine) *Registry {
	return &Registry{engine: e}
}

func (r *Registry) GetDriver(name string) (pluginsdk.Driver, bool) {
	drv, ok := r.engine.DriversByName[name]
	return drv, ok
}
