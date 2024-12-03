package fileserver

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/containerd/plugin"
	"github.com/containerd/plugin/registry"
)

func init() {
	registry.Register(&plugin.Registration{
		Type: "plugin-example.http",
		ID:   "fileserver",
		Requires: []plugin.Type{
			"plugin-example.fs",
		},
		InitFn: func(ic *plugin.InitContext) (interface{}, error) {
			inst, err := ic.GetSingle("plugin-example.fs")
			if err != nil {
				return nil, err
			}

			f, ok := inst.(fs.FS)
			if !ok {
				return nil, fmt.Errorf("unknown fs type: %T", inst)
			}

			return http.FileServerFS(f), nil
		},
	})
}
