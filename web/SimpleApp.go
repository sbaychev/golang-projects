//package main
//
//import (
//	"context"
//	"flag"
//	"log"
//	"net/http"
//	"os"
//	"os/signal"
//	"time"
//)
//
//func indexHandler(w http.ResponseWriter, r *http.Request) {
//	w.Write([]byte("<h1>Hello World!</h1>"))
//}
//
//func main() {
//	port := os.Getenv("PORT")
//	if port == "" {
//		port = "3000"
//	}
//
//	mux := http.NewServeMux()
//
//	mux.HandleFunc("/", indexHandler)
//	http.ListenAndServe(":"+port, mux)
//}

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	listenAddr string
)

func main() {
	flag.StringVar(&listenAddr, "listen-addr", ":5000", "server listen address")
	flag.Parse()

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	server := newWebServer(logger)
	go gracefullShutdown(server, logger, quit, done)

	logger.Println("Server is ready to handle requests at ", listenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}

	<-done
	logger.Println("Server has stopped")
}

func gracefullShutdown(server *http.Server, logger *log.Logger, quit <-chan os.Signal, done chan<- bool) {
	<-quit
	logger.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	close(done)
}

func newWebServer(logger *log.Logger) *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return &http.Server{
		Addr:         listenAddr,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}
