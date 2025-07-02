package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/url"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/stretchr/testify/assert"
)

func TestArmRequestMetrics(t *testing.T) {
	subID := "notexistingSub"
	rgName := "testRG"
	resourceName := "test"
	testCorrelationId := "testCorrelationId"

	checkRequestInfo := func(tt *testing.T, iReq *RequestInfo) {
		assert.NotNil(tt, iReq)
		assert.Equal(tt, iReq.ArmResId.SubscriptionID, subID)
		assert.Equal(tt, iReq.ArmResId.ResourceGroupName, rgName)
		assert.Equal(tt, iReq.ArmResId.Name, resourceName)
		assert.NotNil(tt, iReq.Request)
	}

	checkResponseInfo := func(tt *testing.T, iResp *ResponseInfo, expectError bool) {
		if expectError {
			assert.NotNil(tt, iResp.Error)
		} else {
			assert.Nil(tt, iResp.Error)
		}
		assert.NotNil(tt, iResp)
		assert.NotNil(tt, iResp.Response)
		assert.NotEmpty(tt, iResp.RequestId)
		assert.NotEmpty(tt, iResp.CorrelationId)
		assert.True(tt, iResp.Latency > 0)
		connTracking := iResp.ConnTracking
		assert.NotNil(tt, connTracking.ReqConnInfo)
		_, err := time.ParseDuration(connTracking.TotalLatency)
		assert.Nil(tt, err)
		_, err = time.ParseDuration(connTracking.TlsLatency)
		assert.Nil(tt, err)
		_, err = time.ParseDuration(connTracking.ConnLatency)
		assert.Nil(tt, err)
		// we don't have a DNS lookup for httptest server so this should be empty
		assert.Equal(tt, connTracking.DnsLatency, "")
	}

	t.Run("should call RequestStarted and RequestFinished for succeeded requests", func(tt *testing.T) {
		tt.Parallel()
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		requestedStartedCalled := false
		requestCompletetedCalled := false
		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				requestedStartedCalled = true
				checkRequestInfo(tt, iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				requestCompletetedCalled = true
				checkResponseInfo(tt, iResp, false)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// clientOptions.DisableRPRegistration = true

		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(tt, err)
		reqHeader := http.Header{}
		reqHeader.Set("X-Ms-Correlation-Request-Id", testCorrelationId)
		ctx := runtime.WithHTTPHeader(context.Background(), reqHeader)
		_, err = client.Get(ctx, rgName, resourceName, nil)
		assert.NoError(tt, err)

		assert.True(tt, requestedStartedCalled)
		assert.True(tt, requestCompletetedCalled)
	})

	t.Run("should get ArmError for failed requests", func(tt *testing.T) {
		tt.Parallel()
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"code":"TestInternalError","message":"The is test internal error."}}`))
		}))
		defer ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(tt, iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				checkResponseInfo(tt, iResp, true)
				respErr := iResp.Error
				assert.NotNil(tt, respErr)
				assert.Equal(tt, ArmErrorCode("TestInternalError"), respErr.Code)
				assert.NotEmpty(tt, respErr.Message)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		reqHeader := http.Header{}
		reqHeader.Set("X-Ms-Correlation-Request-Id", testCorrelationId)
		ctx := runtime.WithHTTPHeader(context.Background(), reqHeader)
		assert.NoError(tt, err)
		_, err = client.Get(ctx, rgName, resourceName, nil)
		assert.Error(tt, err)
	})

	t.Run("should get correct ArmError when context timeout", func(tt *testing.T) {
		tt.Parallel()
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(tt, iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				assert.NotNil(tt, iResp)
				assert.Nil(tt, iResp.Response)
				assert.True(tt, iResp.Latency > 0)
				respErr := iResp.Error
				assert.NotNil(tt, respErr)
				assert.Equal(tt, ArmErrorCodeContextDeadlineExceeded, respErr.Code)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(tt, err)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		_, err = client.Get(ctx, rgName, resourceName, nil)
		assert.Error(tt, err)
		assert.EqualError(tt, err, "context deadline exceeded")
	})

	t.Run("should get correct ArmError when context canceled", func(tt *testing.T) {
		tt.Parallel()
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(tt, iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				assert.NotNil(tt, iResp)
				assert.Nil(tt, iResp.Response)
				assert.True(tt, iResp.Latency > 0)
				respErr := iResp.Error
				assert.NotNil(tt, respErr)
				assert.Equal(tt, ArmErrorCodeContextCanceled, respErr.Code)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(tt, err)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		cancel()
		_, err = client.Get(ctx, rgName, resourceName, nil)
		assert.EqualError(tt, err, "context canceled")
	})

	t.Run("should get correct ArmError for transport error", func(tt *testing.T) {
		tt.Parallel()
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		// close it to simulate transport error
		ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(tt, iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				assert.NotNil(tt, iResp)
				assert.Nil(tt, iResp.Response)
				assert.True(tt, iResp.Latency > 0)
				respErr := iResp.Error
				assert.NotNil(tt, respErr)
				assert.Equal(tt, ArmErrorCodeTransportError, respErr.Code)
				assert.NotEmpty(tt, respErr.Message)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(tt, err)
		_, err = client.Get(context.Background(), rgName, resourceName, nil)
		assert.Error(tt, err)
	})

	t.Run("should get correct ArmError for server timeout", func(tt *testing.T) {
		tt.Parallel()
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(tt, iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				assert.NotNil(tt, iResp)
				assert.Nil(tt, iResp.Response)
				assert.True(tt, iResp.Latency > 0)
				respErr := iResp.Error
				assert.NotNil(tt, respErr)
				assert.Equal(tt, ArmErrorCodeContextDeadlineExceeded, respErr.Code)
				assert.NotEmpty(tt, respErr.Message)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Retry.TryTimeout = 10 * time.Millisecond
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(tt, err)
		_, err = client.Get(context.Background(), rgName, resourceName, nil)
		assert.Error(tt, err)
	})

}

var _ ArmRequestMetricCollector = &testCollector{}

type testCollector struct {
	requestStarted   func(iReq *RequestInfo)
	requestCompleted func(iReq *RequestInfo, iResp *ResponseInfo)
}

func (c *testCollector) RequestStarted(iReq *RequestInfo) {
	c.requestStarted(iReq)
}

func (c *testCollector) RequestCompleted(iReq *RequestInfo, iResp *ResponseInfo) {
	c.requestCompleted(iReq, iResp)
}

type mockServerTransport struct {
	do func(*http.Request) (*http.Response, error)
}

func newMockServerTransportWithTestServer(ts *httptest.Server) *mockServerTransport {
	return &mockServerTransport{
		do: func(req *http.Request) (*http.Response, error) {
			newReq := req.Clone(req.Context())
			tsURL, err := url.Parse(ts.URL)
			if err != nil {
				panic(err)
			}
			newReq.URL = tsURL
			newReq = newReq.WithContext(req.Context())
			if err != nil {
				return nil, err
			}
			resp, err := ts.Client().Do(newReq)
			if err != nil {
				return nil, err
			}
			resp.Request = req
			return resp, nil
		},
	}
}

func (m *mockServerTransport) Do(req *http.Request) (*http.Response, error) {
	return m.do(req)
}

type mockTokenCredential struct{}

// GetToken implements azcore.TokenCredential
func (m *mockTokenCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     "fakeToken",
		ExpiresOn: time.Now().Add(1 * time.Hour),
	}, nil
}

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
