package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hamburghammer/gsave/controller"
	"github.com/hamburghammer/gsave/db"
)

var servePort int

func init() {
	flag.IntVar(&servePort, "port", 8080, "The port for the HTTP server.")
	flag.Parse()
}

func main() {
	log.Println("Initializing the DB...")
	stats := []db.Stats{
		{Hostname: "foo", CPU: 0},
		{Hostname: "foo", CPU: 1},
		{Hostname: "bar", CPU: 0},
	}
	hostDB, err := initDB(stats)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Initializing the routes...")
	controllers := []controller.Router{
		controller.NewHostsRouter(hostDB),
	}
	router := initRouter(hostDB, controllers)

	log.Println("Starting the HTTP server...")
	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", servePort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
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
	log.Printf("The HTTP server is running: http://localhost:%d/hosts\n", servePort)
	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Println("Shutting down the server...")
			return
		}
		log.Fatalf("An unexpected error happend while running the HTTP server: %v\n", err)
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
		log.Printf("An error happened on the shutdown of the server: %v", err)
	}
}
