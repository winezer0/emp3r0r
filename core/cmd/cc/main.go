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
	"github.com/jm33-m0/emp3r0r/core/lib/netutil"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
	cdn2proxy "github.com/jm33-m0/go-cdn2proxy"
)

// Options struct to hold flag values
type Options struct {
	isServer           bool   // Run as C2 operator server
	wg_server_ip       string // C2 operator server IP, default: 127.0.0.1
	wg_server_port     int    // C2 operator server port (WireGuard), default: 13377
	wg_server_peer_key string // C2 operator server wireguard public key
	cdnProxy           string // Start cdn2proxy server on this port
	debug              bool   // Do not kill tmux session when crashing
}

const (
	operatorDefaultPort = 13377
	operatorDefaultIP   = "127.0.0.1"
)

func parseFlags() *Options {
	opts := &Options{}
	flag.StringVar(&opts.cdnProxy, "cdn2proxy", "", "Start cdn2proxy server on this port")
	flag.IntVar(&opts.wg_server_port, "port", operatorDefaultPort, "C2 server port")
	flag.StringVar(&opts.wg_server_ip, "ip", operatorDefaultIP, "Connect to this C2 server to start operations")
	flag.StringVar(&opts.wg_server_peer_key, "peer", "", "WireGuard public key provided by the C2 server")
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
	err = live.SetupFilePaths()
	if err != nil {
		log.Fatalf("C2 file paths setup: %v", err)
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
		live.IsServer = true
		logging.AddWriter(os.Stderr)
		err := live.InitCertsAndConfig()
		if err != nil {
			logging.Fatalf("Failed to init certs and config: %v", err)
		}
		err = live.LoadConfig()
		if err != nil {
			logging.Fatalf("Failed to load config: %v", err)
		}
		server.ServerMain(opts.wg_server_port)
	} else {
		if opts.wg_server_ip == operatorDefaultIP {
			logging.Warningf("Operator server IP is %s, C2 server will run along with CLI", operatorDefaultIP)
			go server.ServerMain(opts.wg_server_port)
		}
		connectWg(opts)
		err := live.LoadConfig()
		if err != nil {
			logging.Fatalf("Failed to load config: %v", err)
		}
		operator.CliMain(opts.wg_server_ip, opts.wg_server_port)
	}
}

func connectWg(opts *Options) {
	if opts.wg_server_peer_key == "" {
		logging.Fatalf("Please provide the server's WireGuard public key")
	}
	// Connect to C2 wireguard server with given wireguard keypair
	wg_key := live.Prompt("Enter operator's WireGuard private key provided by the server: ")
	_, err := netutil.PublicKeyFromPrivate(wg_key)
	if err != nil {
		log.Fatalf("Invalid key: %v", err)
	}
	wgConfig := netutil.WireGuardConfig{
		PrivateKey: wg_key,
		IPAddress:  netutil.WgOperatorIP + "/24",
		ListenPort: util.RandInt(1024, 65535),
		Peers: []netutil.PeerConfig{
			{
				PublicKey:  opts.wg_server_peer_key,
				AllowedIPs: netutil.WgServerIP + "/32",
			},
		},
	}
	go func() {
		_, err = netutil.WireGuardMain(wgConfig)
		if err != nil {
			logging.Fatalf("Connecting to C2 WireGuard server: %v", err)
		}
		logging.Successf("Connected to C2 WireGuard server at %s:%d", opts.wg_server_ip, opts.wg_server_port)
	}()
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
