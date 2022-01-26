package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
)

type AppOptions struct {
	SidecarImage string
}

const (
	CodeNoError int = iota
	CodeNoSidecarImage
)

func main() {
	opts := &AppOptions{}

	flag.StringVar(&opts.SidecarImage, "sidecar-image", "",
		"Image to inject as a sidecar")
	flag.Parse()

	os.Exit(run(opts))
}

func run(opts *AppOptions) int {
	log := zerolog.New(os.Stderr)
	log.Info().Msg("starting...")

	if opts.SidecarImage == "" {
		log.Error().Msg("no sidecar image provided")
		return CodeNoSidecarImage
	}

	return CodeNoError
}
