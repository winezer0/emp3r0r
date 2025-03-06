//go:build linux
// +build linux

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/ftp"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/tools"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/operator"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/server"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/netutil"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
	cdn2proxy "github.com/jm33-m0/go-cdn2proxy"
	"github.com/spf13/cobra"
)

// Options struct to hold flag values
type Options struct {
	c2_server_ip   string // C2 server IP
	c2_server_port int    // C2 server port
	wg_server_key  string // C2 server's WireGuard public key
	wg_server_ip   string // C2 server's WireGuard IP
	wg_operator_ip string // Operator's WireGuard IP
	c2_hosts       string // C2 hosts to generate cert for
	cdnProxy       string // Start cdn2proxy server on this port
	debug          bool   // Do not kill tmux session when crashing
}

const (
	operatorDefaultPort = 13377
	operatorDefaultIP   = "127.0.0.1"
)

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

	// set up dirs and default variables
	live.Prompt = cli.Prompt
	err = live.SetupFilePaths()
	if err != nil {
		log.Fatalf("C2 file paths setup: %v", err)
	}
}

func main() {
	opts := &Options{}

	// Root command
	rootCmd := &cobra.Command{
		Use:   "emp3r0r",
		Short: "emp3r0r C2 framework",
		Long:  "A Linux C2 made by a Linux user",
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&opts.cdnProxy, "cdn2proxy", "", "Start cdn2proxy server on this port")

	// Client subcommand
	clientCmd := &cobra.Command{
		Use:   "client",
		Short: "Run as C2 operator client",
		Run: func(cmd *cobra.Command, args []string) {
			runClientMode(opts)
		},
	}

	// Client-specific flags
	clientCmd.Flags().StringVar(&opts.c2_server_ip, "c2-host", operatorDefaultIP, "Connect to this C2 server to start operations")
	clientCmd.Flags().IntVar(&opts.c2_server_port, "c2-port", operatorDefaultPort, "C2 server port")
	clientCmd.Flags().StringVar(&opts.wg_server_key, "server-wg-key", "", "WireGuard public key provided by the C2 server")
	clientCmd.Flags().StringVar(&opts.wg_server_ip, "server-wg-ip", "", "WireGuard server IP provided by the C2 server")
	clientCmd.Flags().StringVar(&opts.wg_operator_ip, "operator-wg-ip", "", "Operator's wireguard IP")
	clientCmd.Flags().BoolVar(&opts.debug, "debug", false, "Do not kill tmux session when crashing, so you can see the crash log")

	// Mark required flags for client mode
	clientCmd.MarkFlagRequired("c2-host")
	clientCmd.MarkFlagRequired("c2-port")
	clientCmd.MarkFlagRequired("server-wg-key")
	clientCmd.MarkFlagRequired("server-wg-ip")
	clientCmd.MarkFlagRequired("operator-wg-ip")

	// Server subcommand
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Run as C2 operator server",
		Run: func(cmd *cobra.Command, args []string) {
			runServerMode(opts)
		},
	}

	// Server-specific flags
	serverCmd.Flags().IntVar(&opts.c2_server_port, "port", operatorDefaultPort, "Server port to listen on")
	serverCmd.Flags().StringVar(&opts.c2_hosts, "c2-hosts", "", "C2 hosts to generate cert for, separated by whitespace")

	// Completion command
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh]",
		Short: "Generate shell completion scripts",
		Long: `To load completions:

Bash:
  $ source <(emp3r0r completion bash)
  # To load completions for each session, execute once:
  $ emp3r0r completion bash > /etc/bash_completion.d/emp3r0r

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  
  # To load completions for each session, execute once:
  $ emp3r0r completion zsh > "${fpath[1]}/_emp3r0r"
  # You will need to start a new shell for this setup to take effect.
`,
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"bash", "zsh"},
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			}
		},
	}

	// Add subcommands to root
	rootCmd.AddCommand(clientCmd, serverCmd, completionCmd)

	// Default behavior if no subcommand is given (backwards compatibility)
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		// Default to client mode with local server
		opts.c2_server_ip = operatorDefaultIP
		runClientMode(opts)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

func runClientMode(opts *Options) {
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

	if opts.c2_server_ip == operatorDefaultIP {
		logging.Warningf("Operator server IP is %s, C2 server will run along with CLI", operatorDefaultIP)
		go server.ServerMain(opts.c2_server_port, opts.c2_hosts)
	}

	connectWg(opts)

	// download and extract config files
	url := fmt.Sprintf("http://%s:%d/%s", netutil.WgServerIP, netutil.WgFileServerPort, "emp3r0r_operator_config.tar.xz")
	err := live.DownloadExtractConfig(url, ftp.DownloadFile)
	if err != nil {
		logging.Fatalf("Failed to extract config: %v", err)
	}
	err = live.LoadConfig()
	if err != nil {
		logging.Fatalf("Failed to load config: %v", err)
	}
	operator.CliMain(opts.c2_server_ip, opts.c2_server_port)
}

func runServerMode(opts *Options) {
	live.IsServer = true
	logging.AddWriter(os.Stderr)

	// abort if CC is already running
	if tools.IsCCRunning() {
		logging.Fatalf("CC is already running")
	}

	// Start cdn2proxy server if specified
	if opts.cdnProxy != "" {
		startCDN2Proxy(opts)
	}

	err := live.InitCertsAndConfig()
	if err != nil {
		logging.Fatalf("Failed to init certs and config: %v", err)
	}
	err = live.LoadConfig()
	if err != nil {
		logging.Fatalf("Failed to load config: %v", err)
	}
	server.ServerMain(opts.c2_server_port, opts.c2_hosts)
}

func connectWg(opts *Options) {
	if opts.wg_server_key == "" {
		logging.Fatalf("Please provide the server's WireGuard public key")
	}
	if opts.wg_server_ip == "" {
		logging.Fatalf("Please provide the server's WireGuard IP")
	}
	if opts.wg_operator_ip == "" {
		logging.Fatalf("Please provide the operator's WireGuard IP")
	}
	netutil.WgServerIP = opts.wg_server_ip
	netutil.WgOperatorIP = opts.wg_operator_ip
	operator.SERVER_IP = opts.c2_server_ip
	operator.SERVER_KEY = opts.wg_server_key
	// Connect to C2 wireguard server with given wireguard keypair
	wg_key := live.Prompt("Enter operator's WireGuard private key provided by the server")
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
				PublicKey:  opts.wg_server_key,
				AllowedIPs: netutil.WgServerIP + "/32",
				Endpoint:   fmt.Sprintf("%s:%d", opts.c2_server_ip, opts.c2_server_port),
			},
		},
	}
	go func() {
		_, err = netutil.WireGuardMain(wgConfig)
		if err != nil {
			logging.Fatalf("Connecting to C2 WireGuard server: %v", err)
		}
		logging.Successf("Connected to C2 WireGuard server at %s:%d", opts.c2_server_ip, opts.c2_server_port)
	}()
	time.Sleep(2 * time.Second)
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
