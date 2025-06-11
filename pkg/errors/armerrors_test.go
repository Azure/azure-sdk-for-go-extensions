package errors

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

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
	description   string
	responseError error
	expected      bool
}

type errorTestFunc func(error) bool

func createResponseError(errorCode string, statusCode int, errorMessage string) *azcore.ResponseError {
	errorBody := fmt.Sprintf(`{"error": {"code": "%s", "message": "%s"}}`, errorCode, errorMessage)
	return &azcore.ResponseError{
		ErrorCode:  errorCode,
		StatusCode: statusCode,
		RawResponse: &http.Response{
			Body: io.NopCloser(strings.NewReader(errorBody)),
		},
	}
}

func runErrorTests(t *testing.T, testName string, testCases []testCase, testFunc errorTestFunc) {
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := testFunc(tc.responseError)
			if got != tc.expected {
				t.Errorf("%s() = %t, want %t for %s", testName, got, tc.expected, tc.description)
			}
		})
	}
}

// creates test cases for simple error code comparisons
func createSimpleErrorCodeTests(errorCode string, description string) []testCase {
	return []testCase{
		{
			description:   description,
			responseError: createResponseError(errorCode, http.StatusBadRequest, "Some error message"),
			expected:      true,
		},
		{
			description:   "Different Error Code",
			responseError: createResponseError(ResourceNotFound, http.StatusNotFound, ""),
			expected:      false,
		},
	}
}

// creates test cases for errors that depend on both error code and message content
func createMessageContainsTests(errorCode string, statusCode int, message string, description string) []testCase {
	return []testCase{
		{
			description:   description,
			responseError: createResponseError(errorCode, statusCode, message),
			expected:      true,
		},
		{
			description:   "Different Error Code",
			responseError: createResponseError(ResourceNotFound, http.StatusNotFound, ""),
			expected:      false,
		},
	}
}

func TestSKUFamilyQuotaHasBeenReached(t *testing.T) {
	testCases := createMessageContainsTests(
		OperationNotAllowed,
		http.StatusForbidden,
		"Family Cores quota exceeded",
		"Quota Exceeded",
	)
	runErrorTests(t, "SKUFamilyQuotaHasBeenReached", testCases, SKUFamilyQuotaHasBeenReached)
}

func TestZonalAllocationFailureOccurred(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		ZoneAllocationFailed,
		"Zonal Allocation Failed",
	)
	runErrorTests(t, "ZonalAllocationFailureOccurred", testCases, ZonalAllocationFailureOccurred)
}

func TestAllocationFailureOccured(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		AllocationFailed,
		"Allocation Failed",
	)
	runErrorTests(t, "AllocationFailureOccurred", testCases, AllocationFailureOccurred)
}

func TestOverConstrainedAllocationFailureOccurred(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		OverconstrainedAllocationRequest,
		"Overconstrained Allocation Failed",
	)
	runErrorTests(t, "OverconstrainedAllocationFailureOccurred", testCases, OverconstrainedAllocationFailureOccurred)
}

func TestOverConstrainedZonalAllocationFailureOccurred(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		OverconstrainedZonalAllocationRequest,
		"Overconstrained Zonal Allocation Failed",
	)
	runErrorTests(t, "OverconstrainedZonalAllocationFailureOccurred", testCases, OverconstrainedZonalAllocationFailureOccurred)
}

func TestRegionalQuotaHasBeenReached(t *testing.T) {
	testCases := createMessageContainsTests(
		OperationNotAllowed,
		http.StatusForbidden,
		"exceeding approved Total Regional Cores quota",
		"Regional Quota Exceeded",
	)
	runErrorTests(t, "RegionalQuotaHasBeenReached", testCases, RegionalQuotaHasBeenReached)
}

func TestIsNicReservedForAnotherVM(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		NicReservedForAnotherVM,
		"NIC Reserved for Another VM",
	)
	runErrorTests(t, "IsNicReservedForAnotherVM", testCases, IsNicReservedForAnotherVM)
}

func TestIsSKUNotAvailable(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		SKUNotAvailableErrorCode,
		"SKU Not Available",
	)
	runErrorTests(t, "IsSKUNotAvailable", testCases, IsSKUNotAvailable)
}

func TestLowPriorityQuotaHasBeenReached(t *testing.T) {
	testCases := createMessageContainsTests(
		OperationNotAllowed,
		http.StatusForbidden,
		"Operation could not be completed as it results in exceeding approved LowPriorityCores quota",
		"LowPriority Quota Exceeded",
	)
	runErrorTests(t, "LowPriorityQuotaHasBeenReached", testCases, LowPriorityQuotaHasBeenReached)
}
