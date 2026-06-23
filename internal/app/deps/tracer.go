package deps

import (
	"github.com/assanoff/service-kit-x/internal/app/provider"
	"github.com/assanoff/servicekit/dim"
)

// initTracer initializes the application tracer (no-op when tracing is disabled).
var initTracer = func(c *Deps) (cleanup dim.CleanupFunc, err error) {
	c.Tracer, cleanup = dim.NewResource("Tracer", provider.Tracer(&c.Opts))
	return cleanup, nil
}
