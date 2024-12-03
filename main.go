package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/containerd/plugin"
	"github.com/containerd/plugin/registry"

	// Load all plugins
	_ "github.com/dmcgowan/plugin-example/plugins/fileserver"
	_ "github.com/dmcgowan/plugin-example/plugins/fs"
)

type config struct {
	Address string                      `json:"address"`
	Plugins map[string]*json.RawMessage `json:"plugins"`
}

var configFlag = flag.String("config", "config.json", "config file")

func main() {
	flag.Parse()
	ctx := context.Background()

	cfg := config{
		Address: ":3333",
	}
	if err := readConfig(*configFlag, &cfg); err != nil {
		log.Fatal("Failed to read config:", err)
	}

	// Set of initialized plugins
	initialized := plugin.NewPluginSet()

	// Global properties used by plugins
	// Examples: root directories, service name, address
	properties := map[string]string{}

	// Filter option to avoid computing graph for disabled plugins
	filter := func(*plugin.Registration) bool {
		return false
	}

	for _, reg := range registry.Graph(filter) {
		id := reg.URI()

		// Create the init context which is passed to the plugin init
		ic := plugin.NewContext(ctx, initialized, properties)

		// `reg.Config` will be the default configuration for the plugin
		// `ic.Config` is what is passed to the plugin
		ic.Config = reg.Config
		if pluginCfg := cfg.Plugins[id]; pluginCfg != nil {
			// In this example, the plugin's JSON config
			// will be marshalled back into JSON then
			// unmarshalled into the config provided by the plugin
			b, err := pluginCfg.MarshalJSON()
			if err != nil {
				log.Fatal("Failed to marshal plugin %q config: %v", id, err)
			}
			if err := json.Unmarshal(b, ic.Config); err != nil {
				log.Fatal("Failed to unmarshal plugin %q config: %v", id, err)
			}

			// Here is where you can use your own configuration
			// logic to configure the plugin.
			// `ic.Config` is the config object used for init
			ic.Config = reg.Config
		}

		p := reg.Init(ic)

		// Adds to the initialized set so future plugins can use
		initialized.Add(p)

		// Instance could be retrieved here to handle plugin errors
		// immediately or build a list of a specific type of plugin,
		// for example, all plugins which register a GRPC service.
		instance, err := p.Instance()
		if err != nil {
			log.Fatalf("Plugin %s failed to load: %v", id, err)
		}

		if handler, ok := instance.(http.Handler); ok && reg.Type == "plugin-example.http" {
			http.Handle("/", handler)
		}
	}

	if err := http.ListenAndServe(cfg.Address, nil); err != nil {
		log.Fatal("ListenAndServer error:", err)
	}
}

func readConfig(f string, cfg any) error {
	file, err := os.Open(f)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(cfg)
}
