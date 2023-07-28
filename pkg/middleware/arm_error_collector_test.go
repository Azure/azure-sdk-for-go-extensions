package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/stretchr/testify/assert"
)

func TestArmRequestMetrics(t *testing.T) {
	myPolicy := &ArmRequestMetricPolicy{
		Collector: &myCollector{logger: t},
	}
	clientOptions := &arm.ClientOptions{
		ClientOptions: policy.ClientOptions{
			PerCallPolicies: []policy.Policy{myPolicy},
			Transport:       &mockServerTransport{},
		},
		DisableRPRegistration: true,
	}
	client, err := armcontainerservice.NewManagedClustersClient("notexistingSub", &mockTokenCredential{}, clientOptions)
	assert.NoError(t, err)

	_, err = client.BeginCreateOrUpdate(context.Background(), "test", "test", armcontainerservice.ManagedCluster{Location: to.Ptr("eastus")}, nil)
	// here the error is parsed from response body twice
	// 1. by ArmRequestMetricPolicy, parse and log, and throw away.
	// 2. by generated client function: runtime.HasStatusCode, and return to here.
	assert.Error(t, err)

	respErr := &azcore.ResponseError{}
	assert.True(t, errors.As(err, &respErr))
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

func (c *myCollector) RequestFailed(iReq *RequestInfo, iResp *ResponseInfo) {
	c.logger.Logf("RequestFailed with %d %s, on %s, URL=%s\n", iResp.Response.StatusCode, iResp.Error.Code, c.formatResourceId(iReq.ArmResId), iReq.Request.URL)
}

func (c *myCollector) formatResourceId(resId *arm.ResourceID) string {
	return fmt.Sprintf("ResourceType=%s, subscription=%s, resourceGroup=%s", resId.ResourceType.String(), resId.SubscriptionID, resId.ResourceGroupName)
}

type mockServerTransport struct{}

func (m *mockServerTransport) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 400,
		Body:       http.NoBody,
	}, nil
}

type mockTokenCredential struct{}

// GetToken implements azcore.TokenCredential
func (m *mockTokenCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token:     "fakeToken",
		ExpiresOn: time.Now().Add(1 * time.Hour),
	}, nil
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
