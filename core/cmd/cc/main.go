//go:build linux
// +build linux

package main

import (
	"flag"
	"log"
	"os"

	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/tools"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/operator"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/server"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	cdn2proxy "github.com/jm33-m0/go-cdn2proxy"
)

// Options struct to hold flag values
type Options struct {
	isServer      bool
	operator_ip   string
	operator_port int
	cdnProxy      string
	config        string
	debug         bool
}

const (
	operatorDefaultPort = 13377
	operatorDefaultIP   = "127.0.0.1"
)

func parseFlags() *Options {
	opts := &Options{}
	flag.StringVar(&opts.cdnProxy, "cdn2proxy", "", "Start cdn2proxy server on this port")
	flag.IntVar(&opts.operator_port, "port", operatorDefaultPort, "C2 server port")
	flag.StringVar(&opts.operator_ip, "ip", operatorDefaultIP, "Connect to this C2 server to start operations")
	flag.BoolVar(&opts.debug, "debug", false, "Do not kill tmux session when crashing, so you can see the crash log")
	flag.BoolVar(&opts.isServer, "server", false, "Run as C2 operator server (default: false, run as operator client)")
	flag.Parse()
	return opts
}

func init() {
	// log to file
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home dir: %v", err)
	}
	live.EmpLogFile = home + "/.emp3r0r/emp3r0r.log"
	logf, err := os.OpenFile(live.EmpLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	logging.SetOutput(logf)

	// set up dirs and default varaibles
	// including config file location
	live.Prompt = cli.Prompt // implement prompt_func
	err = live.InitCC()
	if err != nil {
		log.Fatalf("C2 file paths setup: %v", err)
	}

	// read config file
	err = live.ReadJSONConfig()
	if err != nil {
		logging.Fatalf("Failed to read config: %v", err)
	}

	// set up magic string
	live.InitMagicAgentOneTimeBytes()
}

func main() {
	// Parse command-line flags
	opts := parseFlags()

	// do not kill tmux session when crashing
	if opts.debug {
		live.TmuxPersistence = true
	}

	// abort if CC is already running
	if tools.IsCCRunning() {
		logging.Fatalf("CC is already running")
	}

	// Start cdn2proxy server if specified
	if opts.cdnProxy != "" {
		startCDN2Proxy(opts)
	}

	if opts.isServer {
		server.ServerMain(opts.operator_port)
	} else {
		if opts.operator_ip == operatorDefaultIP {
			logging.Warningf("Operator server IP is %s, C2 server will run along with CLI", operatorDefaultIP)
			go server.ServerMain(opts.operator_port)
		}
		operator.CliMain(opts.operator_ip, opts.operator_port)
	}
}

// helper function to start the cdn2proxy server
func startCDN2Proxy(opts *Options) {
	go func() {
		logFile, openErr := os.OpenFile("/tmp/ws.log", os.O_CREATE|os.O_RDWR, 0o600)
		if openErr != nil {
			logging.Fatalf("OpenFile: %v", openErr)
		}
		openErr = cdn2proxy.StartServer(opts.cdnProxy, "127.0.0.1:"+live.RuntimeConfig.CCPort, "ws", logFile)
		if openErr != nil {
			logging.Fatalf("CDN StartServer: %v", openErr)
		}
	}()
}
