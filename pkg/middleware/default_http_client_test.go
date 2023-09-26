package middleware

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigureHttp2TransportPing(t *testing.T) {
	t.Run("transport should be setup with http2Transport h2 middleware", func(t *testing.T) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
		require.NotContains(t, tr.TLSClientConfig.NextProtos, "h2")
		configureHttp2TransportPing(tr)
		require.Contains(t, tr.TLSClientConfig.NextProtos, "h2")
	})
}
