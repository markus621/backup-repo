package main

import (
	"github.com/deweppro/goppy"
	"github.com/deweppro/goppy/plugins"
	"github.com/deweppro/goppy/plugins/web"
	"github.com/markus621/backup-repo/internal"
)

func main() {
	app := goppy.New()
	app.WithConfig("./config.yaml") // Reassigned via the `--config` argument when run via the console.
	app.Plugins(
		web.WithHTTPClient(),
	)
	app.Plugins(
		plugins.Plugin{
			Config: &internal.Config{},
			Inject: internal.NewBackup,
		},
	)
	app.Run()
}
