package temporal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/taonic/ticketfu/config"
	"go.temporal.io/sdk/client"
)

// headerProvider implements the HeadersProvider interface required by Temporal
type headerProvider struct {
	apiKey string
}

// GetHeaders implements the HeadersProvider interface
func (h *headerProvider) GetHeaders(ctx context.Context) (map[string]string, error) {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", h.apiKey),
	}, nil
}

// NewClient creates a new Temporal client using the provided configuration
func NewClient(config config.TemporalClientConfig) (client.Client, error) {
	options := client.Options{
		HostPort:  config.Address,
		Namespace: config.Namespace,
	}

	// Configure TLS if enabled
	if config.UseSSL {
		tlsConfig, err := createTLSConfig(
			config.TLSCertPath,
			config.TLSKeyPath,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		options.ConnectionOptions = client.ConnectionOptions{
			TLS: tlsConfig,
		}
	}

	// Add API key if provided
	if config.APIKey != "" {
		options.HeadersProvider = &headerProvider{apiKey: config.APIKey}
	}

	// Create Temporal client
	c, err := client.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	return c, nil
}

// createTLSConfig creates a TLS configuration for secure connections
func createTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509 key pair: %w", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("failed to append CA certs")
	}

	// Create TLS configuration
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}, nil
}
