package main

import (
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type AppOptions struct {
	SidecarImage string
	DebugMode    bool
}

const (
	fiberAppName string = "PIA Mutating Webhook"
)

const (
	CodeNoError int = iota
	CodeNoSidecarImage
)

func main() {
	opts := &AppOptions{}

	flag.StringVar(&opts.SidecarImage, "sidecar-image", "",
		"Image to inject as a sidecar")
	flag.BoolVar(&opts.DebugMode, "debug", false,
		"Whether to show debug log lines")
	flag.Parse()

	os.Exit(run(opts))
}

func run(opts *AppOptions) int {
	log := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
	log.Info().Msg("starting...")

	// -----------------------------
	// Parse options
	// -----------------------------

	if opts.SidecarImage == "" {
		log.Error().Msg("no sidecar image provided")
		return CodeNoSidecarImage
	}

	if opts.DebugMode {
		log = log.Level(zerolog.DebugLevel)
	}

	// -----------------------------
	// Server and paths
	// -----------------------------

	app := fiber.New(fiber.Config{
		AppName:               fiberAppName,
		ReadTimeout:           time.Minute,
		DisableStartupMessage: opts.DebugMode,
	})

	app.Get("/readyz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Err(err).Msg("error while starting server")
		}
	}()

	// -----------------------------
	// Graceful shutdown & clean ups
	// -----------------------------

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Info().Msg("shutting down...")
	if err := app.Shutdown(); err != nil {
		log.Err(err).Msg("error while waiting for server to shutdown")
	}
	log.Info().Msg("goodbye!")

	return CodeNoError
}
