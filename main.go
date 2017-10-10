package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/websocket"
)

/*
 * On build time, they will be set with -X option
 * Version software version
 */
var (
	Version  string
	distName string
)

// config is master configuration
type config struct {
	server serverConfig
	log    LogConfig
}

// serverConfig is configuration for websocket server
type serverConfig struct {
	Port     uint
	Endpoint string
	Debug    bool
}

// Global variables are usually a bad practice but we will use them this time for simplicity.
var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// registHandlers maps URL paths to handler functions
func registHandlers(logPath string) http.Handler {
	log.Printf("[DEBUG] registHandlers")
	logger := openLogFile(logPath)
	go hub.run()

	r := mux.NewRouter()
	// Configure websocket route
	r.HandleFunc("/ws/{room}", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			serveWs(&hub, w, r)
		},
	))
	// Create a simple file server
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	return handlers.LoggingHandler(logger, r)
}

var conf config

func init() {
	var confPath string
	flag.StringVar(&confPath, "c", "tavle.tml", "Path to config file")
	flag.Parse()

	if _, err := toml.DecodeFile(confPath, &conf); err != nil {
		log.Println(err)
		log.Fatalln("Failed to load config file.", confPath)
	}
	ConfigLogging(conf.log)
}

var activeConnWaiting sync.WaitGroup
var numberOfActive = 0

func connectionStateChange(c net.Conn, st http.ConnState) {
	if st == http.StateActive {
		activeConnWaiting.Add(1)
		numberOfActive++
	} else if st == http.StateIdle || st == http.StateHijacked {
		activeConnWaiting.Done()
		numberOfActive--
	}
	log.Printf("[INFO] %d active connections.\n", numberOfActive)
}

func main() {
	// Channel to catch signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	binding := fmt.Sprintf("%v:%d", conf.server.Endpoint, conf.server.Port)
	log.Printf("[INFO] %s bound on %v", distName, binding)

	laddr, _ := net.ResolveTCPAddr("tcp", binding)
	listener, _ := net.ListenTCP("tcp", laddr)

	exitCh := make(chan int)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP:
			log.Println("[INFO] Reloading talk history.")
			//update()
		default:
			log.Println("[WARN] Receive a signal.", sig)
			listener.Close()
			log.Printf("[INFO] %v have went down. Bye.", distName)
			exitCh <- 0
		}
	}()

	// handler on http endpoint
	router := registHandlers(conf.log.accessLog)
	server := &http.Server{Handler: router, ConnState: connectionStateChange}
	server.Serve(listener)
	activeConnWaiting.Wait()

	code := <-exitCh
	os.Exit(code)
}
