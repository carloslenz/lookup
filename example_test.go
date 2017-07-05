package lookup_test

import (
	"log"
	"os"
	"path"

	"github.com/carloslenz/lookup"
)

func Example() {
	// This snippet tries to lookup from (in order of precedence):
	// - command line
	// - environment
	// - json file at home dir
	// - json file at system config dir
	// - defaults embedded into the binary
	args := lookup.NewArgs("-", os.Args)

	const cfgName = "my-server.json"
	defaultsHome := lookup.NewJSONFile(path.Join(os.ExpandEnv("${HOME}"), cfgName))
	defaultsSystem := lookup.NewJSONFile(path.Join("/etc", cfgName))

	defaultsBinary := lookup.Map{
		"PORT": "8080",
	}

	var cfg struct {
		Port   bool `json:"PORT"`
		DBAddr int  `json:"DB_ADDR"`
	}
	err := lookup.Lookup(&cfg, nil, args, lookup.Env, defaultsHome, defaultsSystem, defaultsBinary)
	if err != nil {
		log.Fatal(err)
	}
	os.Args = args.ExtraArgs()
	// Your server here.
}
