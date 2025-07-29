package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

var errorMessageRegex = regexp.MustCompile(`"message"\s*:\s*("(?:[^"\\]|\\.)*")`)

var jsonUnescaper = strings.NewReplacer(
	"\n", " ",
	"\t", " ",
	"\r", " ",
)

type ResponseErrorWrapper struct {
	respErr *azcore.ResponseError
	message string
}

func NewResponseErrorWrapper(respErr *azcore.ResponseError) *ResponseErrorWrapper {
	return &ResponseErrorWrapper{
		respErr: respErr,
	}
}

func AsWrappedResponseError(err error) error {
	if azErr := IsResponseError(err); azErr != nil {
		return NewResponseErrorWrapper(azErr)
	}
	return err
}

func (c *ResponseErrorWrapper) Error() string {
	if c.message != "" {
		return c.message
	}

	if c.respErr == nil {
		// TODO - special handling if this is nil? But for now, just return empty string to not pollute logs
		return ""
	}

	// Attempt to build error message - this is best effort since format can vary depending on the Azure service
	c.message = buildWrapperErrorMessage(c.respErr)

	return c.message
}

func buildWrapperErrorMessage(respErr *azcore.ResponseError) string {
	httpCode := respErr.StatusCode
	errorCode := respErr.ErrorCode
	if errorCode == "" {
		errorCode = "UNAVAILABLE"
	}

	// Extract HTTP Method and URL
	httpMethod, url := extractRequestInfo(respErr)

	// Extract error message
	errorMessage := extractErrorMessage(respErr)

	wrapperMessage := fmt.Sprintf("HTTP CODE: %d, ERROR CODE: %s, MESSAGE: %s, REQUEST: %s %s",
		httpCode, errorCode, errorMessage, httpMethod, url)

	return wrapperMessage
}

// extractRequestInfo extracts HTTP method and URL with proper nil checks
func extractRequestInfo(respErr *azcore.ResponseError) (string, string) {
	method := "UNKNOWN"
	requestURL := "UNAVAILABLE"

	if respErr.RawResponse == nil || respErr.RawResponse.Request == nil {
		return method, requestURL
	}

	req := respErr.RawResponse.Request
	method = req.Method

	if req.URL != nil {
		requestURL = req.URL.String()
	}

	return method, requestURL
}

type AzureErrorResponse struct {
	Error AzureError `json:"error"`
}

type AzureError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []interface{} `json:"details"`
}

func extractErrorMessage(respErr *azcore.ResponseError) string {
	// these 2 cases shouldn't happen in real-world scenarios as a
	// response with no body should set it to http.NoBody
	if respErr.RawResponse == nil {
		return "UNAVAILABLE"
	}

	if respErr.RawResponse.Body == nil {
		return "UNAVAILABLE"
	}

	// Read the body content once and save it in case we need to use one of the fallback approaches for message extraction
	respBody := respErr.RawResponse.Body
	bodyBytes, err := io.ReadAll(respBody)
	if err != nil {
		return "UNAVAILABLE"
	}

	// First try parsing as wrapped format (with "error" wrapper)
	var wrappedResult AzureErrorResponse
	err = json.Unmarshal(bodyBytes, &wrappedResult)
	if err == nil && wrappedResult.Error.Code != "" {
		return jsonUnescaper.Replace(wrappedResult.Error.Message)
	}

	// Try parsing as unwrapped format (without "error" wrapper)
	var unwrappedResult AzureError
	err = json.Unmarshal(bodyBytes, &unwrappedResult)
	if err == nil && unwrappedResult.Code != "" {
		return jsonUnescaper.Replace(unwrappedResult.Message)
	}

	// If both JSON parsing attempts failed, fallback to regex extraction on the body
	matches := errorMessageRegex.FindStringSubmatch(string(bodyBytes))
	if len(matches) >= 2 {
		unquoted, err := strconv.Unquote(matches[1])
		if err != nil {
			// Fallback to raw message if unquoting fails
			return matches[1]
		}
		return jsonUnescaper.Replace(unquoted)
	}

	// If all attempts fail, return a generic unavailable message
	return "UNAVAILABLE"
}
