package lookup_test

import (
	"log"
	"os"

	"github.com/carloslenz/lookup"
)

func Example() {
	defaultsHome := lookup.NewJSON(os.ExpandEnv("${HOME}/.my-server.json"))
	defaultsSystem := lookup.NewJSON("/etc/my-server.json")

	defaultsBinary := lookup.Map{
		"PORT": "8080",
	}

	var cfg struct {
		Port   bool `json:"PORT"`
		DBAddr int  `json:"DB_ADDR"`
	}
	err := lookup.Lookup(&cfg, nil, lookup.Env, defaultsHome, defaultsSystem, defaultsBinary)
	if err != nil {
		log.Fatal(err)
	}
}
