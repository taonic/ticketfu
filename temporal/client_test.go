package temporal

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
