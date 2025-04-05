package temporal

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taonic/ticketfu/config"
	"go.temporal.io/server/common/log"
)

// TestHeaderProvider tests that the headerProvider correctly returns headers
func TestHeaderProvider(t *testing.T) {
	provider := &headerProvider{namespace: "test-namespace"}
	headers, err := provider.GetHeaders(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "test-namespace", headers["temporal-namespace"])
}

// TestClientIdentity tests that the client identity is correctly formatted
func TestClientIdentity(t *testing.T) {
	identity := clientIdentity()

	assert.Contains(t, identity, "ticketfu:")

	hostname, err := os.Hostname()
	if err == nil {
		assert.Contains(t, identity, hostname)
	}
}

// TestCreateTLSConfig tests the creation of TLS configuration
func TestFailedCreateTLSConfig(t *testing.T) {
	certFile, err := os.CreateTemp("", "test-cert-*.pem")
	require.NoError(t, err)
	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "test-key-*.pem")
	require.NoError(t, err)
	defer os.Remove(keyFile.Name())

	// Write test certificate and key content
	// This is not a valid cert/key pair, just to test file loading
	certContent := `-----BEGIN CERTIFICATE-----
MIICMzCCAZygAwIBAgIJALiPnVsvq8dsMA0GCSqGSIb3DQEBBQUAMFMxCzAJBgNV
-----END CERTIFICATE-----`

	keyContent := `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDB/xdITyjuMKvmLMQAufdwdfKENQQRW31odTvEhpDFeXi9UfEZ
-----END RSA PRIVATE KEY-----`

	_, err = certFile.WriteString(certContent)
	require.NoError(t, err)

	_, err = keyFile.WriteString(keyContent)
	require.NoError(t, err)

	certFile.Close()
	keyFile.Close()

	tlsConfig, err := createTLSConfig(certFile.Name(), keyFile.Name())

	// This should fail because our test cert is not valid
	assert.Error(t, err)
	assert.Nil(t, tlsConfig)
	assert.Contains(t, err.Error(), "failed to load X509 key pair")
}

// TestNewClientWithAPIKey tests client creation with API key configuration
func TestNewClientWithAPIKey(t *testing.T) {
	testConfig := config.TemporalClientConfig{
		Address:   "localhost:7233", // Use local address for test
		Namespace: "test-namespace",
		APIKey:    "test-api-key",
	}

	testLogger := log.NewTestLogger()

	_, err := NewClient(testConfig, testLogger)

	// We expect an error because the API key is not valid,
	// but we can still test some of the configuration
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Temporal client")
}

func TestCreateTLSConfigWithValidCerts(t *testing.T) {
	// Create temporary files for test certificates
	certFile, err := os.CreateTemp("", "test-cert-*.pem")
	require.NoError(t, err)
	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "test-key-*.pem")
	require.NoError(t, err)
	defer os.Remove(keyFile.Name())

	// Write test certificate and key content
	// This is a self-signed certificate/key pair for testing only
	// DO NOT use this in production!
	certContent := `-----BEGIN CERTIFICATE-----
MIICEjCCAXsCAg36MA0GCSqGSIb3DQEBBQUAMIGbMQswCQYDVQQGEwJKUDEOMAwG
A1UECBMFVG9reW8xEDAOBgNVBAcTB0NodW8ta3UxETAPBgNVBAoTCEZyYW5rNERE
MRgwFgYDVQQLEw9XZWJDZXJ0IFN1cHBvcnQxGDAWBgNVBAMTD0ZyYW5rNEREIFdl
YiBDQTEjMCEGCSqGSIb3DQEJARYUc3VwcG9ydEBmcmFuazRkZC5jb20wHhcNMTIw
ODIyMDUyNjU0WhcNMTcwODIxMDUyNjU0WjBKMQswCQYDVQQGEwJKUDEOMAwGA1UE
CAwFVG9reW8xETAPBgNVBAoMCEZyYW5rNEREMRgwFgYDVQQDDA93d3cuZXhhbXBs
ZS5jb20wXDANBgkqhkiG9w0BAQEFAANLADBIAkEAm/xmkHmEQrurE/0re/jeFRLl
8ZPjBop7uLHhnia7lQG/5zDtZIUC3RVpqDSwBuw/NTweGyuP+o8AG98HxqxTBwID
AQABMA0GCSqGSIb3DQEBBQUAA4GBABS2TLuBeTPmcaTaUW/LCB2NYOy8GMdzR1mx
8iBIu2H6/E2tiY3RIevV2OW61qY2/XRQg7YPxx3ffeUugX9F4J/iPnnu1zAxxyBy
2VguKv4SWjRFoRkIfIlHX0qVviMhSlNy2ioFLy7JcPZb+v3ftDGywUqcBiVDoea0
Hn+GmxZA
-----END CERTIFICATE-----`

	keyContent := `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBAJv8ZpB5hEK7qxP9K3v43hUS5fGT4waKe7ix4Z4mu5UBv+cw7WSF
At0Vaag0sAbsPzU8Hhsrj/qPABvfB8asUwcCAwEAAQJAG0r3ezH35WFG1tGGaUOr
QA61cyaII53ZdgCR1IU8bx7AUevmkFtBf+aqMWusWVOWJvGu2r5VpHVAIl8nF6DS
kQIhAMjEJ3zVYa2/Mo4ey+iU9J9Vd+WoyXDQD4EEtwmyG1PpAiEAxuZlvhDIbbce
7o5BvOhnCZ2N7kYb1ZC57g3F+cbJyW8CIQCbsDGHBto2qJyFxbAO7uQ8Y0UVHa0J
BO/g900SAcJbcQIgRtEljIShOB8pDjrsQPxmI1BLhnjD1EhRSubwhDw5AFUCIQCN
A24pDtdOHydwtSB5+zFqFLfmVZplQM/g5kb4so70Yw==
-----END RSA PRIVATE KEY-----`

	_, err = certFile.WriteString(certContent)
	require.NoError(t, err)

	_, err = keyFile.WriteString(keyContent)
	require.NoError(t, err)

	certFile.Close()
	keyFile.Close()

	// Now try to create TLS config with the valid test certificates
	tlsConfig, err := createTLSConfig(certFile.Name(), keyFile.Name())

	// This should succeed with our valid test certificates
	require.NoError(t, err)
	require.NotNil(t, tlsConfig)

	// Verify the TLS config
	assert.Len(t, tlsConfig.Certificates, 1, "Should have 1 certificate")
	assert.NotNil(t, tlsConfig.RootCAs, "RootCAs should be set")
}
