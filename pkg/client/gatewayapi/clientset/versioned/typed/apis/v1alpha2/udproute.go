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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha2

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	scheme "knative.dev/net-gateway-api/pkg/client/gatewayapi/clientset/versioned/scheme"
	v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// UDPRoutesGetter has a method to return a UDPRouteInterface.
// A group's client should implement this interface.
type UDPRoutesGetter interface {
	UDPRoutes(namespace string) UDPRouteInterface
}

// UDPRouteInterface has methods to work with UDPRoute resources.
type UDPRouteInterface interface {
	Create(ctx context.Context, uDPRoute *v1alpha2.UDPRoute, opts v1.CreateOptions) (*v1alpha2.UDPRoute, error)
	Update(ctx context.Context, uDPRoute *v1alpha2.UDPRoute, opts v1.UpdateOptions) (*v1alpha2.UDPRoute, error)
	UpdateStatus(ctx context.Context, uDPRoute *v1alpha2.UDPRoute, opts v1.UpdateOptions) (*v1alpha2.UDPRoute, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.UDPRoute, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.UDPRouteList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.UDPRoute, err error)
	UDPRouteExpansion
}

// uDPRoutes implements UDPRouteInterface
type uDPRoutes struct {
	client rest.Interface
	ns     string
}

// newUDPRoutes returns a UDPRoutes
func newUDPRoutes(c *GatewayV1alpha2Client, namespace string) *uDPRoutes {
	return &uDPRoutes{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the uDPRoute, and returns the corresponding uDPRoute object, and an error if there is any.
func (c *uDPRoutes) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.UDPRoute, err error) {
	result = &v1alpha2.UDPRoute{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("udproutes").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of UDPRoutes that match those selectors.
func (c *uDPRoutes) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.UDPRouteList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha2.UDPRouteList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("udproutes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested uDPRoutes.
func (c *uDPRoutes) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("udproutes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a uDPRoute and creates it.  Returns the server's representation of the uDPRoute, and an error, if there is any.
func (c *uDPRoutes) Create(ctx context.Context, uDPRoute *v1alpha2.UDPRoute, opts v1.CreateOptions) (result *v1alpha2.UDPRoute, err error) {
	result = &v1alpha2.UDPRoute{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("udproutes").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(uDPRoute).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a uDPRoute and updates it. Returns the server's representation of the uDPRoute, and an error, if there is any.
func (c *uDPRoutes) Update(ctx context.Context, uDPRoute *v1alpha2.UDPRoute, opts v1.UpdateOptions) (result *v1alpha2.UDPRoute, err error) {
	result = &v1alpha2.UDPRoute{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("udproutes").
		Name(uDPRoute.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(uDPRoute).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *uDPRoutes) UpdateStatus(ctx context.Context, uDPRoute *v1alpha2.UDPRoute, opts v1.UpdateOptions) (result *v1alpha2.UDPRoute, err error) {
	result = &v1alpha2.UDPRoute{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("udproutes").
		Name(uDPRoute.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(uDPRoute).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the uDPRoute and deletes it. Returns an error if one occurs.
func (c *uDPRoutes) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("udproutes").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *uDPRoutes) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("udproutes").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched uDPRoute.
func (c *uDPRoutes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.UDPRoute, err error) {
	result = &v1alpha2.UDPRoute{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("udproutes").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
