package core

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"github.com/jm33-m0/emp3r0r/core/internal/live"
	"golang.org/x/net/http2"
)

var (
	// OperatorHTTPClient is an HTTP/2 client for the mTLS C2 operator server
	OperatorHTTPClient *http.Client

	// OperatorRootURL is the root URL of the mTLS C2 operator server
	OperatorRootURL string
)

// createMTLSHttpClient connects to the mTLS server and returns an HTTP/2 client
func createMTLSHttpClient() (*http.Client, error) {
	// Load client certificate
	clientCert, err := tls.LoadX509KeyPair(live.OperatorCrtFile, live.OperatorKeyFile)
	if err != nil {
		return nil, err
	}

	// Load CA certificate
	caCert, err := os.ReadFile(live.OperatorCACrtFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Configure TLS
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caCertPool,
	}

	// Create HTTP/2 transport
	transport := &http2.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Create HTTP client
	client := &http.Client{
		Transport: transport,
	}

	return client, nil
}
