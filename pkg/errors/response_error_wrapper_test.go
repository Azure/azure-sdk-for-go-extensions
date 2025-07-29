package errors

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/stretchr/testify/assert"
)

func TestResponseErrorWrapper_Error(t *testing.T) {
	tests := []struct {
		name           string
		setupError     func() *azcore.ResponseError
		expectedOutput string
	}{
		{
			name: "SubnetIsFull error with all fields",
			setupError: func() *azcore.ResponseError {
				body := `{
                    "error": {
                        "code": "SubnetIsFull",
                        "message": "Subnet /subscriptions/123/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/test-subnet with address prefix 10.0.0.0/24 does not have enough capacity for 1 IP addresses.",
                        "details": []
                    }
                }`

				resp := &http.Response{
					StatusCode: 400,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Request: &http.Request{
						Method: "PUT",
						URL: &url.URL{
							Scheme: "https",
							Host:   "management.azure.com",
							Path:   "/subscriptions/123/resourceGroups/test-rg/providers/Microsoft.Network/networkInterfaces/test-nic",
						},
					},
				}

				return &azcore.ResponseError{
					ErrorCode:   "SubnetIsFull",
					StatusCode:  400,
					RawResponse: resp,
				}
			},
			expectedOutput: "HTTP CODE: 400, ERROR CODE: SubnetIsFull, MESSAGE: Subnet /subscriptions/123/resourceGroups/test-rg/providers/Microsoft.Network/virtualNetworks/test-vnet/subnets/test-subnet with address prefix 10.0.0.0/24 does not have enough capacity for 1 IP addresses., REQUEST: PUT https://management.azure.com/subscriptions/123/resourceGroups/test-rg/providers/Microsoft.Network/networkInterfaces/test-nic",
		},
		{
			name: "Error with missing error code",
			setupError: func() *azcore.ResponseError {
				body := `{
		            "error": {
		                "message": "Something went wrong"
		            }
		        }`

				resp := &http.Response{
					StatusCode: 500,
					Status:     "500 Internal Server Error",
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Request: &http.Request{
						Method: "GET",
						URL: &url.URL{
							Scheme: "https",
							Host:   "management.azure.com",
							Path:   "/subscriptions/123/resourceGroups/test-rg",
						},
					},
				}

				return &azcore.ResponseError{
					ErrorCode:   "", // No error code
					StatusCode:  500,
					RawResponse: resp,
				}
			},
			expectedOutput: "HTTP CODE: 500, ERROR CODE: UNAVAILABLE, MESSAGE: Something went wrong, REQUEST: GET https://management.azure.com/subscriptions/123/resourceGroups/test-rg",
		},
		{
			name: "Error with nil RawResponse",
			setupError: func() *azcore.ResponseError {
				return &azcore.ResponseError{
					ErrorCode:   "TestError",
					StatusCode:  403,
					RawResponse: nil,
				}
			},
			expectedOutput: "HTTP CODE: 403, ERROR CODE: TestError, MESSAGE: UNAVAILABLE, REQUEST: UNKNOWN UNAVAILABLE",
		},
		{
			name: "format without error wrapper in json body",
			setupError: func() *azcore.ResponseError {
				body := `{
                    "code": "InvalidConfiguration",
                    "message": "The specified configuration is invalid and cannot be processed.",
                    "details": []
                }`

				resp := &http.Response{
					StatusCode: 400,
					Status:     "400 Bad Request",
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Request: &http.Request{
						Method: "PUT",
						URL: &url.URL{
							Scheme: "https",
							Host:   "management.azure.com",
							Path:   "/subscriptions/123/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster",
						},
					},
				}

				return &azcore.ResponseError{
					ErrorCode:   "InvalidConfiguration",
					StatusCode:  400,
					RawResponse: resp,
				}
			},
			expectedOutput: "HTTP CODE: 400, ERROR CODE: InvalidConfiguration, MESSAGE: The specified configuration is invalid and cannot be processed., REQUEST: PUT https://management.azure.com/subscriptions/123/resourceGroups/test-rg/providers/Microsoft.ContainerService/managedClusters/test-cluster",
		},
		{
			name: "Nil ResponseError",
			setupError: func() *azcore.ResponseError {
				return nil
			},
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			azErr := tt.setupError()
			responseErrWrapper := NewResponseErrorWrapper(azErr)

			actual := responseErrWrapper.Error()
			assert.Equal(t, tt.expectedOutput, actual)
		})
	}
}

