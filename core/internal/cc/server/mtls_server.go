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
	"github.com/jm33-m0/emp3r0r/core/internal/transport"
	"github.com/jm33-m0/emp3r0r/core/lib/logging"
)

// StartMTLSServer starts the operator TLS server with mTLS.
func StartMTLSServer(port int) {
	r := mux.NewRouter()
	r.HandleFunc(fmt.Sprintf("/%s/{api}/{token}", transport.OperatorRoot), operationDispatcher)
	if network.MTLSServer != nil {
		network.MTLSServer.Shutdown(network.MTLSServerCtx)
	}

	// Load client CA certificate
	clientCACert, err := os.ReadFile(transport.OperatorCaCrtFile)
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
	logging.Printf("Starting C2 operator service with mTLS at port %d", port)
	err = network.MTLSServer.ListenAndServeTLS(transport.OperatorServerCrtFile, transport.OperatorServerKeyFile)
	if err != nil {
		if err == http.ErrServerClosed {
			logging.Warningf("C2 operator service is shutdown")
			return
		}
		logging.Fatalf("Failed to start HTTPS (mTLS) server at *:%d: %v", port, err)
	}
}
