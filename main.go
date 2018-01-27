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
	"path/filepath"
	"sync"
	"syscall"

	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"

	"github.com/BurntSushi/toml"
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
	Server serverConfig
	Log    LogConfig
}

// serverConfig is configuration for websocket server
type serverConfig struct {
	Port      uint
	Endpoint  string
	Debug     bool
	EnableTLS bool
	KeyFile   string
	CertFile  string
	DataDir   string
}

// Global variables are usually a bad practice but we will use them this time for simplicity.
//var clients = make(map[*websocket.Conn]bool) // connected clients
//var broadcast = make(chan Message) // broadcast channel
var writer = make(chan Message) // exporter channels

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
		log.Fatalf("Failed to load config file '%s'. ", confPath)
	}

	ConfigLogging(conf.Log)
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

	binding := fmt.Sprintf("%v:%d", conf.Server.Endpoint, conf.Server.Port)
	log.Printf("[INFO] %s bound on %v", distName, binding)

	laddr, err := net.ResolveTCPAddr("tcp", binding)
	if err != nil {
		log.Printf("[WARN] %v", err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Printf("[WARN] %v", err)
	}
	defer listener.Close()

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

	// 保管用ディレクトリの準備
	dataDirPath := conf.Server.DataDir
	if _, err := os.Stat(dataDirPath); err != nil {
		if err := os.Mkdir(dataDirPath, os.ModePerm); err != nil {
			log.Printf("[WARN] %v", err)
			dataDirPath = "."
		}
	}

	// 標準出力用ゴルーチン起動
	go func() {
	loop:
		for {
			select {
			case msg, ok := <-writer:
				if !ok { // selectでchanのクローズを検知する方法
					fmt.Println("writer channel is closed")
					break loop
				}

				dectateCSV(msg, dataDirPath)
				SavePost(msg, dataDirPath) //
			}
		}
	}()

	// handler on http endpoint
	router := registHandlers(conf.Log.AccessLog)
	server := &http.Server{Handler: router, ConnState: connectionStateChange}

	//TODO
	if conf.Server.EnableTLS {
		certFilePath, _ := filepath.Abs(conf.Server.CertFile)
		keyFilePath, _ := filepath.Abs(conf.Server.KeyFile)
		log.Printf("[INFO] TLS enabled. cert: %v key: %v", certFilePath, keyFilePath)
		err := server.ServeTLS(listener, certFilePath, keyFilePath)
		log.Fatal(err)
	} else {
		log.Println("[INFO] TLS disabled.")
		err := server.Serve(listener)
		log.Fatal(err)
	}
	activeConnWaiting.Wait()

	code := <-exitCh
	os.Exit(code)
}
