package middleware

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/stretchr/testify/assert"
)

func TestArmRequestMetrics(t *testing.T) {
	testInfo, err := testInfoFromEnv()
	if err != nil {
		t.Skipf("test requires setup: %s", err)
	}
	token, err := azidentity.NewClientSecretCredential(testInfo.TenantID, testInfo.SPNClientID, testInfo.SPNClientSecret, nil)
	assert.NoError(t, err)

	clientOptions := DefaultArmOpts("test", &myCollector{logger: t})
	clientOptions.DisableRPRegistration = true
	client, err := armcontainerservice.NewManagedClustersClient("notexistingSub", token, clientOptions)
	assert.NoError(t, err)

	_, err = client.BeginCreateOrUpdate(context.Background(), "test", "test", armcontainerservice.ManagedCluster{Location: to.Ptr("eastus")}, nil)
	// here the error is parsed from response body twice
	// 1. by ArmRequestMetricPolicy, parse and log, and throw away.
	// 2. by generated client function: runtime.HasStatusCode, and return to here.
	assert.Error(t, err)

	respErr := &azcore.ResponseError{}
	assert.True(t, errors.As(err, &respErr))
	assert.Equal(t, respErr.ErrorCode, "InvalidSubscriptionId")
}

var _ ArmRequestMetricCollector = &myCollector{}

type logger interface {
	Logf(format string, args ...interface{})
}

type myCollector struct {
	logger logger
}

func (c *myCollector) RequestStarted(iReq *RequestInfo) {
	c.logger.Logf("RequestStarted, on %s, URL=%s\n", c.formatResourceId(iReq.ArmResId), iReq.Request.URL)
}

func (c *myCollector) RequestCompleted(iReq *RequestInfo, iResp *ResponseInfo) {
	c.logger.Logf("RequestFinished with %d, on %s, URL=%s\n", iResp.Response.StatusCode, c.formatResourceId(iReq.ArmResId), iReq.Request.URL)
}

func (c *myCollector) formatResourceId(resId *arm.ResourceID) string {
	return fmt.Sprintf("ResourceType=%s, subscription=%s, resourceGroup=%s", resId.ResourceType.String(), resId.SubscriptionID, resId.ResourceGroupName)
}

type aadInfo struct {
	TenantID        string
	SPNClientID     string
	SPNClientSecret string
	SubscriptionID  string
}

func testInfoFromEnv() (*aadInfo, error) {
	tenantID, ok := os.LookupEnv("AAD_Tenant")
	if !ok {
		return nil, errors.New("AAD_Tenant is not set")
	}

	clientID, ok := os.LookupEnv("AAD_ClientID")
	if !ok {
		return nil, errors.New("AAD_ClientID is not set")
	}

	clientSecret, ok := os.LookupEnv("AAD_ClientSecret")
	if !ok {
		return nil, errors.New("AAD_ClientSecret is not set")
	}

	subscriptionID, ok := os.LookupEnv("Azure_Subscription")
	if !ok {
		return nil, errors.New("Azure_Subscription is not set")
	}

	return &aadInfo{
		TenantID:        tenantID,
		SPNClientID:     clientID,
		SPNClientSecret: clientSecret,
		SubscriptionID:  subscriptionID,
	}, nil
}
