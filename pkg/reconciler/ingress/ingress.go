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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"knative.dev/networking/pkg/apis/networking/v1alpha1"
	ingressreconciler "knative.dev/networking/pkg/client/injection/reconciler/networking/v1alpha1/ingress"
	"knative.dev/networking/pkg/ingress"
	"knative.dev/networking/pkg/status"
	"knative.dev/pkg/network"
	pkgreconciler "knative.dev/pkg/reconciler"

	gwapiclientset "knative.dev/net-gateway-api/pkg/client/gatewayapi/clientset/versioned"
	gatewayListers "knative.dev/net-gateway-api/pkg/client/gatewayapi/listers/apis/v1alpha2"
	"knative.dev/net-gateway-api/pkg/reconciler/ingress/config"
)

const (
	notReconciledReason  = "ReconcileIngressFailed"
	notReconciledMessage = "Ingress reconciliation failed"
)

// Reconciler implements controller.Reconciler for Route resources.
type Reconciler struct {
	statusManager status.Manager

	gwapiclient gwapiclientset.Interface

	// Listers index properties about resources
	httprouteLister gatewayListers.HTTPRouteLister

	referencePolicyLister gatewayListers.ReferencePolicyLister

	gatewayLister gatewayListers.GatewayLister
}

var (
	_ ingressreconciler.Interface = (*Reconciler)(nil)
)

// ReconcileKind implements Interface.ReconcileKind.
func (c *Reconciler) ReconcileKind(ctx context.Context, ingress *v1alpha1.Ingress) pkgreconciler.Event {
	reconcileErr := c.reconcileIngress(ctx, ingress)

	if reconcileErr != nil {
		ingress.Status.MarkIngressNotReady(notReconciledReason, notReconciledMessage)
		return reconcileErr
	}

	return nil
}

// FinalizeKind implements Interface.FinalizeKind
func (c *Reconciler) FinalizeKind(ctx context.Context, ingress *v1alpha1.Ingress) pkgreconciler.Event {
	gatewayConfig := config.FromContext(ctx).Gateway

	// We currently only support TLS on the external IP
	return c.clearGatewayListeners(ctx, ingress, gatewayConfig.Gateways[v1alpha1.IngressVisibilityExternalIP].Gateway)
}

func (c *Reconciler) reconcileIngress(ctx context.Context, ing *v1alpha1.Ingress) error {
	gatewayConfig := config.FromContext(ctx).Gateway

	// We may be reading a version of the object that was stored at an older version
	// and may not have had all of the assumed defaults specified.  This won't result
	// in this getting written back to the API Server, but lets downstream logic make
	// assumptions about defaulting.
	ing.SetDefaults(ctx)
	before := ing.DeepCopy()

	ing.Status.InitializeConditions()

	if _, err := ingress.InsertProbe(ing); err != nil {
		return fmt.Errorf("failed to add knative probe header: %w", err)
	}

	for _, rule := range ing.Spec.Rules {
		rule := rule

		httproutes, err := c.reconcileHTTPRoute(ctx, ing, &rule)
		if err != nil {
			return err
		}

		if isHTTPRouteReady(httproutes) {
			ing.Status.MarkNetworkConfigured()
		} else {
			ing.Status.MarkIngressNotReady("HTTPRouteNotReady", "Waiting for HTTPRoute becomes Ready.")
		}
	}

	listeners := make([]*gatewayv1alpha2.Listener, 0, len(ing.Spec.TLS))
	for _, tls := range ing.Spec.TLS {
		tls := tls

		l, err := c.reconcileTLS(ctx, &tls, ing)
		if err != nil {
			return err
		}
		listeners = append(listeners, l...)
	}

	if len(listeners) > 0 {
		// For now, we only reconcile the external visibility, because there's
		// no way to provide TLS for internal listeners.
		err := c.reconcileGatewayListeners(
			ctx, listeners, ing, *gatewayConfig.Gateways[v1alpha1.IngressVisibilityExternalIP].Gateway)
		if err != nil {
			return err
		}
	}

	// TODO: check Gateway readiness before reporting Ingress ready

	ready, err := c.statusManager.IsReady(ctx, before)
	if err != nil {
		return fmt.Errorf("failed to probe Ingress: %w", err)
	}

	if ready {
		namespacedNameService := gatewayConfig.Gateways[v1alpha1.IngressVisibilityExternalIP].Service
		publicLbs := []v1alpha1.LoadBalancerIngressStatus{
			{DomainInternal: network.GetServiceHostname(namespacedNameService.Name, namespacedNameService.Namespace)},
		}

		namespacedNameLocalService := gatewayConfig.Gateways[v1alpha1.IngressVisibilityClusterLocal].Service
		privateLbs := []v1alpha1.LoadBalancerIngressStatus{
			{DomainInternal: network.GetServiceHostname(namespacedNameLocalService.Name, namespacedNameLocalService.Namespace)},
		}

		ing.Status.MarkLoadBalancerReady(publicLbs, privateLbs)
	} else {
		ing.Status.MarkLoadBalancerNotReady()
	}

	return nil
}

// isHTTPRouteReady will check the status conditions of the ingress and return true if
// all gateways have been admitted.
func isHTTPRouteReady(r *gatewayv1alpha2.HTTPRoute) bool {
	if r.Status.Parents == nil {
		return false
	}
	for _, gw := range r.Status.Parents {
		if !isGatewayAdmitted(gw) {
			// Return false if _any_ of the gateways isn't admitted yet.
			return false
		}
	}
	return true
}

func isGatewayAdmitted(gw gatewayv1alpha2.RouteParentStatus) bool {
	for _, condition := range gw.Conditions {
		if condition.Type == string(gatewayv1alpha2.RouteConditionAccepted) {
			return condition.Status == metav1.ConditionTrue
		}
	}
	return false
}
