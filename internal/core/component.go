package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-argmapper"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint/internal/factory"
	"github.com/hashicorp/waypoint/internal/plugin"
)

// startPlugin starts a plugin with the given type and name. The returned
// value must be closed to clean up the plugin properly.
func (a *App) startPlugin(
	ctx context.Context,
	typ component.Type,
	f *factory.Factory,
	n string,
) (*plugin.Instance, error) {
	log := a.logger.Named(strings.ToLower(typ.String()))

	// Get the factory function for this type
	fn := f.Func(n)
	if fn == nil {
		return nil, fmt.Errorf("unknown type: %q", n)
	}

	// Call the factory to get our raw value (interface{} type)
	fnResult := fn.Call(argmapper.Typed(ctx, a.source, log))
	if err := fnResult.Err(); err != nil {
		return nil, err
	}
	log.Info("initialized component", "type", typ.String())
	raw := fnResult.Out(0)

	// If we have a plugin.Instance then we can extract other information
	// from this plugin. We accept pure factories too that don't return
	// this so we type-check here.
	pinst, ok := raw.(*plugin.Instance)
	if !ok {
		pinst = &plugin.Instance{
			Component: raw,
			Close:     func() {},
		}
	}

	return pinst, nil
}