func TestResponseErrorWrapper_MessageCaching(t *testing.T) {
	t.Run("Message caching - Error() called multiple times should return same cached message", func(t *testing.T) {
		body := `{
			"error": {
				"code": "TestCode",
				"message": "Original test message"
			}
		}`

		resp := &http.Response{
			StatusCode: 400,
			Status:     "400 Bad Request",
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Request: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Scheme: "https",
					Host:   "management.azure.com",
					Path:   "/test",
				},
			},
		}

		respErr := &azcore.ResponseError{
			ErrorCode:   "TestCode",
			StatusCode:  400,
			RawResponse: resp,
		}

		wrapper := NewResponseErrorWrapper(respErr)

		// Call Error() multiple times
		firstCall := wrapper.Error()
		secondCall := wrapper.Error()
		thirdCall := wrapper.Error()

		// All calls should return the exact same string
		assert.Equal(t, firstCall, secondCall)
		assert.Equal(t, secondCall, thirdCall)
		assert.Equal(t, firstCall, thirdCall)

		expectedMessage := "HTTP CODE: 400, ERROR CODE: TestCode, MESSAGE: Original test message, REQUEST: GET https://management.azure.com/test"
		assert.Equal(t, expectedMessage, firstCall)
	})

	t.Run("Message caching with nil ResponseError", func(t *testing.T) {
		wrapper := NewResponseErrorWrapper(nil)

		// Call Error() multiple times with nil ResponseError
		firstCall := wrapper.Error()
		secondCall := wrapper.Error()

		// Both calls should return empty string
		assert.Equal(t, "", firstCall)
		assert.Equal(t, "", secondCall)
		assert.Equal(t, firstCall, secondCall)
	})
}

func TestResponseErrorWrapper_ErrorMessageExtractionEdgeCases(t *testing.T) {
	createResponseErrorWithBody := func(body string) *azcore.ResponseError {
		resp := &http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Request: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Scheme: "https",
					Host:   "example.com",
					Path:   "/test",
				},
			},
		}
		return &azcore.ResponseError{
			ErrorCode:   "TestCode",
			StatusCode:  400,
			RawResponse: resp,
		}
	}

	tests := []struct {
		name            string
		jsonBody        string
		expectedMessage string
	}{
		{
			name: "JSON with escaped quotes in message",
			jsonBody: `{
				"error": {
					"code": "TestCode",
					"message": "Error with \"quoted text\" inside"
				}
			}`,
			expectedMessage: `Error with "quoted text" inside`,
		},
		{
			name: "JSON with special characters in message",
			jsonBody: `{
				"error": {
					"code": "TestCode", 
					"message": "Error with unicode: ñáéíóú and newlines\nand tabs\t"
				}
			}`,
			expectedMessage: "Error with unicode: ñáéíóú and newlines and tabs ",
		},
		{
			name: "Malformed JSON in error response",
			jsonBody: `{
				"error": {
					"code": "TestCode",
					"message": "Valid message"
				}
				"invalid": "json"
			}`,
			expectedMessage: "Valid message", // Should still extract the message despite malformed JSON
		},
		{
			name: "Multiple message fields in JSON",
			jsonBody: `{
				"error": {
					"code": "TestCode",
					"message": "First message",
					"details": {
						"message": "Second message"
					}
				}
			}`,
			expectedMessage: "First message", // Should pick the first one
		},
		{
			name: "Empty message field",
			jsonBody: `{
				"error": {
					"code": "TestCode",
					"message": ""
				}
			}`,
			expectedMessage: "", // Should return empty string, not "UNAVAILABLE"
		},
		{
			name: "Message field with only whitespace",
			jsonBody: `{
				"error": {
					"code": "TestCode",
					"message": "   "
				}
			}`,
			expectedMessage: "   ", // Should preserve whitespace
		},
		{
			name: "No message field in JSON",
			jsonBody: `{
				"error": {
					"code": "TestCode",
					"details": "Some other field"
				}
			}`,
			expectedMessage: "UNAVAILABLE", // Should fallback to UNAVAILABLE
		},
		{
			name:            "Completely invalid JSON",
			jsonBody:        `invalid json content`,
			expectedMessage: "UNAVAILABLE", // Should fallback to UNAVAILABLE
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respErr := createResponseErrorWithBody(tt.jsonBody)
			wrapper := NewResponseErrorWrapper(respErr)

			result := wrapper.Error()

			// Check that the message part contains the expected message
			expectedFull := "HTTP CODE: 400, ERROR CODE: TestCode, MESSAGE: " + tt.expectedMessage + ", REQUEST: GET https://example.com/test"
			assert.Equal(t, expectedFull, result)
		})
	}
}

