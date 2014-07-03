package main

import (
	"flag"
	"github.com/cratonica/trayhost"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// Autoupdate github repository (eg. blang/hornet).
// Inserted on compiletime
var gh_repo string

// Url to webinterface
// Inserted on compiletime
var webif_url string

// Current version
// Inserted on compiletime
var version string

func main() {
	var (
		listen    = flag.String("listen", "127.0.0.1:7000", "REST API, e.g. 127.0.0.1:7000")
		storeFile = flag.String("store", "store.gob", "Store file")
		debug     = flag.Bool("debug", false, "Enable debug mode")
	)
	flag.Parse()

	// Setup logging
	if *debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Could not write debug.log: %s", err)
		}
		log.SetOutput(io.MultiWriter(f, os.Stdout))
	} else {
		log.SetOutput(ioutil.Discard)
	}

	log.Println("Version: " + version)
	go func() {
		err := AutoUpdate("0.1.0", gh_repo)
		if err != nil {
			log.Printf("Error: %v\n", err)
		} else {
			log.Println("Successfully updated/Already up to date")
		}
	}()

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
	go tray(trayCloseCh)

	// Block until a signal is received.
	select {
	case s := <-c:
		log.Printf("Received signal %q, shut down gracefully\n", s)
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

	// Be sure to call this to link the tray icon to the target url
	trayhost.SetUrl(webif_url)

	// Enter the host system's event loop
	trayhost.EnterLoop("Hornet", iconData)

	// This is only reached once the user chooses the Exit menu item
	log.Println("Tray Exiting")
	close(ch)
}
