package errors


import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"testing"

    "bytes"
    "io"
    "net/http"

)

func TestIsNotFoundErr(t *testing.T) {
	err1 := &azcore.ResponseError{ErrorCode: ResourceNotFound}
	if !IsNotFoundErr(err1) {
		t.Errorf("Expected IsNotFoundErr to return true for err1, but it returned false")
	}

	err2 := &azcore.ResponseError{ErrorCode: "SomeOtherErrorCode"}
	if IsNotFoundErr(err2) {
		t.Errorf("Expected IsNotFoundErr to return false for err2, but it returned true")
	}

	if IsNotFoundErr(nil) {
		t.Errorf("Expected IsNotFoundErr to return false for nil input, but it returned true")
	}
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
                t.Errorf("SKUFamilyQuotaHasBeenReached() = %v, want %v", got, tc.expected)
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
                t.Errorf("ZonalAllocationFailureOccurred() = %v, want %v for %s", got, tc.expected, tc.description)
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
                t.Errorf("RegionalQuotaHasBeenReached() = %v, want %v for %s", got, tc.expected, tc.description)
            }
        })
    }
}

