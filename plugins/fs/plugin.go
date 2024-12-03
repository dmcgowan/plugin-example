package plugin

import (
	"os"

	"github.com/containerd/plugin"
	"github.com/containerd/plugin/registry"
)

func init() {
	type config struct {
		Public string `json:"public"`
	}

	registry.Register(&plugin.Registration{
		Type: "plugin-example.fs",
		ID:   "local",
		Config: &config{
			Public: "/var/www/plugin-example-default",
		},
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			publicDir := ic.Config.(*config).Public
			ic.Meta.Exports["public"] = publicDir
			return os.DirFS(publicDir), nil
		},
	})
}
