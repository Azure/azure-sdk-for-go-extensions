/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package errors

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v7"
)

// Note: these functions should generally be kept in sync with armerrors.go counterparts
// An alternative is to expose a generic interface that works for both.
// Although, there seems to be no direct common interface, if not code/message extraction.

// extractCloudErrorDetails extracts code and message from CloudErrorBody safely
func extractCloudErrorDetails(cloudError armcontainerservice.CloudErrorBody) (code, message string) {
	if cloudError.Code != nil {
		code = *cloudError.Code
	}
	if cloudError.Message != nil {
		message = *cloudError.Message
	}
	return code, message
}

// ZonalAllocationFailureOccurredInCloudError communicates if we have failed to allocate a resource in a zone, and should try another zone.
// To learn more about zonal allocation failures, visit: http://aka.ms/allocation-guidance
func ZonalAllocationFailureOccurredInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, _ := extractCloudErrorDetails(cloudError)
	return isZonalAllocationFailed(code)
}

// AllocationFailureOccurredInCloudError communicates if we have failed to allocate a resource in a region, and should try another region.
func AllocationFailureOccurredInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, _ := extractCloudErrorDetails(cloudError)
	return isAllocationFailed(code)
}

// OverconstrainedAllocationFailureOccurredInCloudError communicates if we have failed to allocate a resource that meets constraints specified in the request, and should try another region.
func OverconstrainedAllocationFailureOccurredInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, _ := extractCloudErrorDetails(cloudError)
	return isOverconstrainedAllocationFailed(code)
}

// OverconstrainedZonalAllocationFailureOccurredInCloudError communicates if we have failed to allocate a resource that meets constraints specified in the request, and should try another zone.
func OverconstrainedZonalAllocationFailureOccurredInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, _ := extractCloudErrorDetails(cloudError)
	return isOverconstrainedZonalAllocationFailed(code)
}

// SKUFamilyQuotaHasBeenReachedInCloudError tells us if we have exceeded our Quota.
func SKUFamilyQuotaHasBeenReachedInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, message := extractCloudErrorDetails(cloudError)
	return isSKUFamilyQuotaExceeded(code, message)
}

// SubscriptionQuotaHasBeenReachedInCloudError tells us if we have exceeded our Quota.
func SubscriptionQuotaHasBeenReachedInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, message := extractCloudErrorDetails(cloudError)
	return isSubscriptionQuotaExceeded(code, message)
}

// RegionalQuotaHasBeenReachedInCloudError communicates if we have reached the quota limit for a given region under a specific subscription
func RegionalQuotaHasBeenReachedInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, message := extractCloudErrorDetails(cloudError)
	return isRegionalQuotaExceeded(code, message)
}

// LowPriorityQuotaHasBeenReachedInCloudError communicates if we have reached the quota limit for low priority VMs under a specific subscription
// Low priority VMs are generally Spot VMs, but can also be low priority VMs created via the Azure CLI or Azure Portal
func LowPriorityQuotaHasBeenReachedInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, message := extractCloudErrorDetails(cloudError)
	return isLowPriorityQuotaExceeded(code, message)
}

// IsNicReservedForAnotherVMInCloudError occurs when a NIC is associated with another VM during deletion. See https://aka.ms/deletenic
func IsNicReservedForAnotherVMInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, _ := extractCloudErrorDetails(cloudError)
	return isNicReservedForVM(code)
}

// IsSKUNotAvailableInCloudError https://aka.ms/azureskunotavailable: either not available for a location or zone, or out of capacity for Spot.
func IsSKUNotAvailableInCloudError(cloudError armcontainerservice.CloudErrorBody) bool {
	code, _ := extractCloudErrorDetails(cloudError)
	return isSKUNotAvailable(code)
}
