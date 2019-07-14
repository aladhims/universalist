package main

import (
	"flag"
	"log"
	"os"
	"text/tabwriter"

	"github.com/aladhims/universalist"
)

var (
	path       string
	configPath string
	indent     bool
)

func init() {
	flag.StringVar(&path, "path", "./", "the path of the workdir that will be scanned")
	flag.StringVar(&configPath, "config", "", "config path for overwriting default config values")
	flag.BoolVar(&indent, "indent", false, "The output will be indented if the flag is specified")
}

func main() {
	flag.Parse()

	if path == "" {
		path = "./"
	}

	var options []universalist.Option

	options = append(options, universalist.WithPath(path))

	if indent {
		options = append(options, universalist.WithWriter(tabwriter.NewWriter(os.Stdout, 5, 1, 10, '\t', tabwriter.AlignRight)))
	}

	ul, err := universalist.New(configPath, options...)
	if err != nil {
		log.Fatalf("error: %s", err)
	}

	err = ul.Start()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}
