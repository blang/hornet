package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		listen    = flag.String("listen", "127.0.0.1:7000", "REST API, e.g. 127.0.0.1:7000")
		storeFile = flag.String("store", "store.gob", "Store file")
	)
	flag.Parse()

	store, err := RestoreStore(*storeFile)
	if err != nil {
		log.Printf("Could not read store from file %q, error: %q", *storeFile, err)
	}

	service := NewService(store)
	restapi := NewRestAPI(service)
	go func() {
		log.Printf("Start RestAPI listening on %q", *listen)
		if err := http.ListenAndServe(*listen, restapi); err != nil {
			log.Fatalf("HTTP Server crashed: %q", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	s := <-c
	log.Println("Received signal %q, shut down gracefully", s)

	err = store.SaveStore(*storeFile)
	if err != nil {
		log.Printf("Could not save store to file %q: %q", *storeFile, err)
	}

	// Shutdown service: All instances
	service.Shutdown()

	log.Printf("Graceful shutdown complete")
}
