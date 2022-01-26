package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
)

type AppOptions struct {
	SidecarImage string
}

func main() {
	opts := &AppOptions{}

	flag.StringVar(&opts.SidecarImage, "sidecar-image", "",
		"Image to inject as a sidecar")
	flag.Parse()
}

func run(opts *AppOptions) int {
	log := zerolog.New(os.Stderr)
	log.Info().Msg("starting...")

	return 0
}
