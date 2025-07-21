package templates

import (
	"github.com/acheevo/template-engine/internal/core"
)

// init registers all template types at startup
func init() {
	// Register frontend template
	core.RegisterTemplate(&FrontendTemplate{})

	// Register Go API template
	core.RegisterTemplate(&GoAPITemplate{})

	// Future template types will be registered here:
	// core.RegisterTemplate(&MobileTemplate{})
}
