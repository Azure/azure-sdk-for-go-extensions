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

// creates test cases for simple error code comparisons
func createSimpleErrorCodeTests[T error](errorCode string, description string, createErr func(string, int, string) T) []testCase {
	return createMessageContainsTests(errorCode, http.StatusBadRequest, "irrelevant message", description, createErr)
}

// creates test cases for errors that depend on both error code and message content
func createMessageContainsTests[T error](errorCode string, statusCode int, message string, description string, createErr func(string, int, string) T) []testCase {
	return []testCase{
		{
			description:   description,
			responseError: createErr(errorCode, statusCode, message),
			expected:      true,
		},
		{
			description:   "Different Error Code",
			responseError: createErr("nooo im not found", http.StatusNotFound, ""),
			expected:      false,
		},
	}
}

func checkErrors(t *testing.T, testName string, testCases []testCase, testFunc errorTestFunc) {
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := testFunc(tc.responseError)
			if got != tc.expected {
				t.Errorf("%s() = %t, want %t for %s", testName, got, tc.expected, tc.description)
			}
		})
	}
}

// Basic Response Error Tests
func TestIsNotFoundErr(t *testing.T) {
	err1 := &azcore.ResponseError{StatusCode: http.StatusNotFound}
	assert.Equal(t, IsNotFoundErr(err1), true)
	err2 := &azcore.ResponseError{StatusCode: http.StatusOK}
	assert.Equal(t, IsNotFoundErr(err2), false)
	err3 := &azcore.ResponseError{StatusCode: http.StatusBadRequest}
	assert.Equal(t, IsNotFoundErr(err3), false)
	err4 := &azcore.ResponseError{StatusCode: http.StatusInternalServerError}
	assert.Equal(t, IsNotFoundErr(err4), false)
	assert.Equal(t, IsNotFoundErr(nil), false)
}

// Azure Allocation Error Tests
func TestZonalAllocationFailureOccurred(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		ZoneAllocationFailed,
		"Zonal Allocation Failed",
		createResponseError,
	)
	checkErrors(t, "ZonalAllocationFailureOccurred", testCases, ZonalAllocationFailureOccurred)
}

func TestAllocationFailureOccurred(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		AllocationFailed,
		"Allocation Failed",
		createResponseError,
	)
	checkErrors(t, "AllocationFailureOccurred", testCases, AllocationFailureOccurred)
}

func TestOverConstrainedAllocationFailureOccurred(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		OverconstrainedAllocationRequest,
		"Overconstrained Allocation Failed",
		createResponseError,
	)
	checkErrors(t, "OverconstrainedAllocationFailureOccurred", testCases, OverconstrainedAllocationFailureOccurred)
}

func TestOverConstrainedZonalAllocationFailureOccurred(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		OverconstrainedZonalAllocationRequest,
		"Overconstrained Zonal Allocation Failed",
		createResponseError,
	)
	checkErrors(t, "OverconstrainedZonalAllocationFailureOccurred", testCases, OverconstrainedZonalAllocationFailureOccurred)
}

// Azure Quota Error Tests
func TestSKUFamilyQuotaHasBeenReached(t *testing.T) {
	testCases := createMessageContainsTests(
		OperationNotAllowed,
		http.StatusForbidden,
		"Family Cores quota exceeded",
		"Quota Exceeded",
		createResponseError,
	)
	checkErrors(t, "SKUFamilyQuotaHasBeenReached", testCases, SKUFamilyQuotaHasBeenReached)
}

func TestRegionalQuotaHasBeenReached(t *testing.T) {
	testCases := createMessageContainsTests(
		OperationNotAllowed,
		http.StatusForbidden,
		"exceeding approved Total Regional Cores quota",
		"Regional Quota Exceeded",
		createResponseError,
	)
	checkErrors(t, "RegionalQuotaHasBeenReached", testCases, RegionalQuotaHasBeenReached)
}

func TestLowPriorityQuotaHasBeenReached(t *testing.T) {
	testCases := createMessageContainsTests(
		OperationNotAllowed,
		http.StatusForbidden,
		"Operation could not be completed as it results in exceeding approved LowPriorityCores quota",
		"LowPriority Quota Exceeded",
		createResponseError,
	)
	checkErrors(t, "LowPriorityQuotaHasBeenReached", testCases, LowPriorityQuotaHasBeenReached)
}

// Azure Resource Error Tests
func TestIsNicReservedForAnotherVM(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		NicReservedForAnotherVM,
		"NIC Reserved for Another VM",
		createResponseError,
	)
	checkErrors(t, "IsNicReservedForAnotherVM", testCases, IsNicReservedForAnotherVM)
}

func TestIsSKUNotAvailable(t *testing.T) {
	testCases := createSimpleErrorCodeTests(
		SKUNotAvailableErrorCode,
		"SKU Not Available",
		createResponseError,
	)
	checkErrors(t, "IsSKUNotAvailable", testCases, IsSKUNotAvailable)
}
