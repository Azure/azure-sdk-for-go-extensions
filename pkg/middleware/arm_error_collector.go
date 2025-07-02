package middleware

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

const (
	headerKeyRequestID                                    = "X-Ms-Client-Request-Id"
	headerKeyCorrelationID                                = "X-Ms-Correlation-Request-id"
	ArmErrorCodeCastToArmResponseErrorFailed ArmErrorCode = "CastToArmResponseErrorFailed"
	ArmErrorCodeTransportError               ArmErrorCode = "TransportError"
	ArmErrorCodeUnexpectedTransportError     ArmErrorCode = "UnexpectedTransportError"
	ArmErrorCodeContextCanceled              ArmErrorCode = "ContextCanceled"
	ArmErrorCodeContextDeadlineExceeded      ArmErrorCode = "ContextDeadlineExceeded"
)

// ArmError is unified Error Experience across AzureResourceManager, it contains Code Message.
type ArmError struct {
	Code    ArmErrorCode `json:"code"`
	Message string       `json:"message"`
}

type ArmErrorCode string

type RequestInfo struct {
	Request  *http.Request
	ArmResId *arm.ResourceID
}

func newRequestInfo(req *http.Request, resId *arm.ResourceID) *RequestInfo {
	return &RequestInfo{Request: req, ArmResId: resId}
}

type ResponseInfo struct {
	Response      *http.Response
	Error         *ArmError
	Latency       time.Duration
	RequestId     string
	CorrelationId string
	ConnTracking  *HttpConnTracking
}

type HttpConnTracking struct {
	mu sync.RWMutex
	// Thread-safe access to these fields is provided via getter methods.
	// Direct field access may not be thread-safe during concurrent HTTP operations.
	TotalLatency string
	DnsLatency   string
	ConnLatency  string
	TlsLatency   string
	Protocol     string
	ReqConnInfo  *httptrace.GotConnInfo
}

// GetTotalLatency returns the total latency in a thread-safe manner
func (h *HttpConnTracking) GetTotalLatency() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.TotalLatency
}

// GetDnsLatency returns the DNS latency in a thread-safe manner
func (h *HttpConnTracking) GetDnsLatency() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.DnsLatency
}

// GetConnLatency returns the connection latency in a thread-safe manner
func (h *HttpConnTracking) GetConnLatency() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ConnLatency
}

// GetTlsLatency returns the TLS latency in a thread-safe manner
func (h *HttpConnTracking) GetTlsLatency() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.TlsLatency
}

// GetProtocol returns the negotiated protocol in a thread-safe manner
func (h *HttpConnTracking) GetProtocol() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.Protocol
}

// GetReqConnInfo returns the connection info in a thread-safe manner
func (h *HttpConnTracking) GetReqConnInfo() *httptrace.GotConnInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ReqConnInfo
}

func (h *HttpConnTracking) setTotalLatency(latency string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.TotalLatency = latency
}

func (h *HttpConnTracking) setDnsLatency(latency string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.DnsLatency = latency
}

func (h *HttpConnTracking) setConnLatency(latency string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ConnLatency = latency
}

func (h *HttpConnTracking) setTlsLatency(latency string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.TlsLatency = latency
}

func (h *HttpConnTracking) setProtocol(protocol string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Protocol = protocol
}

func (h *HttpConnTracking) setReqConnInfo(info *httptrace.GotConnInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ReqConnInfo = info
}

// ArmRequestMetricCollector is a interface that collectors need to implement.
// TODO: use *policy.Request or *http.Request?
type ArmRequestMetricCollector interface {
	// RequestStarted is called when a request is about to be sent.
	// context is not provided, get it from RequestInfo.Request.Context()
	RequestStarted(*RequestInfo)

	// RequestCompleted is called when a request is finished
	// context is not provided, get it from RequestInfo.Request.Context()
	// if an error occurred, ResponseInfo.Error will be populated
	RequestCompleted(*RequestInfo, *ResponseInfo)
}

// ArmRequestMetricPolicy is a policy that collects metrics/telemetry for ARM requests.
type ArmRequestMetricPolicy struct {
	Collector ArmRequestMetricCollector
}

