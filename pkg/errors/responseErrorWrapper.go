package errors

import (
	"fmt"
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

func AsWrappedResponseError(err error) *ResponseErrorWrapper {
	if azErr := IsResponseError(err); azErr != nil {
		return NewResponseErrorWrapper(azErr)
	}
	return nil
}

func (c *ResponseErrorWrapper) Error() string {
	if c.message != "" {
		return c.message
	}

	if c.respErr == nil {
		// TODO - special handling if this is nil? But for now, just return empty string to not pollute logs
		return ""
	}

	// Attempt to build error message
	// TODO - should we handle failures here in some special way? E.g. if we fail to extract with regex, do we just fallback to ResponseError.Error()?
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

// attempt to extract the error message from ResponseError using a regex.
// This is best effort based on the formats we encountered - if we fail to find a match, we return "UNAVAILABLE"
func extractErrorMessage(respErr *azcore.ResponseError) string {
	// Get the full error string which contains the JSON body
	fullError := respErr.Error()

	// Use regex to extract the message from the JSON in the error string
	// This pattern looks for "message": "..." in the JSON
	// See responseErrorWrapper_test.go for an example of the expected format
	matches := errorMessageRegex.FindStringSubmatch(fullError)

	if len(matches) < 1 {
		return "UNAVAILABLE"
	}

	unquoted, err := strconv.Unquote(matches[1])
	if err != nil {
		// Fallback to raw message if unquoting fails
		return matches[1]
	}

	return jsonUnescaper.Replace(unquoted)
}
