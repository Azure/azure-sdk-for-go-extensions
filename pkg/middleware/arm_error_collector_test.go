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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v7"
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
