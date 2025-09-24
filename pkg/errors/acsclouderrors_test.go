package errors

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v7"
)

type cloudErrorTestCase struct {
	description string
	cloudError  armcontainerservice.CloudErrorBody
	expected    bool
}

type cloudErrorTestFunc func(armcontainerservice.CloudErrorBody) bool

func createCloudError(errorCode string, errorMessage string) armcontainerservice.CloudErrorBody {
	return armcontainerservice.CloudErrorBody{
		Code:    &errorCode,
		Message: &errorMessage,
	}
}

// creates test cases for simple error code comparisons
func createSimpleCloudErrorCodeTests(errorCode string, description string) []cloudErrorTestCase {
	return createCloudErrorMessageContainsTests(errorCode, "irrelevant message", description)
}

// creates test cases for errors that depend on both error code and message content
func createCloudErrorMessageContainsTests(errorCode string, message string, description string) []cloudErrorTestCase {
	return []cloudErrorTestCase{
		{
			description: description,
			cloudError:  createCloudError(errorCode, message),
			expected:    true,
		},
		{
			description: "Different Error Code",
			cloudError:  createCloudError("nooo im not found", ""),
			expected:    false,
		},
	}
}

func checkCloudErrors(t *testing.T, testName string, testCases []cloudErrorTestCase, testFunc cloudErrorTestFunc) {
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := testFunc(tc.cloudError)
			if got != tc.expected {
				t.Errorf("%s() = %t, want %t for %s", testName, got, tc.expected, tc.description)
			}
		})
	}
}

func TestZonalAllocationFailureOccurredInCloudError(t *testing.T) {
	testCases := createSimpleCloudErrorCodeTests(
		ZoneAllocationFailed,
		"Zonal Allocation Failed",
	)
	checkCloudErrors(t, "ZonalAllocationFailureOccurredInCloudError", testCases, ZonalAllocationFailureOccurredInCloudError)
}

func TestAllocationFailureOccurredInCloudError(t *testing.T) {
	testCases := createSimpleCloudErrorCodeTests(
		AllocationFailed,
		"Allocation Failed",
	)
	checkCloudErrors(t, "AllocationFailureOccurredInCloudError", testCases, AllocationFailureOccurredInCloudError)
}

func TestOverConstrainedAllocationFailureOccurredInCloudError(t *testing.T) {
	testCases := createSimpleCloudErrorCodeTests(
		OverconstrainedAllocationRequest,
		"Overconstrained Allocation Failed",
	)
	checkCloudErrors(t, "OverconstrainedAllocationFailureOccurredInCloudError", testCases, OverconstrainedAllocationFailureOccurredInCloudError)
}

func TestOverConstrainedZonalAllocationFailureOccurredInCloudError(t *testing.T) {
	testCases := createSimpleCloudErrorCodeTests(
		OverconstrainedZonalAllocationRequest,
		"Overconstrained Zonal Allocation Failed",
	)
	checkCloudErrors(t, "OverconstrainedZonalAllocationFailureOccurredInCloudError", testCases, OverconstrainedZonalAllocationFailureOccurredInCloudError)
}

// Azure Quota Error Tests
func TestSKUFamilyQuotaHasBeenReachedInCloudError(t *testing.T) {
	testCases := createCloudErrorMessageContainsTests(
		OperationNotAllowed,
		"Family Cores quota exceeded",
		"Quota Exceeded",
	)
	checkCloudErrors(t, "SKUFamilyQuotaHasBeenReachedInCloudError", testCases, SKUFamilyQuotaHasBeenReachedInCloudError)
}

func TestSubscriptionQuotaHasBeenReachedInCloudError(t *testing.T) {
	testCases := createCloudErrorMessageContainsTests(
		OperationNotAllowed,
		"Submit a request for Quota increase",
		"Subscription Quota Exceeded",
	)
	checkCloudErrors(t, "SubscriptionQuotaHasBeenReachedInCloudError", testCases, SubscriptionQuotaHasBeenReachedInCloudError)
}

func TestRegionalQuotaHasBeenReachedInCloudError(t *testing.T) {
	testCases := createCloudErrorMessageContainsTests(
		OperationNotAllowed,
		"exceeding approved Total Regional Cores quota",
		"Regional Quota Exceeded",
	)
	checkCloudErrors(t, "RegionalQuotaHasBeenReachedInCloudError", testCases, RegionalQuotaHasBeenReachedInCloudError)
}

func TestLowPriorityQuotaHasBeenReachedInCloudError(t *testing.T) {
	testCases := createCloudErrorMessageContainsTests(
		OperationNotAllowed,
		"Operation could not be completed as it results in exceeding approved LowPriorityCores quota",
		"LowPriority Quota Exceeded",
	)
	checkCloudErrors(t, "LowPriorityQuotaHasBeenReachedInCloudError", testCases, LowPriorityQuotaHasBeenReachedInCloudError)
}

func TestIsNicReservedForAnotherVMInCloudError(t *testing.T) {
	testCases := createSimpleCloudErrorCodeTests(
		NicReservedForAnotherVM,
		"NIC Reserved for Another VM",
	)
	checkCloudErrors(t, "IsNicReservedForAnotherVMInCloudError", testCases, IsNicReservedForAnotherVMInCloudError)
}

func TestIsSKUNotAvailableInCloudError(t *testing.T) {
	testCases := createSimpleCloudErrorCodeTests(
		SKUNotAvailableErrorCode,
		"SKU Not Available",
	)
	checkCloudErrors(t, "IsSKUNotAvailableInCloudError", testCases, IsSKUNotAvailableInCloudError)
}
