package errors

const (

	// Error codes
	ResourceNotFound                      = "ResourceNotFound"
	OperationNotAllowed                   = "OperationNotAllowed"
	AllocationFailed                      = "AllocationFailed"
	OverconstrainedAllocationRequest      = "OverconstrainedAllocationRequest"
	ZoneAllocationFailed                  = "ZonalAllocationFailed"
	OverconstrainedZonalAllocationRequest = "OverconstrainedZonalAllocationRequest"
	NicReservedForAnotherVM               = "NicReservedForAnotherVm"
	SKUNotAvailableErrorCode              = "SkuNotAvailable"

	// Error search terms
	LowPriorityQuotaExceededTerm  = "LowPriorityCores"
	SKUFamilyQuotaExceededTerm    = "Family Cores quota"
	SubscriptionQuotaExceededTerm = "Submit a request for Quota increase"
	RegionalQuotaExceededTerm     = "exceeding approved Total Regional Cores quota"
)
