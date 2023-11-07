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
			TLSClientConfig: &tls.Config{},
		}
		require.NotContains(t, tr.TLSClientConfig.NextProtos, "h2")
		configureHttp2TransportPing(tr)
		require.Contains(t, tr.TLSClientConfig.NextProtos, "h2")
	})

	t.Run("configuring transport twice panics", func(t *testing.T) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{},
		}
		require.NotContains(t, tr.TLSClientConfig.NextProtos, "h2")
		require.NotPanics(t, func() { configureHttp2TransportPing(tr) })
		require.Panics(t, func() { configureHttp2TransportPing(tr) })
		require.Contains(t, tr.TLSClientConfig.NextProtos, "h2")
	})

	t.Run("defaultTransport is configured with h2 by default", func(t *testing.T) {
		// should panic because it's already configured
		require.Panics(t, func() { configureHttp2TransportPing(defaultTransport) })
		require.Contains(t, defaultTransport.TLSClientConfig.NextProtos, "h2")
	})
}
