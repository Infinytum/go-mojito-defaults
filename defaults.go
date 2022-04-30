package defaults

import "github.com/infinytum/go-mojito"

func init() {
	mojito.Register(func() mojito.Router {
		return newBunRouter()
	}, true)

	_ = mojito.Register(func() mojito.Renderer {
		return newHandlebarsRenderer()
	}, true)

	_ = mojito.Register(func() mojito.Logger {
		return newZerologLogger()
	}, true)
}
