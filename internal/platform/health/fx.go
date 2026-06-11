package health

import "go.uber.org/fx"

var FxModule = fx.Module(
	"platform-health",
	fx.Provide(NewRegistry),
)
