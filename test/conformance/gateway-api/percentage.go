/*
Copyright 2021 The Knative Authors

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

package ingress

import (
	"context"
	"errors"
	"math"
	"testing"

	"golang.org/x/sync/errgroup"
	"knative.dev/net-gateway-api/test"
	"knative.dev/networking/pkg/apis/networking"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// TestPercentage verifies that an Ingress splitting over multiple backends respects
// the given percentage distribution.
func TestPercentage(t *testing.T) {
	t.Parallel()
	ctx, clients := context.Background(), test.Setup(t)

	// Use a post-split injected header to establish which split we are sending traffic to.
	const headerName = "Foo-Bar-Baz"

	backends := make([]gatewayv1alpha2.HTTPBackendRef, 0, 10)
	weights := make(map[string]float64, len(backends))

	// Double the percentage of the split each iteration until it would overflow, and then
	// give the last route the remainder.
	percent, total := int32(1), int32(0)
	for i := 0; i < 10; i++ {
		weight := percent
		name, port, _ := CreateRuntimeService(ctx, t, clients, networking.ServicePortNameHTTP1)
		backends = append(backends,
			gatewayv1alpha2.HTTPBackendRef{
				BackendRef: gatewayv1alpha2.BackendRef{
					BackendObjectReference: gatewayv1alpha2.BackendObjectReference{
						Port: portNumPtr(port),
						Name: gatewayv1alpha2.ObjectName(name),
					},
					Weight: &weight,
				},
				// Append different headers to each split, which lets us identify
				// which backend we hit.
				Filters: []gatewayv1alpha2.HTTPRouteFilter{{
					Type: gatewayv1alpha2.HTTPRouteFilterRequestHeaderModifier,
					RequestHeaderModifier: &gatewayv1alpha2.HTTPRequestHeaderFilter{
						Set: []gatewayv1alpha2.HTTPHeader{{
							Name:  headerName,
							Value: name,
						}},
					}},
				}},
		)
		weights[name] = float64(percent)

		total += percent
		percent *= 2
		// Cap the final non-zero bucket so that we total 100%
		// After that, this will zero out remaining buckets.
		if total+percent > 100 {
			percent = 100 - total
		}
	}

	// Create a simple HTTPRoute over the 10 Services.
	name := test.ObjectNameForTest(t)
	_, client, _ := CreateHTTPRouteReady(ctx, t, clients, gatewayv1alpha2.HTTPRouteSpec{
		CommonRouteSpec: gatewayv1alpha2.CommonRouteSpec{ParentRefs: []gatewayv1alpha2.ParentReference{
			testGateway,
		}},
		Hostnames: []gatewayv1alpha2.Hostname{gatewayv1alpha2.Hostname(name + ".example.com")},
		Rules: []gatewayv1alpha2.HTTPRouteRule{{
			BackendRefs: backends,
		}},
	})

	// Create a large enough population of requests that we can reasonably assess how
	// well the Ingress respected the percentage split.
	seen := make(map[string]float64, len(backends))

	const (
		// The total number of requests to make (as a float to avoid conversions in later computations).
		totalRequests = 1000.0
		// The increment to make for each request, so that the values of seen reflect the
		// percentage of the total number of requests we are making.
		increment = 100.0 / totalRequests
		// Allow the Ingress to be within 10% of the configured value.
		margin = 10.0
	)
	var g errgroup.Group
	g.SetLimit(8)
	resultCh := make(chan string, totalRequests)

	for i := 0.0; i < totalRequests; i++ {
		g.Go(func() error {
			ri := RuntimeRequest(ctx, t, client, "http://"+name+".example.com")
			if ri == nil {
				return errors.New("failed to request")
			}
			resultCh <- ri.Request.Headers.Get(headerName)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Error("Error while sending requests:", err)
	}
	close(resultCh)

	for r := range resultCh {
		seen[r] += increment
	}

	for name, want := range weights {
		got := seen[name]
		switch {
		case want == 0.0 && got > 0.0:
			// For 0% targets, we have tighter requirements.
			t.Errorf("Target %q received traffic, wanted none (0%% target).", name)
		case math.Abs(got-want) > margin:
			t.Errorf("Target %q received %f%%, wanted %f +/- %f", name, got, want, margin)
		}
	}
}
