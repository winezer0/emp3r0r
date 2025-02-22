package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/jm33-m0/emp3r0r/core/internal/cc/base/network"
	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
)

// StartMTLSServer starts the operator TLS server with mTLS.
func StartMTLSServer(port int) {
	r := mux.NewRouter()
	transport.CACrtPEM = []byte(live.RuntimeConfig.CAPEM)
	r.HandleFunc(fmt.Sprintf("/%s/{api}/{token}", transport.OperatorRoot), operationDispatcher)
	if network.MTLSServer != nil {
		network.MTLSServer.Shutdown(network.MTLSServerCtx)
	}

	// Load client CA certificate
	clientCACert, err := os.ReadFile(live.OperatorCACrtFile)
	if err != nil {
		logging.Fatalf("Failed to read client CA certificate: %v", err)
	}
	clientCAs := x509.NewCertPool()
	clientCAs.AppendCertsFromPEM(clientCACert)

	// Configure TLS with mTLS
	tlsConfig := &tls.Config{
		ClientCAs:  clientCAs,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	network.MTLSServer = &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   r,
		TLSConfig: tlsConfig,
	}
	network.MTLSServerCtx, network.MTLSServerCancel = context.WithCancel(context.Background())
	logging.Debugf("Starting C2 TLS service with mTLS at port %s", live.RuntimeConfig.CCPort)
	err = network.MTLSServer.ListenAndServeTLS(live.ServerCrtFile, live.ServerKeyFile)
	if err != nil {
		if err == http.ErrServerClosed {
			logging.Warningf("C2 TLS service is shutdown")
			return
		}
		logging.Fatalf("Failed to start HTTPS server at *:%s: %v", live.RuntimeConfig.CCPort, err)
	}
}
