package errors

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v8"
)

type errorDetailTestCase struct {
	description string
	errorDetail armcontainerservice.ErrorDetail
	expected    bool
}

type errorDetailTestFunc func(armcontainerservice.ErrorDetail) bool

func createErrorDetail(errorCode string, errorMessage string) armcontainerservice.ErrorDetail {
	return armcontainerservice.ErrorDetail{
		Code:    &errorCode,
		Message: &errorMessage,
	}
}

// creates test cases for simple error code comparisons
func createSimpleErrorDetailCodeTests(errorCode string, description string) []errorDetailTestCase {
	return createErrorDetailMessageContainsTests(errorCode, "irrelevant message", description)
}

// creates test cases for errors that depend on both error code and message content
func createErrorDetailMessageContainsTests(errorCode string, message string, description string) []errorDetailTestCase {
	return []errorDetailTestCase{
		{
			description: description,
			errorDetail: createErrorDetail(errorCode, message),
			expected:    true,
		},
		{
			description: "Different Error Code",
			errorDetail: createErrorDetail("nooo im not found", ""),
			expected:    false,
		},
	}
}

func checkErrorDetails(t *testing.T, testName string, testCases []errorDetailTestCase, testFunc errorDetailTestFunc) {
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := testFunc(tc.errorDetail)
			if got != tc.expected {
				t.Errorf("%s() = %t, want %t for %s", testName, got, tc.expected, tc.description)
			}
		})
	}
}

func TestZonalAllocationFailureOccurredInErrorDetail(t *testing.T) {
	testCases := createSimpleErrorDetailCodeTests(
		ZoneAllocationFailed,
		"Zonal Allocation Failed",
	)
	checkErrorDetails(t, "ZonalAllocationFailureOccurredInErrorDetail", testCases, ZonalAllocationFailureOccurredInErrorDetail)
}

func TestAllocationFailureOccurredInErrorDetail(t *testing.T) {
	testCases := createSimpleErrorDetailCodeTests(
		AllocationFailed,
		"Allocation Failed",
	)
	checkErrorDetails(t, "AllocationFailureOccurredInErrorDetail", testCases, AllocationFailureOccurredInErrorDetail)
}

func TestOverConstrainedAllocationFailureOccurredInErrorDetail(t *testing.T) {
	testCases := createSimpleErrorDetailCodeTests(
		OverconstrainedAllocationRequest,
		"Overconstrained Allocation Failed",
	)
	checkErrorDetails(t, "OverconstrainedAllocationFailureOccurredInErrorDetail", testCases, OverconstrainedAllocationFailureOccurredInErrorDetail)
}

func TestOverConstrainedZonalAllocationFailureOccurredInErrorDetail(t *testing.T) {
	testCases := createSimpleErrorDetailCodeTests(
		OverconstrainedZonalAllocationRequest,
		"Overconstrained Zonal Allocation Failed",
	)
	checkErrorDetails(t, "OverconstrainedZonalAllocationFailureOccurredInErrorDetail", testCases, OverconstrainedZonalAllocationFailureOccurredInErrorDetail)
}

// Azure Quota Error Tests
func TestSKUFamilyQuotaHasBeenReachedInErrorDetail(t *testing.T) {
	testCases := createErrorDetailMessageContainsTests(
		OperationNotAllowed,
		"Family Cores quota exceeded",
		"Quota Exceeded",
	)
	checkErrorDetails(t, "SKUFamilyQuotaHasBeenReachedInErrorDetail", testCases, SKUFamilyQuotaHasBeenReachedInErrorDetail)
}

func TestSubscriptionQuotaHasBeenReachedInErrorDetail(t *testing.T) {
	testCases := createErrorDetailMessageContainsTests(
		OperationNotAllowed,
		"Submit a request for Quota increase",
		"Subscription Quota Exceeded",
	)
	checkErrorDetails(t, "SubscriptionQuotaHasBeenReachedInErrorDetail", testCases, SubscriptionQuotaHasBeenReachedInErrorDetail)
}

func TestRegionalQuotaHasBeenReachedInErrorDetail(t *testing.T) {
	testCases := createErrorDetailMessageContainsTests(
		OperationNotAllowed,
		"exceeding approved Total Regional Cores quota",
		"Regional Quota Exceeded",
	)
	checkErrorDetails(t, "RegionalQuotaHasBeenReachedInErrorDetail", testCases, RegionalQuotaHasBeenReachedInErrorDetail)
}

func TestLowPriorityQuotaHasBeenReachedInErrorDetail(t *testing.T) {
	testCases := createErrorDetailMessageContainsTests(
		OperationNotAllowed,
		"Operation could not be completed as it results in exceeding approved LowPriorityCores quota",
		"LowPriority Quota Exceeded",
	)
	checkErrorDetails(t, "LowPriorityQuotaHasBeenReachedInErrorDetail", testCases, LowPriorityQuotaHasBeenReachedInErrorDetail)
}

func TestIsNicReservedForAnotherVMInErrorDetail(t *testing.T) {
	testCases := createSimpleErrorDetailCodeTests(
		NicReservedForAnotherVM,
		"NIC Reserved for Another VM",
	)
	checkErrorDetails(t, "IsNicReservedForAnotherVMInErrorDetail", testCases, IsNicReservedForAnotherVMInErrorDetail)
}

func TestIsSKUNotAvailableInErrorDetail(t *testing.T) {
	testCases := createSimpleErrorDetailCodeTests(
		SKUNotAvailableErrorCode,
		"SKU Not Available",
	)
	checkErrorDetails(t, "IsSKUNotAvailableInErrorDetail", testCases, IsSKUNotAvailableInErrorDetail)
}

func TestIsInsufficientSubnetSizeDetails(t *testing.T) {
	testCases := createSimpleErrorDetailCodeTests(
		InsufficientSubnetSizeErrorCode,
		"Insufficient Subnet Size",
	)
	checkErrorDetails(t, "IsInsufficientSubnetSizeErrorDetails", testCases, IsInsufficientSubnetSizeErrorDetails)
}
