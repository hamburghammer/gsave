package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hamburghammer/gsave/controller"
	"github.com/hamburghammer/gsave/controller/middleware"
	"github.com/hamburghammer/gsave/db"
	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

var (
	servePort  int
	logPackage = log.WithField("Package", "main")
)

type arguments struct {
	Port        int    `short:"p" long:"port" default:"8080" description:"The port for the HTTP server." env:"GSAVE_PORT"`
	Token       string `short:"t" long:"token" required:"yes" description:"The token for the authentication through HTTP." env:"GSAVE_TOKEN"`
	Verbose     bool   `short:"v" long:"verbose" description:"Enable trace logging level output."`
	Quiet       bool   `short:"q" long:"quiet" description:"Disable standard logging output and only prints errors."`
	JSONLogging bool   `long:"json" description:"Set the logging format to json."`
}

func init() {
	args := arguments{}
	_, err := flags.Parse(&args)
	if err != nil {
		if _, ok := err.(*flags.Error); ok {
			os.Exit(1)
			return
		}
		logPackage.Fatal(err)
	}

	servePort = args.Port

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if args.Verbose {
		log.SetLevel(log.TraceLevel)
	}
	if args.Quiet {
		log.SetLevel(log.ErrorLevel)
	}
	if args.JSONLogging {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func main() {
	logPackage.Info("Initializing the DB...")
	stats := []db.Stats{
		{Hostname: "foo", CPU: 0, Disk: db.Memory{Total: 10, Used: 5}, Mem: db.Memory{Total: 20, Used: 10}, Processes: []db.Process{{Name: "foo", Pid: 1, CPU: 0.5}}},
		{Hostname: "foo", CPU: 1},
		{Hostname: "bar", CPU: 0},
	}
	hostDB, err := initDB(stats)
	if err != nil {
		logPackage.Fatal(err)
	}

	logPackage.Info("Initializing the routes...")
	controllers := []controller.Router{
		controller.NewHostsRouter(hostDB),
	}
	router := initRouter(hostDB, controllers)

	auth := middleware.NewAuthMiddleware([]string{"foo"})
	// Add default middlewares
	router.Use(middleware.RequestTimeLoggingHandler)
	router.Use(middleware.PanicRecoverHandler)
	router.Use(auth.AuthHandler)

	logPackage.Info("Starting the HTTP server...")
	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", servePort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go startHTTPServer(server, &wg)
	go listenToStopHTTPServer(server, &wg)

	wg.Wait()
}

func initDB(stats []db.Stats) (db.HostDB, error) {
	hostDB := db.NewInMemoryDB()
	for _, stat := range stats {
		if err := hostDB.InsertStats(stat.Hostname, stat); err != nil {
			return hostDB, fmt.Errorf("Init db throwed an error: %w", err)
		}
	}

	return hostDB, nil
}

func initRouter(hostDB db.HostDB, controllers []controller.Router) *mux.Router {
	router := mux.NewRouter()
	for _, controller := range controllers {
		subrouter := router.PathPrefix(controller.GetPrefix()).Name(controller.GetRouteName()).Subrouter()
		controller.Register(subrouter)
	}

	return router
}

func startHTTPServer(server *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	logPackage.Infof("The HTTP server is running: http://localhost:%d/hosts\n", servePort)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logPackage.Info("Shutting down the server...")
			return
		}
		logPackage.Fatalf("An unexpected error happend while running the HTTP server: %v\n", err)
	}
}

func listenToStopHTTPServer(server *http.Server, wg *sync.WaitGroup) {
	defer wg.Done()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logPackage.Errorf("An error happened on the shutdown of the server: %v", err)
	}
}
