package server

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jm33-m0/arc"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/relay"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/lib/cli"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/netutil"
)

func ServerMain(wg_port int, hosts string) {
	// start all services
	go KCPC2ListenAndServe()
	go tarConfig(hosts)
	wg(wg_port)
	time.Sleep(3 * time.Second)
	go StartC2AgentTLSServer()
	StartOperatorMTLSServer(wg_port + 1)
}

func wg(wg_port int) {
	server_privkey, err := netutil.GeneratePrivateKey()
	if err != nil {
		logging.Fatalf("Failed to generate server private key: %v", err)
	}
	operator_privkey, err := netutil.GeneratePrivateKey()
	if err != nil {
		logging.Fatalf("Failed to generate operator private key: %v", err)
	}
	operator_pubkey, err := netutil.PublicKeyFromPrivate(operator_privkey)
	if err != nil {
		logging.Fatalf("Failed to generate operator public key: %v", err)
	}
	server_pubkey, err := netutil.PublicKeyFromPrivate(server_privkey)
	if err != nil {
		logging.Fatalf("Failed to generate server public key: %v", err)
	}

	// network address
	subnet := netutil.GenerateRandomPrivateSubnet24()
	netutil.WgServerIP, _ = netutil.GenerateRandomIPInSubnet24(subnet)
	netutil.WgOperatorIP, _ = netutil.GenerateRandomIPInSubnet24(subnet)

	wgConfig := netutil.WireGuardConfig{
		IPAddress:     netutil.WgServerIP + "/24",
		InterfaceName: "emp_server",
		ListenPort:    wg_port,
		PrivateKey:    server_privkey,
		Peers: []netutil.PeerConfig{
			{
				PublicKey:  operator_pubkey,
				AllowedIPs: netutil.WgOperatorIP + "/32",
			},
		},
	}
	go func() {
		netutil.WgServer, err = netutil.WireGuardMain(wgConfig)
		if err != nil {
			logging.Fatalf("Failed to start WireGuard server: %v", err)
		}
	}()

	// Create table headers and rows for WireGuard configuration
	headers := []string{"Parameter", "Value"}
	rows := [][]string{
		{"C2 Server IP (WG)", netutil.WgServerIP},
		{"C2 Server Port", strconv.Itoa(wg_port)},
		{"C2 Public Key", server_pubkey},
		{"Operator WG IP", netutil.WgOperatorIP},
		{"Operator Private Key", operator_privkey},
	}

	// Build the table
	tableStr := cli.BuildTable(headers, rows)

	// Print the table with a title
	logging.Successf("\n══════════════════ WireGuard Configuration ════════════════════════════\n\n%s\n", tableStr)
}

func tarConfig(hosts string) {
	err := live.GenC2Certs(hosts)
	if err != nil {
		logging.Fatalf("Failed to generate C2 certs: %v", err)
	}
	err = os.Chdir(live.EmpWorkSpace)
	if err != nil {
		logging.Fatalf("Failed to change directory: %v", err)
	}
	// tar all config files
	filter := func(path string) bool {
		return strings.HasSuffix(path, ".log") || strings.HasPrefix(path, "stub") || strings.HasSuffix(path, ".history")
	}
	os.Chdir(filepath.Dir(live.EmpWorkSpace))
	defer os.Chdir(live.EmpWorkSpace)

	err = arc.ArchiveWithFilter(filepath.Base(live.EmpWorkSpace), live.EmpConfigTar, arc.CompressionMap["xz"], arc.ArchivalMap["tar"], filter)
	if err != nil {
		logging.Errorf("Failed to tar config files: %v", err)
	}
	err = relay.WgFileServer(live.EmpConfigTar)
	if err != nil {
		logging.Errorf("Failed to start file server to serve config tarball: %v", err)
	}
}
