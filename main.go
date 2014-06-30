package main

import (
	"flag"
	"github.com/cratonica/trayhost"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
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

	trayCloseCh := make(chan bool)
	tray(trayCloseCh)

	// Block until a signal is received.
	select {
	case s := <-c:
		log.Println("Received signal %q, shut down gracefully", s)
	case <-trayCloseCh:
	}

	err = store.SaveStore(*storeFile)
	if err != nil {
		log.Printf("Could not save store to file %q: %q", *storeFile, err)
	}

	// Shutdown service: All instances
	service.Shutdown()

	log.Printf("Graceful shutdown complete")
}

func tray(ch chan bool) {
	runtime.LockOSThread()

	go func() {
		// Be sure to call this to link the tray icon to the target url
		trayhost.SetUrl("http://hornet.echo12.de")
	}()

	// Enter the host system's event loop
	trayhost.EnterLoop("Hornet", iconData)

	// This is only reached once the user chooses the Exit menu item
	log.Println("Tray Exiting")
	close(ch)
}
