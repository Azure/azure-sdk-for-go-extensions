package errors

import (
	"testing"
        "bytes"
	"io"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/stretchr/testify/assert"
)

func TestIsNotFoundErr(t *testing.T) {
	err1 := &azcore.ResponseError{ErrorCode: ResourceNotFound}
	assert.Equal(t, IsNotFoundErr(err1), true)
	err2 := &azcore.ResponseError{ErrorCode: "SomeOtherErrorCode"}
        assert.Equal(t, IsNotFoundErr(err2), false)
	assert.Equal(t, IsNotFoundErr(nil), false)
}



type testCase struct {
    description        string
    responseError      *azcore.ResponseError
    expected bool
}

func TestSKUFamilyQuotaHasBeenReached(t *testing.T) {
    testCases := []testCase{
        {
            description: "Quota Exceeded",
            responseError: &azcore.ResponseError{
                ErrorCode:   "OperationNotAllowed",
                StatusCode: http.StatusForbidden,
                RawResponse: &http.Response{
                    Body: io.NopCloser(bytes.NewReader([]byte(`{"error": {"code": "OperationNotAllowed", "message": "Family Cores quota exceeded"}}`))),
                },
            },
            expected: true,
        },
        {
            description: "Different Error Code",
            responseError: &azcore.ResponseError{
                ErrorCode:   "ResourceNotFound",
                StatusCode: http.StatusNotFound,
                RawResponse: &http.Response{
                    Body: io.NopCloser(bytes.NewReader([]byte(`{"error": {"code": "ResourceNotFound"}}`))),
                },
            },
            expected: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.description, func(t *testing.T) {
            if got := SKUFamilyQuotaHasBeenReached(tc.responseError); got != tc.expected {
                t.Errorf("SKUFamilyQuotaHasBeenReached() = %t, want %t", got, tc.expected)
            }
        })
    }
}


func TestZonalAllocationFailureOccurred(t *testing.T) {
    testCases := []testCase{
        {
            description: "Zonal Allocation Failed",
            responseError: &azcore.ResponseError{
                ErrorCode:   ZoneAllocationFailed,
                StatusCode: http.StatusBadRequest,
                RawResponse: &http.Response{
                    Body: io.NopCloser(bytes.NewReader([]byte(`{"error": {"code": "ZonalAllocationFailed", "message": "Failed to allocate resources in the zone"}}`))),
                },
            },
            expected: true,
        },
        {
            description: "Different Error Code",
            responseError: &azcore.ResponseError{
                ErrorCode:   "ResourceNotFound",
                StatusCode: http.StatusNotFound,
                RawResponse: &http.Response{
                    Body: io.NopCloser(bytes.NewReader([]byte(`{"error": {"code": "ResourceNotFound"}}`))),
                },
            },
            expected: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.description, func(t *testing.T) {
            if got := ZonalAllocationFailureOccurred(tc.responseError); got != tc.expected {
                t.Errorf("ZonalAllocationFailureOccurred() = %t, want %t for %s", got, tc.expected, tc.description)
            }
        })
    }
}




func TestRegionalQuotaHasBeenReached(t *testing.T) {
    testCases := []testCase{
        {
            description: "Regional Quota Exceeded",
            responseError: &azcore.ResponseError{
                ErrorCode:   OperationNotAllowed,
                StatusCode: http.StatusForbidden,
                RawResponse: &http.Response{
                    Body: io.NopCloser(bytes.NewReader([]byte(`{"error": {"code": "OperationNotAllowed", "message": "exceeding approved Total Regional Cores quota"}}`))),
                },
            },
            expected: true,
        },
        {
            description: "Different Error Code",
            responseError: &azcore.ResponseError{
                ErrorCode:   "ResourceNotFound",
                StatusCode: http.StatusNotFound,
                RawResponse: &http.Response{
                    Body: io.NopCloser(bytes.NewReader([]byte(`{"error": {"code": "ResourceNotFound"}}`))),
                },
            },
            expected: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.description, func(t *testing.T) {
            if got := RegionalQuotaHasBeenReached(tc.responseError); got != tc.expected {
                t.Errorf("RegionalQuotaHasBeenReached() = %t, want %t for %s", got, tc.expected, tc.description)
            }
        })
    }
}



func TestLowPriorityQuotaHasBeenReached(t *testing.T) {
    testCases := []testCase{
        {
            description: "LowPriority Quota Exceeded",
            responseError: &azcore.ResponseError{
                ErrorCode:   OperationNotAllowed,
                StatusCode: http.StatusForbidden,
                RawResponse: &http.Response{
                    Body: io.NopCloser(bytes.NewReader([]byte(`{"error": {"code": "OperationNotAllowed", "message": "Operation could not be completed as it results in exceeding approved LowPriorityCores quota"}}`))),
                },
            },
            expected: true,
        },
        {
            description: "Different Error Code",
            responseError: &azcore.ResponseError{
                ErrorCode:   "ResourceNotFound",
                StatusCode: http.StatusNotFound,
                RawResponse: &http.Response{
                    Body: io.NopCloser(bytes.NewReader([]byte(`{"error": {"code": "ResourceNotFound"}}`))),
                },
            },
            expected: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.description, func(t *testing.T) {
            if got := LowPriorityQuotaHasBeenReached(tc.responseError); got != tc.expected {
                t.Errorf("LowPriorityQuotaHasBeenReached() = %t, want %t for %s", got, tc.expected, tc.description)
            }
        })
    }
}