// Do implements the azcore/policy.Policy interface.
func (p *ArmRequestMetricPolicy) Do(req *policy.Request) (*http.Response, error) {
	httpReq := req.Raw()
	if httpReq == nil || httpReq.URL == nil {
		// not able to collect telemetry, just pass through
		return req.Next()
	}

	armResId, err := arm.ParseResourceID(httpReq.URL.Path)
	if err != nil {
		// TODO: error handling without break the request.
	}

	connTracking := &HttpConnTracking{}
	// have to add to the context at first - then clone the policy.Request struct
	// this allows the connection tracing to happen
	// otherwise we can't change the underlying http request of req, we have to use
	// newARMReq
	newCtx := addConnectionTracingToRequestContext(httpReq.Context(), connTracking)
	newARMReq := req.Clone(newCtx)
	requestInfo := newRequestInfo(httpReq, armResId)
	started := time.Now()

	p.requestStarted(requestInfo)

	var resp *http.Response
	var reqErr error

	// defer this function in case there's a panic somewhere down the pipeline.
	// It's the calling user's responsibility to handle panics, not this policy
	defer func() {
		latency := time.Since(started)
		respInfo := &ResponseInfo{
			Response:     resp,
			Latency:      latency,
			ConnTracking: connTracking,
		}

		if reqErr != nil {
			// either it's a transport error
			// or it is already handled by previous policy
			respInfo.Error = parseTransportError(reqErr)
		} else {
			respInfo.Error = parseArmErrorFromResponse(resp)
		}

		// need to get the request id and correlation id from the response.request header
		// because the headers were added by policy and might be called after this policy
		if resp != nil && resp.Request != nil {
			respInfo.RequestId = resp.Request.Header.Get(headerKeyRequestID)
			respInfo.CorrelationId = resp.Request.Header.Get(headerKeyCorrelationID)
		}

		p.requestCompleted(requestInfo, respInfo)
	}()

	resp, reqErr = newARMReq.Next()
	return resp, reqErr
}

// shortcut function to handle nil collector
func (p *ArmRequestMetricPolicy) requestStarted(iReq *RequestInfo) {
	if p.Collector != nil {
		p.Collector.RequestStarted(iReq)
	}
}

// shortcut function to handle nil collector
func (p *ArmRequestMetricPolicy) requestCompleted(iReq *RequestInfo, iResp *ResponseInfo) {
	if p.Collector != nil {
		p.Collector.RequestCompleted(iReq, iResp)
	}
}

func parseArmErrorFromResponse(resp *http.Response) *ArmError {
	if resp == nil {
		return &ArmError{Code: ArmErrorCodeUnexpectedTransportError, Message: "nil response"}
	}
	if resp.StatusCode > 399 {
		// for 4xx, 5xx response, ARM should include {error:{code, message}} in the body
		err := runtime.NewResponseError(resp)
		respErr := &azcore.ResponseError{}
		if errors.As(err, &respErr) {
			return &ArmError{Code: ArmErrorCode(respErr.ErrorCode), Message: respErr.Error()}
		}
		return &ArmError{Code: ArmErrorCodeCastToArmResponseErrorFailed, Message: fmt.Sprintf("Response body is not in ARM error form: {error:{code, message}}: %s", err.Error())}
	}
	return nil
}

// distinguash
// - Context Cancelled (request configured context to have timeout)
// - ClientTimeout (context still valid, http client have timeout configured)
// - Transport Error (DNS/Dial/TLS/ServerTimeout)
func parseTransportError(err error) *ArmError {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) {
		return &ArmError{Code: ArmErrorCodeContextCanceled, Message: err.Error()}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &ArmError{Code: ArmErrorCodeContextDeadlineExceeded, Message: err.Error()}
	}
	return &ArmError{Code: ArmErrorCodeTransportError, Message: err.Error()}
}

func addConnectionTracingToRequestContext(ctx context.Context, connTracking *HttpConnTracking) context.Context {
	var getConn, dnsStart, connStart, tlsStart *time.Time

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			getConn = to.Ptr(time.Now())
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			if getConn != nil {
				connTracking.setTotalLatency(fmt.Sprintf("%dms", time.Now().Sub(*getConn).Milliseconds()))
			}

			connTracking.setReqConnInfo(&connInfo)
		},
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = to.Ptr(time.Now())
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			if dnsInfo.Err == nil {
				if dnsStart != nil {
					connTracking.setDnsLatency(fmt.Sprintf("%dms", time.Now().Sub(*dnsStart).Milliseconds()))
				}
			} else {
				connTracking.setDnsLatency(dnsInfo.Err.Error())
			}
		},
		ConnectStart: func(_, _ string) {
			connStart = to.Ptr(time.Now())
		},
		ConnectDone: func(_, _ string, err error) {
			if err == nil {
				if connStart != nil {
					connTracking.setConnLatency(fmt.Sprintf("%dms", time.Now().Sub(*connStart).Milliseconds()))
				}
			} else {
				connTracking.setConnLatency(err.Error())
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = to.Ptr(time.Now())
		},
		TLSHandshakeDone: func(t tls.ConnectionState, err error) {
			if err == nil {
				if tlsStart != nil {
					connTracking.setTlsLatency(fmt.Sprintf("%dms", time.Now().Sub(*tlsStart).Milliseconds()))
				}
				connTracking.setProtocol(t.NegotiatedProtocol)
			} else {
				connTracking.setTlsLatency(err.Error())
			}
		},
	}
	ctx = httptrace.WithClientTrace(ctx, trace)
	return ctx
}
