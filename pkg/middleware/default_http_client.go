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

package middleware

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/Azure/go-armbalancer"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
)

var (
	defaultHTTPClient *http.Client
	defaultTransport  http.RoundTripper
)

// DefaultHTTPClient returns a shared http client, and transport leveraging armbalancer for
// clientside loadbalancing, so we can leverage HTTP/2, and not get throttled by arm at the instance level.
func DefaultHTTPClient() *http.Client {
	return defaultHTTPClient
}

func init() {
	defaultTransport = armbalancer.New(armbalancer.Options{
		// PoolSize is the number of clientside http/2 persistent connections
		// we want to have configured in our transport. Note, that without clientside loadbalancing
		// with arm, HTTP/2 Will force persistent connection to stick to a single arm instance, and will
		// result in a substantial amount of throttling
		PoolSize: 100,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	})
	defaultHTTPClient = &http.Client{
		Transport: otelhttp.NewTransport(defaultTransport, otelhttp.WithPropagators(propagation.TraceContext{})),
	}
}
