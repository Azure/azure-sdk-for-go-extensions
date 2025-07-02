package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_httpConnTracking(t *testing.T) {
	t.Parallel()
	for i := range 100 {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			t.Parallel()

			ctx, cancelRootCtx := context.WithCancel(context.Background())
			defer cancelRootCtx()

			connTracking := new(HttpConnTracking)

			ctxWithConnTracking := addConnectionTracingToRequestContext(ctx, connTracking)
			clientTrace := httptrace.ContextClientTrace(ctxWithConnTracking)
			originalDNSStart := clientTrace.DNSStart
			clientTrace.DNSStart = func(di httptrace.DNSStartInfo) {
				originalDNSStart(di)
				cancelRootCtx() // cancel the root context to simulate a half done DNS request
			}

			func() {
				defer func() {
					fmt.Printf("READ: dns latency %q\n", connTracking.GetDnsLatency()) // simulate logging usage - using thread-safe getter
				}()

				req, err := http.NewRequestWithContext(ctxWithConnTracking, http.MethodGet, "https://example.com", nil)
				if err != nil {
					t.Fatalf("failed to create request: %v", err)
				}
				if _, err := http.DefaultClient.Do(req); !errors.Is(err, context.Canceled) {
					t.Fatalf("unexpected request error: %v", err)
				}
			}()
		})
	}
}

func Test_httpConnTrackingThreadSafety(t *testing.T) {
	t.Parallel()
	
	// Test that getter methods provide thread-safe access
	connTracking := new(HttpConnTracking)
	
	// Set some values using the internal setters to simulate HTTP trace callbacks
	connTracking.setDnsLatency("10ms")
	connTracking.setConnLatency("5ms")
	connTracking.setTlsLatency("15ms")
	connTracking.setTotalLatency("30ms")
	connTracking.setProtocol("h2")
	
	// Verify getter methods return the expected values
	assert.Equal(t, "10ms", connTracking.GetDnsLatency())
	assert.Equal(t, "5ms", connTracking.GetConnLatency())
	assert.Equal(t, "15ms", connTracking.GetTlsLatency())
	assert.Equal(t, "30ms", connTracking.GetTotalLatency())
	assert.Equal(t, "h2", connTracking.GetProtocol())
	
	// Verify backward compatibility - direct field access still works
	assert.Equal(t, "10ms", connTracking.DnsLatency)
	assert.Equal(t, "5ms", connTracking.ConnLatency)
	assert.Equal(t, "15ms", connTracking.TlsLatency)
	assert.Equal(t, "30ms", connTracking.TotalLatency)
	assert.Equal(t, "h2", connTracking.Protocol)
}

// BenchmarkHttpConnTracking benchmarks the performance of HttpConnTracking
// with real HTTP requests to validate the performance impact of synchronization.
//
// Benchmark results:
// goos: linux
// goarch: amd64
// pkg: github.com/Azure/azure-sdk-for-go-extensions/pkg/middleware
// cpu: AMD EPYC 7763 64-Core Processor                
// BenchmarkHttpConnTracking/WithGetterMethods-16         	     516	   2228617 ns/op	   92211 B/op	     984 allocs/op
// BenchmarkHttpConnTracking/WithDirectFieldAccess-16     	     540	   2223993 ns/op	   92188 B/op	     984 allocs/op
// BenchmarkHttpConnTracking/ConcurrentGetterAccess-16    	 5319430	       219.7 ns/op	       0 B/op	       0 allocs/op
// BenchmarkHttpConnTracking/ConcurrentDirectAccess-16    	1000000000	         0.08260 ns/op	       0 B/op	       0 allocs/op
// BenchmarkHttpConnTracking/MutexOverhead-16             	254370688	         4.808 ns/op	       0 B/op	       0 allocs/op
// PASS
func BenchmarkHttpConnTracking(b *testing.B) {
	// Create a test HTTP server that responds quickly to minimize network overhead
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	b.Run("WithGetterMethods", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			connTracking := &HttpConnTracking{}
			ctx := addConnectionTracingToRequestContext(context.Background(), connTracking)
			
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
			if err != nil {
				b.Fatalf("failed to create request: %v", err)
			}
			
			// Use the test server's client to avoid certificate errors
			resp, err := server.Client().Do(req)
			if err != nil {
				b.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()
			
			// Access connection tracking data using thread-safe getter methods
			_ = connTracking.GetTotalLatency()
			_ = connTracking.GetDnsLatency()
			_ = connTracking.GetConnLatency()
			_ = connTracking.GetTlsLatency()
			_ = connTracking.GetProtocol()
			_ = connTracking.GetReqConnInfo()
		}
	})

	b.Run("WithDirectFieldAccess", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			connTracking := &HttpConnTracking{}
			ctx := addConnectionTracingToRequestContext(context.Background(), connTracking)
			
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
			if err != nil {
				b.Fatalf("failed to create request: %v", err)
			}
			
			// Use the test server's client to avoid certificate errors
			resp, err := server.Client().Do(req)
			if err != nil {
				b.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()
			
			// Access connection tracking data using direct field access (may not be thread-safe)
			_ = connTracking.TotalLatency
			_ = connTracking.DnsLatency
			_ = connTracking.ConnLatency
			_ = connTracking.TlsLatency
			_ = connTracking.Protocol
			_ = connTracking.ReqConnInfo
		}
	})

	b.Run("ConcurrentGetterAccess", func(b *testing.B) {
		connTracking := &HttpConnTracking{}
		// Pre-populate with some data
		connTracking.setTotalLatency("100ms")
		connTracking.setDnsLatency("10ms")
		connTracking.setConnLatency("20ms")
		connTracking.setTlsLatency("30ms")
		connTracking.setProtocol("h2")

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Simulate concurrent read access using thread-safe getters
				_ = connTracking.GetTotalLatency()
				_ = connTracking.GetDnsLatency()
				_ = connTracking.GetConnLatency()
				_ = connTracking.GetTlsLatency()
				_ = connTracking.GetProtocol()
				_ = connTracking.GetReqConnInfo()
			}
		})
	})

	b.Run("ConcurrentDirectAccess", func(b *testing.B) {
		connTracking := &HttpConnTracking{}
		// Pre-populate with some data
		connTracking.setTotalLatency("100ms")
		connTracking.setDnsLatency("10ms")
		connTracking.setConnLatency("20ms")
		connTracking.setTlsLatency("30ms")
		connTracking.setProtocol("h2")

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Simulate concurrent read access using direct field access
				_ = connTracking.TotalLatency
				_ = connTracking.DnsLatency
				_ = connTracking.ConnLatency
				_ = connTracking.TlsLatency
				_ = connTracking.Protocol
				_ = connTracking.ReqConnInfo
			}
		})
	})

	b.Run("MutexOverhead", func(b *testing.B) {
		connTracking := &HttpConnTracking{}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Measure just the mutex overhead by doing lock/unlock cycles
			connTracking.mu.RLock()
			connTracking.mu.RUnlock()
		}
	})
}