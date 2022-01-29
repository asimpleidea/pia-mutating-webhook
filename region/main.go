package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	defaultWorkersNumber  uint          = 5
	defaultMaxServers     uint          = 25
	defaultServersListURL string        = "https://serverlist.piaservers.net/vpninfo/servers/v6"
	orderByName           string        = "name"
	orderByLatency        string        = "latency"
	defaultOrderBy        string        = orderByName
	ascendingOrder        string        = "asc"
	descendingOrder       string        = "desc"
	defaultOrderDirection string        = ascendingOrder
	defaultVerbosity      int           = 1
	defaultMaxLatency     time.Duration = 50 * time.Millisecond
)

type Options struct {
	MaxLatency     time.Duration
	Workers        uint
	MaxServers     uint
	ServersListURL string
	OrderBy        string
	OrderDirection string
	Verbosity      int
}

func main() {
	// -----------------------------------
	// Flags and inits
	// -----------------------------------
	opts := &Options{}

	flag.DurationVar(&opts.MaxLatency, "max-latency", defaultMaxLatency,
		"Maximum latency tolerated for a server to be kept.")
	flag.UintVar(&opts.Workers, "workers", defaultWorkersNumber,
		"Number of concurrent workers to use for checking latency.")
	flag.UintVar(&opts.MaxServers, "max-regions", defaultMaxServers,
		"Maximum number of servers to keep.")
	flag.StringVar(&opts.ServersListURL, "servers-list-url", defaultServersListURL,
		"The URL where to get the list of servers.")
	flag.StringVar(&opts.OrderBy, "order-by", defaultOrderBy,
		fmt.Sprintf("How to order the the servers list. Accepted values: %s or %s.", orderByName, orderByLatency))
	flag.StringVar(&opts.OrderDirection, "order-direction", defaultOrderDirection,
		fmt.Sprintf("The order direction. Accepted values: %s or %s", ascendingOrder, descendingOrder))
	flag.IntVar(&opts.Verbosity, "verbosity", defaultVerbosity,
		"The log verbosity level, from 0 (verbose) to 3 (silent).")
	flag.Parse()

	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	log.Info().Msg("starting...")

	// -----------------------------------
	// Validations
	// -----------------------------------
	{
		logLevels := []zerolog.Level{
			zerolog.DebugLevel,
			zerolog.InfoLevel,
			zerolog.ErrorLevel,
			zerolog.FatalLevel,
		}
		if opts.Verbosity < 0 || opts.Verbosity > len(logLevels)-1 {
			log.Fatal().Err(fmt.Errorf("invalid verbosity level")).Msg("")
		}
		log = log.Level(logLevels[opts.Verbosity])
	}

	if opts.MaxLatency == 0 {
		log.Fatal().Err(fmt.Errorf("invalid max latency provided")).
			Dur("max-latency", opts.MaxLatency).Msg("")
	}

	if opts.Workers == 0 {
		log.Debug().Uint("workers", opts.Workers).
			Uint("default-workers-number", defaultWorkersNumber).
			Msg("invalid workers flag provided: resetting...")
	}

	if opts.MaxServers == 0 {
		log.Debug().Msg("using no limits for maximum servers to list")
	}

	if _, err := url.Parse(opts.ServersListURL); err != nil {
		log.Fatal().Err(err).Str("servers-list-url", opts.ServersListURL).
			Msg("invalid servers list url provided")
	}

	if !strings.EqualFold(opts.OrderBy, orderByName) &&
		!strings.EqualFold(opts.OrderBy, orderByLatency) {
		log.Fatal().Err(fmt.Errorf("unknown order type")).
			Str("order-by", opts.OrderBy).Msg("invalid order-by flag provided")
	}

	if !strings.EqualFold(opts.OrderDirection, ascendingOrder) &&
		!strings.EqualFold(opts.OrderDirection, descendingOrder) {
		log.Fatal().Err(fmt.Errorf("unknown order direction")).
			Str("order-direction", opts.OrderDirection).
			Msg("invalid order-direction flag provided")
	}
}
