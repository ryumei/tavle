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

	"github.com/BurntSushi/toml"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/logutils"
)

/*
 * On build time, they will be set with -X option
 * Version software version
 * Revision sofotware revision
 */
var (
	Version  string
	Revision string
	distName string
)

// Config is master configuration
type Config struct {
	Server ServerConfig
	Log    LogConfig
}

// ServerConfig is configuration for websocket server
type ServerConfig struct {
	Port     uint
	Endpoint string
	Debug    bool
}

// LogConfig is configuration for logging
type LogConfig struct {
	accesslog string
	ServerLog string
	Level     string
}

// Global variables are usually a bad practice but we will use them this time for simplicity.
var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{}

// Message is message object
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("[ERROR] error: ", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the bradcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the nexe message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func openLogFile(logPath string) *os.File {
	logWriter, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return os.Stderr
	}
	return logWriter
}

func logConfig(conf LogConfig) {
	logWriter := openLogFile(conf.ServerLog)

	// Logging with logutils
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(conf.Level),
		Writer:   logWriter,
	}
	log.SetOutput(filter)
	logFlags := log.LstdFlags | log.Lmicroseconds | log.LUTC
	if filter.MinLevel == "DEBUG" {
		logFlags |= log.Lshortfile
	}
	log.SetFlags(logFlags)
}

var r *http.ServeMux

// registHandlers maps URL paths to handler functions
func registHandlers(logPath string) {
	logger := openLogFile(logPath)

	r = http.NewServeMux()

	// Create a simple file server
	r.Handle("/", handlers.LoggingHandler(logger, http.FileServer(http.Dir("./public"))))

	// Configure websocket route
	for route, f := range map[string]func(http.ResponseWriter, *http.Request){
		"/ws": handleConnections,
	} {
		r.Handle(route, handlers.LoggingHandler(logger, http.HandlerFunc(f)))
	}

	go handleMessages()
}

var config Config

func init() {
	var confPath string
	flag.StringVar(&confPath, "c", "tavle.tml", "Path to config file")
	flag.Parse()

	if _, err := toml.DecodeFile(confPath, &config); err != nil {
		log.Println(err)
		log.Fatalln("Failed to load config file.", confPath)
	}
	logConfig(config.Log)

	// handler on http endpoint
	registHandlers(config.Log.accesslog)
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

	binding := fmt.Sprintf("%v:%d", config.Server.Endpoint, config.Server.Port)
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

	server := &http.Server{Handler: r, ConnState: connectionStateChange}
	server.Serve(listener)
	activeConnWaiting.Wait()

	code := <-exitCh
	os.Exit(code)
}
