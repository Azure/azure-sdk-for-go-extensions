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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// IsNotFoundErr is used to determine if we are failing to find a resource within azure
func IsNotFoundErr(err *azcore.ResponseError) bool {
	if err == nil {
		return false
	}
	if err.ErrorCode == ResourceNotFound {
		return true
	}
	return false
}

// SubscriptionQuotaHasBeenReached tells us if we have exceeded our Quota
func SubscriptionQuotaHasBeenReached(err *azcore.ResponseError) bool {
	if err == nil || err.ErrorCode != OperationNotAllowed {
		return false
	}

	if strings.Contains(err.Error(), SubscriptionQuotaExceededTerm) {
		return true
	}
	return false
}

// RegionalQuotaHasBeenReached communicates if we have reached the quota for a given region
func RegionalQuotaHasBeenReached(err *azcore.ResponseError) bool {
	if err == nil || err.ErrorCode != OperationNotAllowed {
		return false
	}

	if strings.Contains(err.Error(), RegionalQuotaExceededTerm) {
		return true
	}
	return false
}
