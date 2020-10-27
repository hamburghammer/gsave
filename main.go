package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hamburghammer/gsave/controller"
	"github.com/hamburghammer/gsave/db"
	log "github.com/sirupsen/logrus"
)

var (
	servePort  int
	logPackage = log.WithField("Package", "main")
)

func init() {
	flag.IntVar(&servePort, "port", 8080, "The port for the HTTP server.")
	verbose := flag.Bool("verbose", false, "Enable debug logging output.")
	quiet := flag.Bool("quiet", false, "Disable loging output only prints errors.")
	jsonLogging := flag.Bool("json", false, "Set the logging format to json.")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	if *quiet {
		log.SetLevel(log.ErrorLevel)
	}
	if *jsonLogging {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func main() {
	logPackage.Info("Initializing the DB...")
	stats := []db.Stats{
		{Hostname: "foo", CPU: 0},
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
