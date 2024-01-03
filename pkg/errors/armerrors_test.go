package errors

import (
	"testing"
	"io"
        "strings"
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
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "OperationNotAllowed", "message": "Family Cores quota exceeded"}}`)),
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
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "ResourceNotFound"}}`)),
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
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "ZonalAllocationFailed", "message": "Failed to allocate resources in the zone"}}`)),
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
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "ResourceNotFound"}}`)),
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
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "OperationNotAllowed", "message": "exceeding approved Total Regional Cores quota"}}`)),
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
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "ResourceNotFound"}}`)),
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

func TestIsNicReservedForAnotherVM(t *testing.T) {
    testCases := []struct {
        description   string
        responseError error
        expected      bool
    }{
        {
            description: "NIC Reserved for Another VM",
            responseError: &azcore.ResponseError{
                ErrorCode:   NicReservedForAnotherVM,
                StatusCode:  http.StatusForbidden,
                RawResponse: &http.Response{
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "NicReservedForAnotherVm"}}`)),
                },
            },
            expected: true,
        },
        {
            description: "Different Error Code",
            responseError: &azcore.ResponseError{
                ErrorCode:   "ResourceNotFound",
                StatusCode:  http.StatusNotFound,
                RawResponse: &http.Response{
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "ResourceNotFound"}}`)),
                },
            },
            expected: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.description, func(t *testing.T) {
            if got := IsNicReservedForAnotherVM(tc.responseError); got != tc.expected {
                t.Errorf("IsNicReservedForAnotherVM() = %t, want %t for %s", got, tc.expected, tc.description)
            }
        })
    }
}

func TestIsSKUNotAvailable(t *testing.T) {
    testCases := []struct {
        description   string
        responseError error
        expected      bool
    }{
        {
            description: "SKU Not Available",
            responseError: &azcore.ResponseError{
                ErrorCode:   SKUNotAvailableErrorCode,
                StatusCode:  http.StatusForbidden,
                RawResponse: &http.Response{
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "SKUNotAvailable"}}`)),
                },
            },
            expected: true,
        },
        {
            description: "Different Error Code",
            responseError: &azcore.ResponseError{
                ErrorCode:   "ResourceNotFound",
                StatusCode:  http.StatusNotFound,
                RawResponse: &http.Response{
                    Body: io.NopCloser(strings.NewReader(`{"error": {"code": "ResourceNotFound"}}`)),
                },
            },
            expected: false,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.description, func(t *testing.T) {
            if got := IsSKUNotAvailable(tc.responseError); got != tc.expected {
                t.Errorf("IsSKUNotAvailable() = %t, want %t for %s", got, tc.expected, tc.description)
            }
        })
    }
}

