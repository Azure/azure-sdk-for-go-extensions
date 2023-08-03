package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
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

	checkRequestInfo := func(iReq *RequestInfo) {
		assert.NotNil(t, iReq)
		assert.Equal(t, iReq.ArmResId.SubscriptionID, subID)
		assert.Equal(t, iReq.ArmResId.ResourceGroupName, rgName)
		assert.Equal(t, iReq.ArmResId.Name, resourceName)
		assert.NotNil(t, iReq.Request)
	}

	checkResponseInfo := func(iResp *ResponseInfo, expectError bool) {
		if expectError {
			assert.NotNil(t, iResp.Error)
		} else {
			assert.Nil(t, iResp.Error)
		}
		assert.NotNil(t, iResp)
		assert.NotNil(t, iResp.Response)
		assert.NotEmpty(t, iResp.RequestId)
		assert.NotEmpty(t, iResp.CorrelationId)
		assert.True(t, iResp.Latency > 0)
		connTracking := iResp.ConnTracking
		assert.NotNil(t, connTracking.ReqConnInfo)
		_, err := time.ParseDuration(connTracking.TotalLatency)
		assert.Nil(t, err)
		_, err = time.ParseDuration(connTracking.TlsLatency)
		assert.Nil(t, err)
		_, err = time.ParseDuration(connTracking.ConnLatency)
		assert.Nil(t, err)
		// we don't have a DNS lookup for httptest server so this should be empty
		assert.Equal(t, connTracking.DnsLatency, "")
	}

	t.Run("should call RequestStarted and RequestFinished for succeeded requests", func(t *testing.T) {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		requestedStartedCalled := false
		requestCompletetedCalled := false
		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				requestedStartedCalled = true
				checkRequestInfo(iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				requestCompletetedCalled = true
				checkResponseInfo(iResp, false)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// clientOptions.DisableRPRegistration = true

		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(t, err)
		reqHeader := http.Header{}
		reqHeader.Set("X-Ms-Correlation-Request-Id", testCorrelationId)
		ctx := runtime.WithHTTPHeader(context.Background(), reqHeader)
		_, err = client.Get(ctx, rgName, resourceName, nil)
		assert.NoError(t, err)

		assert.True(t, requestedStartedCalled)
		assert.True(t, requestCompletetedCalled)
	})

	t.Run("should get ArmError for failed requests", func(t *testing.T) {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"code":"TestInternalError","message":"The is test internal error."}}`))
		}))
		defer ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				checkResponseInfo(iResp, true)
				respErr := iResp.Error
				assert.NotNil(t, respErr)
				assert.Equal(t, respErr.Code, "TestInternalError")
				assert.NotEmpty(t, respErr.Message)
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
		assert.NoError(t, err)
		_, err = client.Get(ctx, rgName, resourceName, nil)
		assert.Error(t, err)
	})

	t.Run("should get correct ArmError when context timeout", func(t *testing.T) {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				assert.NotNil(t, iResp)
				assert.Nil(t, iResp.Response)
				assert.True(t, iResp.Latency > 0)
				respErr := iResp.Error
				assert.NotNil(t, respErr)
				assert.Equal(t, "ContextTimeout", respErr.Code)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		_, err = client.Get(ctx, rgName, resourceName, nil)
		assert.Error(t, err)
		assert.EqualError(t, err, "context deadline exceeded")
	})

	t.Run("should get correct ArmError when context canceled", func(t *testing.T) {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				assert.NotNil(t, iResp)
				assert.Nil(t, iResp.Response)
				assert.True(t, iResp.Latency > 0)
				respErr := iResp.Error
				assert.NotNil(t, respErr)
				assert.Equal(t, "ContextCanceled", respErr.Code)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		cancel()
		_, err = client.Get(ctx, rgName, resourceName, nil)
		assert.EqualError(t, err, "context canceled")
	})

	t.Run("should get correct ArmError for transport error", func(t *testing.T) {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		// close it to simulate transport error
		ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				assert.NotNil(t, iResp)
				assert.Nil(t, iResp.Response)
				assert.True(t, iResp.Latency > 0)
				respErr := iResp.Error
				assert.NotNil(t, respErr)
				assert.Equal(t, "TransportError", respErr.Code)
				assert.NotEmpty(t, respErr.Message)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(t, err)
		_, err = client.Get(context.Background(), rgName, resourceName, nil)
		assert.Error(t, err)
	})

	t.Run("should get correct ArmError for server timeout", func(t *testing.T) {
		ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(1 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		collector := &testCollector{
			requestStarted: func(iReq *RequestInfo) {
				checkRequestInfo(iReq)
			},
			requestCompleted: func(iReq *RequestInfo, iResp *ResponseInfo) {
				assert.NotNil(t, iResp)
				assert.Nil(t, iResp.Response)
				assert.True(t, iResp.Latency > 0)
				respErr := iResp.Error
				assert.NotNil(t, respErr)
				assert.Equal(t, "ContextTimeout", respErr.Code)
				assert.NotEmpty(t, respErr.Message)
			},
		}

		clientOptions := DefaultArmOpts("testUserAgent", collector)
		// no retry
		clientOptions.Retry.MaxRetries = -1
		clientOptions.Retry.TryTimeout = 10 * time.Millisecond
		clientOptions.Transport = newMockServerTransportWithTestServer(ts)
		client, err := armcontainerservice.NewManagedClustersClient(subID, &mockTokenCredential{}, clientOptions)
		assert.NoError(t, err)
		_, err = client.Get(context.Background(), rgName, resourceName, nil)
		assert.Error(t, err)
	})

	// _, err = client.Get(context.Background(), "test", "test", nil)
	// here the error is parsed from response body twice
	// 1. by ArmRequestMetricPolicy, parse and log, and throw away.
	// 2. by generated client function: runtime.HasStatusCode, and return to here.
	//assert.Error(t, err)

	//respErr := &azcore.ResponseError{}
	//assert.True(t, errors.As(err, &respErr))
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
