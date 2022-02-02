package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	defaultWorkersNumber     uint          = 5
	defaultMaxServers        uint          = 25
	defaultServersListURL    string        = "https://serverlist.piaservers.net/vpninfo/servers/v6"
	orderByName              string        = "name"
	orderByLatency           string        = "latency"
	defaultOrderBy           string        = orderByName
	ascendingOrder           string        = "asc"
	descendingOrder          string        = "desc"
	defaultOrderDirection    string        = ascendingOrder
	defaultVerbosity         int           = 1
	defaultMaxLatency        time.Duration = 50 * time.Millisecond
	defaultFrequency         time.Duration = time.Hour
	defaultResultsWriterFreq time.Duration = 5 * time.Minute
)

type Options struct {
	MaxLatency     time.Duration
	Workers        uint
	MaxServers     uint
	ServersListURL string
	OrderBy        string
	OrderDirection string
	Verbosity      int
	Frequency      time.Duration
}

func main() {
	opts := &Options{}

	// -----------------------------------
	// Flags
	// -----------------------------------

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
	flag.DurationVar(&opts.Frequency, "frequency", defaultFrequency,
		"The frequency of updating the list of servers.")
	flag.Parse()

	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	log.Info().Msg("starting...")

	// -----------------------------------
	// Get Kubernetes clientset
	// -----------------------------------

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Err(err).Msg("could not get Kubernetes clientset")
		return
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		log.Error().Msg("could not get namespace")
		return
	}

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
	// Start workers
	// -----------------------------------

	ctx, canc := context.WithCancel(context.Background())
	regionsChan := make(chan *Region, 256)
	latenciesChan := make(chan *Region, 256)
	wg := sync.WaitGroup{}
	for i := 0; i < int(opts.Workers); i++ {
		wg.Add(1)
		go func(wid int) {
			defer wg.Done()
			log.Info().Int("worker", wid+1).Msg("worker starting...")
			work(ctx, regionsChan, latenciesChan, log, opts.MaxLatency)
			log.Info().Int("worker", wid+1).Msg("worker exited")
		}(i)
	}

	// -----------------------------------
	// Handle events
	// -----------------------------------

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)

	updateTicker := time.NewTicker(opts.Frequency)
	confWriterTimer := time.NewTimer(time.Second)

	// This will be used to trigger the first iteration
	firstTime := time.NewTimer(5 * time.Second)

	latResults := []*Region{}
	stopping := false
	for !stopping {
		select {
		case <-updateTicker.C:
		case <-firstTime.C:
			wg.Add(1)
			go func() {
				defer wg.Done()

				servListCtx, servListCanc := context.WithTimeout(ctx, time.Minute)
				defer servListCanc()

				log.Debug().Msg("getting list of servers...")
				regions, err := getServersList(servListCtx, opts.ServersListURL)
				if err != nil {
					// TODO: auto-exit if failed too many times in a row?
					log.Err(err).Msg("could not load regions, skipping...")
					return
				}

				log.Info().Msg("calculating latencies...")

				for _, region := range regions {
					regionsChan <- region
				}
			}()

			// After some minutes, this will activate and will write results
			confWriterTimer = time.NewTimer(time.Minute)

		case <-confWriterTimer.C:
			_ = clientset
		case lat := <-latenciesChan:
			if lat != nil && len(lat.Servers.WireGuard) > 0 {
				latResults = append(latResults, lat)
			}
		case <-stop:
			stopping = true
			updateTicker.Stop()
			confWriterTimer.Stop()
			fmt.Println()
		}
	}

	close(latenciesChan)
	close(regionsChan)
	canc()
	log.Info().Msg("shutting down...")
	log.Info().Msg("waiting for all goroutines to exit...")

	wg.Wait()
	log.Info().Msg("goodbye!")
}

func getServersList(ctx context.Context, serversListURL string) ([]*Region, error) {
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serversListURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listResp ServersListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}

	return listResp.Regions, nil
}

func work(ctx context.Context, regionsChan, latenciesResult chan *Region, log zerolog.Logger, maxLatency time.Duration) {
	for reg := range regionsChan {
		if reg.Servers == nil {
			continue
		}

		if len(reg.Servers.WireGuard) == 0 {
			// TODO: we're only concentrating on WireGuard for now. So we skip
			// this if it doesn't have any.
			continue
		}

		ips := []*Server{}
		for _, serv := range reg.Servers.WireGuard {
			ip := fmt.Sprintf("%s:443", serv.IP)
			l := log.With().Str("cn", serv.CN).Str("ip", serv.IP).
				Logger()

			now := time.Now()

			conn, err := net.DialTimeout("tcp", ip, maxLatency)
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					l.Debug().Msg("ignoring, as latency is too high")

				} else {
					l.Err(err).Msg("error while connecting to server, skipping...")
				}
				continue
			}

			elapsed := time.Since(now)
			conn.Close()

			l.Debug().Str("latency", elapsed.String()).Msg("connected and retrieved latency")
			ips = append(ips, &Server{IP: serv.IP, CN: serv.CN, VAN: serv.VAN, Latency: &elapsed})
		}

		reg := reg.Clone()
		reg.Servers.WireGuard = ips
		latenciesResult <- reg
	}
}
