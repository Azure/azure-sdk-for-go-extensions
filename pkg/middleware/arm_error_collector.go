package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// ArmError is unified Error Experience across AzureResourceManager, it contains Code Message.
type ArmError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

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
}

func newResponseInfo(resp *http.Response, err *ArmError, latency time.Duration) *ResponseInfo {
	return &ResponseInfo{Response: resp, Error: err, Latency: latency}
}

// ArmRequestMetricCollector is a interface that collectors need to implement.
// TODO: use *policy.Request or *http.Request?
type ArmRequestMetricCollector interface {
	// RequestStarted is called when a request is about to be sent.
	// context is not provided, get it from Request.Context()
	RequestStarted(*RequestInfo)

	// RequestCompleted is called when a request is finished (statusCode < 400)
	// context is not provided, get it from Request.Context()
	RequestCompleted(*RequestInfo, *ResponseInfo)

	// RequestFailed is called when a request is failed (statusCode > 399)
	// context is not provided, get it from Request.Context()
	RequestFailed(*RequestInfo, *ResponseInfo)
}

// ArmRequestMetricPolicy is a policy that collects metrics/telemetry for ARM requests.
type ArmRequestMetricPolicy struct {
	Collector ArmRequestMetricCollector
}

// Do implements the azcore/policy.Policy interface.
func (p *ArmRequestMetricPolicy) Do(req *policy.Request) (*http.Response, error) {
	reqRaw := req.Raw()
	if reqRaw == nil || reqRaw.URL == nil {
		// not able to collect telemetry, just pass through
		return req.Next()
	}

	armResId, err := arm.ParseResourceID(reqRaw.URL.Path)
	if err != nil {
		// TODO: error handling without break the request.
	}

	started := time.Now()
	p.requestStarted(newRequestInfo(reqRaw, armResId))
	resp, err := req.Next()
	latency := time.Since(started)
	if err != nil {
		// either it's a transport error
		// or it is already handled by previous policy
		// TODO: distinguash
		// - Context Cancelled (request configured context to have timeout)
		// - ClientTimeout (context still valid, http client have timeout configured)
		// - Transport Error (DNS/Dail/TLS/ServerTimeout)
		p.requestFailed(newRequestInfo(reqRaw, armResId), newResponseInfo(resp, &ArmError{Code: "TransportError", Message: err.Error()}, latency))
		return resp, err
	}

	if resp == nil {
		p.requestFailed(newRequestInfo(reqRaw, armResId), newResponseInfo(resp, &ArmError{Code: "UnexpectedTransportorBehavior", Message: "transport return nil, nil"}, latency))
		return resp, nil
	}

	if resp != nil && resp.StatusCode > 399 {
		// for 4xx, 5xx response, ARM should include {error:{code, message}} in the body
		err := runtime.NewResponseError(resp)
		respErr := &azcore.ResponseError{}
		if errors.As(err, &respErr) {
			p.requestFailed(newRequestInfo(reqRaw, armResId), newResponseInfo(resp, &ArmError{Code: respErr.ErrorCode, Message: ""}, latency))
		} else {
			p.requestFailed(newRequestInfo(reqRaw, armResId), newResponseInfo(resp, &ArmError{Code: "NotAnArmError", Message: "Response body is not in ARM error form: {error:{code, message}}"}, latency))
		}

		// just an observer, caller/client have responder to handle application error.
		return resp, nil
	}

	p.requestCompleted(newRequestInfo(reqRaw, armResId), newResponseInfo(resp, nil, latency))
	return resp, nil
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

// shortcut function to handle nil collector
func (p *ArmRequestMetricPolicy) requestFailed(iReq *RequestInfo, iResp *ResponseInfo) {
	if p.Collector != nil {
		p.Collector.RequestFailed(iReq, iResp)
	}
}
