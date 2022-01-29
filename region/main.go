package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	ctx, canc := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
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

	// -----------------------------------
	// Get the list
	// -----------------------------------

	var regions []*Region
	{
		servListCtx, servListCanc := context.WithTimeout(ctx, time.Minute)
		resp := getServersList(servListCtx, log, opts.ServersListURL)
		select {
		case <-stop:
			servListCanc()
			fmt.Println()
			log.Info().Msg("shutting down...")
			<-resp
			return
		case resp := <-resp:
			if resp.Err != nil {
				log.Fatal().Err(resp.Err).Msg("could not get list of servers")
			}

			regions = resp.Response.Regions
		}
	}

	// TODO: use regions
	_ = regions

	canc()
}

type ServersListResult struct {
	Err      error
	Response *ServersListResponse
}

func getServersList(ctx context.Context, log zerolog.Logger, serversListURL string) <-chan ServersListResult {
	result := make(chan ServersListResult)
	log.Info().Msg("getting list of servers...")

	go func() {
		client := http.Client{}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, serversListURL, nil)
		if err != nil {
			result <- ServersListResult{Err: err, Response: nil}
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			result <- ServersListResult{Err: err, Response: nil}
			return
		}
		defer resp.Body.Close()
		log.Info().Msg("done")

		var listResp ServersListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
			result <- ServersListResult{Err: err, Response: nil}
			return
		}

		result <- ServersListResult{Err: nil, Response: &listResp}
	}()

	return result
}
