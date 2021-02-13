package main

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/pkg/errors"

	"github.com/josestg/webserver/app/api/handlers"
)

var build = "develop"
var namespace = "WEBSERVER"

func main() {
	logger := log.New(os.Stdout, "WEBSERVER", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if err := run(logger); err != nil {
		logger.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(logger *log.Logger) error {

	// ================================================================================
	// Configuration
	var cfg struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:8080"`
			DebugHost       string        `conf:"default:0.0.0.0:8181"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriterTimeout   time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
	}

	cfg.Version.SVN = build
	cfg.Version.Desc = "Copyright information here."

	if err := conf.Parse(os.Args[1:], namespace, &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage(namespace, &cfg)
			if err != nil {
				return errors.Wrap(err, "run: generating config usage.")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString(namespace, &cfg)
			if err != nil {
				return errors.Wrap(err, "run: generating config version.")
			}
			fmt.Println(version)
			return nil
		default:
			return errors.Wrap(err, "run: Could not parse configuration.")
		}
	}

	// ================================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	logger.Printf("run: Started: Application initializing: version %q.", build)
	defer logger.Println("run: Completed.")

	// Prints configuration to output.
	cfgString, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "run: Could not generating config string for output.")
	}
	logger.Printf("run: Config: \n%v\n", cfgString)

	// ================================================================================
	// Start API Service

	// Start Debug API
	// Not concerned with shutting this down when application is shutdown.
	logger.Println("run: Initializing debugging support.")
	go func() {
		logger.Println("run: Debug listening on %s.", cfg.Web.DebugHost)
		if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
			logger.Printf("run: Debug listener closed: %v", err)
		}
	}()

	logger.Println("run: Initializing API Service support.")
	// Make a channel to listen for an interrupt signal or terminate signal from the OS.
	// Use buffer channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Handler:      handlers.API(logger, shutdown),
		Addr:         cfg.Web.APIHost,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriterTimeout,
	}

	// Make a channel to listen for errors coming from the API services listener.
	// Use a buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the API services for listening the requests.
	go func() {
		logger.Printf("run: API services listening on %s.", cfg.Web.APIHost)
		serverErrors <- api.ListenAndServe()
	}()

	// Blocking run and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "run: Server error.")
	case sig := <-shutdown:
		logger.Printf("run: %v start shutdown.", sig)

		// Give outstanding request a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close() // Hard shutdown.
			return errors.Wrap(err, "run: Could not stop the server gracefully.")
		}
	}

	return nil
}