func TestResponseErrorWrapper_IntegrationRealisticScenarios(t *testing.T) {
	// Helper function to create ResponseError for integration testing
	createResponseErrorWithBody := func(statusCode int, errorCode, body string, method, host, path string) *azcore.ResponseError {
		resp := &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Request: &http.Request{
				Method: method,
				URL: &url.URL{
					Scheme: "https",
					Host:   host,
					Path:   path,
				},
			},
		}
		return &azcore.ResponseError{
			ErrorCode:   errorCode,
			StatusCode:  statusCode,
			RawResponse: resp,
		}
	}

	t.Run("Real Azure ARM error format", func(t *testing.T) {
		// This is based on actual Azure ARM error responses
		body := `{
			"error": {
				"code": "ResourceNotFound",
				"message": "The Resource 'Microsoft.Compute/virtualMachines/myVM' under resource group 'myResourceGroup' was not found. For more details please go to https://aka.ms/ARMResourceNotFoundFix",
				"details": [
					{
						"code": "NotFound",
						"target": "Microsoft.Compute/virtualMachines/myVM",
						"message": "The requested resource does not exist."
					}
				]
			}
		}`

		respErr := createResponseErrorWithBody(
			404,
			"ResourceNotFound",
			body,
			"GET",
			"management.azure.com",
			"/subscriptions/12345/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/myVM",
		)

		wrapper := NewResponseErrorWrapper(respErr)
		result := wrapper.Error()

		expectedMessage := "HTTP CODE: 404, ERROR CODE: ResourceNotFound, MESSAGE: The Resource 'Microsoft.Compute/virtualMachines/myVM' under resource group 'myResourceGroup' was not found. For more details please go to https://aka.ms/ARMResourceNotFoundFix, REQUEST: GET https://management.azure.com/subscriptions/12345/resourceGroups/myResourceGroup/providers/Microsoft.Compute/virtualMachines/myVM"
		assert.Equal(t, expectedMessage, result)
	})

	t.Run("Azure Storage error format", func(t *testing.T) {
		// This is based on actual Azure Storage error responses
		body := `{
			"error": {
				"code": "BlobNotFound",
				"message": "The specified blob does not exist.\nRequestId:12345678-1234-1234-1234-123456789abc\nTime:2023-07-03T12:00:00.0000000Z"
			}
		}`

		respErr := createResponseErrorWithBody(
			404,
			"BlobNotFound",
			body,
			"GET",
			"mystorageaccount.blob.core.windows.net",
			"/mycontainer/myblob.txt",
		)

		wrapper := NewResponseErrorWrapper(respErr)
		result := wrapper.Error()

		// Note: \n should be converted to spaces
		expectedMessage := "HTTP CODE: 404, ERROR CODE: BlobNotFound, MESSAGE: The specified blob does not exist. RequestId:12345678-1234-1234-1234-123456789abc Time:2023-07-03T12:00:00.0000000Z, REQUEST: GET https://mystorageaccount.blob.core.windows.net/mycontainer/myblob.txt"
		assert.Equal(t, expectedMessage, result)
	})

	t.Run("Error response with nested error details", func(t *testing.T) {
		// Complex nested structure with multiple levels
		body := `{
			"error": {
				"code": "ValidationError",
				"message": "Request validation failed with multiple errors.",
				"details": [
					{
						"code": "InvalidParameter", 
						"target": "location",
						"message": "The provided location 'invalid-region' is not available.",
						"details": [
							{
								"code": "LocationNotAvailable",
								"message": "Region 'invalid-region' does not support this resource type."
							}
						]
					},
					{
						"code": "MissingParameter",
						"target": "sku",
						"message": "Required parameter 'sku' is missing from the request."
					}
				],
				"additionalInfo": [
					{
						"type": "PolicyViolation",
						"info": {
							"policyDefinitionDisplayName": "Resource Location Policy"
						}
					}
				]
			}
		}`

		respErr := createResponseErrorWithBody(
			400,
			"ValidationError",
			body,
			"PUT",
			"management.azure.com",
			"/subscriptions/12345/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm",
		)

		wrapper := NewResponseErrorWrapper(respErr)
		result := wrapper.Error()

		expectedMessage := "HTTP CODE: 400, ERROR CODE: ValidationError, MESSAGE: Request validation failed with multiple errors., REQUEST: PUT https://management.azure.com/subscriptions/12345/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm"
		assert.Equal(t, expectedMessage, result)
	})

	t.Run("Azure Key Vault error format", func(t *testing.T) {
		// Key Vault has a slightly different error format
		body := `{
			"error": {
				"code": "Forbidden",
				"message": "The user, group or application 'appid=12345678-1234-1234-1234-123456789abc;oid=87654321-4321-4321-4321-210987654321;iss=https://sts.windows.net/tenant-id/' does not have secrets get permission on key vault 'myvault;location=eastus'. For help resolving this issue, please see https://go.microsoft.com/fwlink/?linkid=2125287",
				"innererror": {
					"code": "AccessDenied"
				}
			}
		}`

		respErr := createResponseErrorWithBody(
			403,
			"Forbidden",
			body,
			"GET",
			"myvault.vault.azure.net",
			"/secrets/mysecret",
		)

		wrapper := NewResponseErrorWrapper(respErr)
		result := wrapper.Error()

		expectedMessage := "HTTP CODE: 403, ERROR CODE: Forbidden, MESSAGE: The user, group or application 'appid=12345678-1234-1234-1234-123456789abc;oid=87654321-4321-4321-4321-210987654321;iss=https://sts.windows.net/tenant-id/' does not have secrets get permission on key vault 'myvault;location=eastus'. For help resolving this issue, please see https://go.microsoft.com/fwlink/?linkid=2125287, REQUEST: GET https://myvault.vault.azure.net/secrets/mysecret"
		assert.Equal(t, expectedMessage, result)
	})
}

func TestAsWrappedResponseError(t *testing.T) {
	t.Run("Returns original error for non-ResponseError", func(t *testing.T) {
		err := errors.New("some other error")
		result := AsWrappedResponseError(err)
		assert.Equal(t, err, result)
	})

	t.Run("Returns nil interface for nil error", func(t *testing.T) {
		result := AsWrappedResponseError(nil)
		assert.Nil(t, result)
	})

	t.Run("Returns wrapped error for valid ResponseError", func(t *testing.T) {
		respErr := &azcore.ResponseError{ErrorCode: "Test", StatusCode: 400}

		result := AsWrappedResponseError(respErr)
		assert.NotNil(t, result)

		// Should be able to type assert back to concrete type
		wrapper, ok := result.(*ResponseErrorWrapper)
		assert.True(t, ok)
		assert.NotNil(t, wrapper)
	})

	t.Run("Implements error interface correctly", func(t *testing.T) {
		respErr := &azcore.ResponseError{ErrorCode: "Test", StatusCode: 400}

		result := AsWrappedResponseError(respErr)

		// Should implement error interface
		var err error = result
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	})
}
