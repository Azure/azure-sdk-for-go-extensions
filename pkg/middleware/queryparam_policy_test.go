package middleware

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/stretchr/testify/assert"
)

func TestQueryPolicyAppend(t *testing.T) {
	t.Parallel()

	req, err := runtime.NewRequest(context.TODO(), "PUT", "http://:13333/?foo=boo")
	assert.NoError(t, err)

	queryPolicy := QueryParameterPolicy{
		Name:    "foo",
		Value:   "bar",
		Replace: false,
	}

	// here we expect an error
	_, err = queryPolicy.Do(req)
	assert.Error(t, err, "no more policies")

	assert.Equal(t, req.Raw().URL.RawQuery, "foo=boo&foo=bar")
}

func TestQueryPolicyReplace(t *testing.T) {
	t.Parallel()

	req, err := runtime.NewRequest(context.TODO(), "PUT", "http://:13333/?foo=boo")
	assert.NoError(t, err)

	queryPolicy := QueryParameterPolicy{
		Name:    "foo",
		Value:   "bar",
		Replace: true,
	}

	// here we expect an error
	_, err = queryPolicy.Do(req)
	assert.Error(t, err, "no more policies")

	assert.Equal(t, req.Raw().URL.RawQuery, "foo=bar")
}

func TestQueryPolicyReplaceEmpty(t *testing.T) {
	t.Parallel()

	req, err := runtime.NewRequest(context.TODO(), "PUT", "http://:13333/")
	assert.NoError(t, err)

	queryPolicy := QueryParameterPolicy{
		Name:    "foo",
		Value:   "bar",
		Replace: true,
	}

	// here we expect an error
	_, err = queryPolicy.Do(req)
	assert.Error(t, err, "no more policies")

	assert.Equal(t, req.Raw().URL.RawQuery, "foo=bar")
}

func TestQueryPolicyAppendEmpty(t *testing.T) {
	t.Parallel()

	req, err := runtime.NewRequest(context.TODO(), "PUT", "http://:13333/")
	assert.NoError(t, err)

	queryPolicy := QueryParameterPolicy{
		Name:    "foo",
		Value:   "bar",
		Replace: false,
	}

	// here we expect an error
	_, err = queryPolicy.Do(req)
	assert.Error(t, err, "no more policies")

	assert.Equal(t, req.Raw().URL.RawQuery, "foo=bar")
}
