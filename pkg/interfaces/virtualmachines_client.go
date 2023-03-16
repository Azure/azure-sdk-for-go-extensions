package armcompute

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// IVirtualMachinesClient ...
type IVirtualMachinesClient interface {
	// BeginAssessPatches - Assess patches on the VM.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginAssessPatchesOptions contains the optional parameters for the VirtualMachinesClient.BeginAssessPatches
	// method.
	BeginAssessPatches(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginAssessPatchesOptions) (*runtime.Poller[armcompute.VirtualMachinesClientAssessPatchesResponse], error)
	// BeginCapture - Captures the VM by copying virtual hard disks of the VM and outputs a template that can be used to create
	// similar VMs.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// parameters - Parameters supplied to the Capture Virtual Machine operation.
	// options - VirtualMachinesClientBeginCaptureOptions contains the optional parameters for the VirtualMachinesClient.BeginCapture
	// method.
	BeginCapture(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.VirtualMachineCaptureParameters, options *armcompute.VirtualMachinesClientBeginCaptureOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCaptureResponse], error)
	// BeginConvertToManagedDisks - Converts virtual machine disks from blob-based to managed disks. Virtual machine must be stop-deallocated
	// before invoking this operation.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginConvertToManagedDisksOptions contains the optional parameters for the VirtualMachinesClient.BeginConvertToManagedDisks
	// method.
	BeginConvertToManagedDisks(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginConvertToManagedDisksOptions) (*runtime.Poller[armcompute.VirtualMachinesClientConvertToManagedDisksResponse], error)
	// BeginCreateOrUpdate - The operation to create or update a virtual machine. Please note some properties can be set only
	// during virtual machine creation.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// parameters - Parameters supplied to the Create Virtual Machine operation.
	// options - VirtualMachinesClientBeginCreateOrUpdateOptions contains the optional parameters for the VirtualMachinesClient.BeginCreateOrUpdate
	// method.
	BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.VirtualMachine, options *armcompute.VirtualMachinesClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientCreateOrUpdateResponse], error)
	// BeginDeallocate - Shuts down the virtual machine and releases the compute resources. You are not billed for the compute
	// resources that this virtual machine uses.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginDeallocateOptions contains the optional parameters for the VirtualMachinesClient.BeginDeallocate
	// method.
	BeginDeallocate(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginDeallocateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeallocateResponse], error)
	// BeginDelete - The operation to delete a virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginDeleteOptions contains the optional parameters for the VirtualMachinesClient.BeginDelete
	// method.
	BeginDelete(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginDeleteOptions) (*runtime.Poller[armcompute.VirtualMachinesClientDeleteResponse], error)
	// Generalize - Sets the OS state of the virtual machine to generalized. It is recommended to sysprep the virtual machine
	// before performing this operation. For Windows, please refer to Create a managed image of a
	// generalized VM in Azure [https://docs.microsoft.com/azure/virtual-machines/windows/capture-image-resource]. For Linux,
	// please refer to How to create an image of a virtual machine or VHD
	// [https://docs.microsoft.com/azure/virtual-machines/linux/capture-image].
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientGeneralizeOptions contains the optional parameters for the VirtualMachinesClient.Generalize
	// method.
	Generalize(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientGeneralizeOptions) (armcompute.VirtualMachinesClientGeneralizeResponse, error)
	// Get - Retrieves information about the model view or the instance view of a virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientGetOptions contains the optional parameters for the VirtualMachinesClient.Get method.
	Get(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientGetOptions) (armcompute.VirtualMachinesClientGetResponse, error)
	// BeginInstallPatches - Installs patches on the VM.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// installPatchesInput - Input for InstallPatches as directly received by the API
	// options - VirtualMachinesClientBeginInstallPatchesOptions contains the optional parameters for the VirtualMachinesClient.BeginInstallPatches
	// method.
	BeginInstallPatches(ctx context.Context, resourceGroupName string, vmName string, installPatchesInput armcompute.VirtualMachineInstallPatchesParameters, options *armcompute.VirtualMachinesClientBeginInstallPatchesOptions) (*runtime.Poller[armcompute.VirtualMachinesClientInstallPatchesResponse], error)
	// InstanceView - Retrieves information about the run-time state of a virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientInstanceViewOptions contains the optional parameters for the VirtualMachinesClient.InstanceView
	// method.
	InstanceView(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientInstanceViewOptions) (armcompute.VirtualMachinesClientInstanceViewResponse, error)
	// NewListPager - Lists all of the virtual machines in the specified resource group. Use the nextLink property in the response
	// to get the next page of virtual machines.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// options - VirtualMachinesClientListOptions contains the optional parameters for the VirtualMachinesClient.List method.
	NewListPager(resourceGroupName string, options *armcompute.VirtualMachinesClientListOptions) *runtime.Pager[armcompute.VirtualMachinesClientListResponse]
	// NewListAllPager - Lists all of the virtual machines in the specified subscription. Use the nextLink property in the response
	// to get the next page of virtual machines.
	// Generated from API version 2022-11-01
	// options - VirtualMachinesClientListAllOptions contains the optional parameters for the VirtualMachinesClient.ListAll method.
	NewListAllPager(options *armcompute.VirtualMachinesClientListAllOptions) *runtime.Pager[VirtualMachinesClientListAllResponse]
	// NewListAvailableSizesPager - Lists all available virtual machine sizes to which the specified virtual machine can be resized.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientListAvailableSizesOptions contains the optional parameters for the VirtualMachinesClient.ListAvailableSizes
	// method.
	NewListAvailableSizesPager(resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientListAvailableSizesOptions) *runtime.Pager[armcompute.VirtualMachinesClientListAvailableSizesResponse]
	// NewListByLocationPager - Gets all the virtual machines under the specified subscription for the specified location.
	// Generated from API version 2022-11-01
	// location - The location for which virtual machines under the subscription are queried.
	// options - VirtualMachinesClientListByLocationOptions contains the optional parameters for the VirtualMachinesClient.ListByLocation
	// method.
	NewListByLocationPager(location string, options *armcompute.VirtualMachinesClientListByLocationOptions) *runtime.Pager[armcompute.VirtualMachinesClientListByLocationResponse]
	// BeginPerformMaintenance - The operation to perform maintenance on a virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginPerformMaintenanceOptions contains the optional parameters for the VirtualMachinesClient.BeginPerformMaintenance
	// method.
	BeginPerformMaintenance(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginPerformMaintenanceOptions) (*runtime.Poller[armcompute.VirtualMachinesClientPerformMaintenanceResponse], error)
	// BeginPowerOff - The operation to power off (stop) a virtual machine. The virtual machine can be restarted with the same
	// provisioned resources. You are still charged for this virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginPowerOffOptions contains the optional parameters for the VirtualMachinesClient.BeginPowerOff
	// method.
	BeginPowerOff(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginPowerOffOptions) (*runtime.Poller[armcompute.VirtualMachinesClientPowerOffResponse], error)
	// BeginReapply - The operation to reapply a virtual machine's state.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginReapplyOptions contains the optional parameters for the VirtualMachinesClient.BeginReapply
	// method.
	BeginReapply(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginReapplyOptions) (*runtime.Poller[armcompute.VirtualMachinesClientReapplyResponse], error)
	// BeginRedeploy - Shuts down the virtual machine, moves it to a new node, and powers it back on.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginRedeployOptions contains the optional parameters for the VirtualMachinesClient.BeginRedeploy
	// method.
	BeginRedeploy(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginRedeployOptions) (*runtime.Poller[armcompute.VirtualMachinesClientRedeployResponse], error)
	// BeginReimage - Reimages (upgrade the operating system) a virtual machine which don't have a ephemeral OS disk, for virtual
	// machines who have a ephemeral OS disk the virtual machine is reset to initial state. NOTE:
	// The retaining of old OS disk depends on the value of deleteOption of OS disk. If deleteOption is detach, the old OS disk
	// will be preserved after reimage. If deleteOption is delete, the old OS disk
	// will be deleted after reimage. The deleteOption of the OS disk should be updated accordingly before performing the reimage.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginReimageOptions contains the optional parameters for the VirtualMachinesClient.BeginReimage
	// method.
	BeginReimage(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginReimageOptions) (*runtime.Poller[armcompute.VirtualMachinesClientReimageResponse], error)
	// BeginRestart - The operation to restart a virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginRestartOptions contains the optional parameters for the VirtualMachinesClient.BeginRestart
	// method.
	BeginRestart(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginRestartOptions) (*runtime.Poller[armcompute.VirtualMachinesClientRestartResponse], error)
	// RetrieveBootDiagnosticsData - The operation to retrieve SAS URIs for a virtual machine's boot diagnostic logs.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientRetrieveBootDiagnosticsDataOptions contains the optional parameters for the VirtualMachinesClient.RetrieveBootDiagnosticsData
	// method.
	RetrieveBootDiagnosticsData(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientRetrieveBootDiagnosticsDataOptions) (armcompute.VirtualMachinesClientRetrieveBootDiagnosticsDataResponse, error)
	// BeginRunCommand - Run command on the VM.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// parameters - Parameters supplied to the Run command operation.
	// options - VirtualMachinesClientBeginRunCommandOptions contains the optional parameters for the VirtualMachinesClient.BeginRunCommand
	// method.
	BeginRunCommand(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.RunCommandInput, options *armcompute.VirtualMachinesClientBeginRunCommandOptions) (*runtime.Poller[armcompute.VirtualMachinesClientRunCommandResponse], error)
	// SimulateEviction - The operation to simulate the eviction of spot virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientSimulateEvictionOptions contains the optional parameters for the VirtualMachinesClient.SimulateEviction
	// method.
	SimulateEviction(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientSimulateEvictionOptions) (armcompute.VirtualMachinesClientSimulateEvictionResponse, error)
	// BeginStart - The operation to start a virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// options - VirtualMachinesClientBeginStartOptions contains the optional parameters for the VirtualMachinesClient.BeginStart
	// method.
	BeginStart(ctx context.Context, resourceGroupName string, vmName string, options *armcompute.VirtualMachinesClientBeginStartOptions) (*runtime.Poller[armcompute.VirtualMachinesClientStartResponse], error)
	// BeginUpdate - The operation to update a virtual machine.
	// If the operation fails it returns an *azcore.ResponseError type.
	// Generated from API version 2022-11-01
	// resourceGroupName - The name of the resource group.
	// vmName - The name of the virtual machine.
	// parameters - Parameters supplied to the Update Virtual Machine operation.
	// options - VirtualMachinesClientBeginUpdateOptions contains the optional parameters for the VirtualMachinesClient.BeginUpdate
	// method.
	BeginUpdate(ctx context.Context, resourceGroupName string, vmName string, parameters armcompute.VirtualMachineUpdate, options *armcompute.VirtualMachinesClientBeginUpdateOptions) (*runtime.Poller[armcompute.VirtualMachinesClientUpdateResponse], error)
}
