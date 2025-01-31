# Knative net-gateway-api
**[This component is ALPHA](https://github.com/knative/community/tree/main/mechanics/MATURITY-LEVELS.md)**

[![GoDoc](https://godoc.org/knative-sandbox.dev/net-gateway-api?status.svg)](https://godoc.org/knative.dev/net-gateway-api)
[![Go Report Card](https://goreportcard.com/badge/knative-sandbox/net-gateway-api)](https://goreportcard.com/report/knative-sandbox/net-gateway-api)

net-gateway-api repository contains a KIngress implementation and testing for Knative integration with the [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/).

This work is still in early development, which means it's _not ready for production_, but also that your feedback can have a big impact. You can find the tested Ingress and unavailable features [here](docs/test-version.md).

This work is still in early development, which means it's _not ready for production_, but also that your feedback can have a big impact.
You can also find the tested Ingress and unavailable features [here](docs/test-version.md).

## Tests
Note: conformance and e2e tests are a wip at the moment. Please see:

- [EPIC - Contour tests · Issue #36 · knative-sandbox/net-gateway-api](https://github.com/knative-sandbox/net-gateway-api/issues/36)
- [EPIC - Istio tests · Issue #23 · knative-sandbox/net-gateway-api](https://github.com/knative-sandbox/net-gateway-api/issues/23)

Versions to be installed are listed in [`hack/test-env.sh`](hack/test-env.sh).
---
## Requirements
1. A Kind cluster
1. Knative serving installed
2. [`ko`](https://github.com/google/ko) (for installing the net-gateway-api)
3. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
4. `export KO_DOCKER_REPO=kind.local`

## Getting started
### Install Knative serving
```bash
kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/latest/download/serving-core.yaml
```

#### Configure Knative
##### Ingress
Configuration so Knative serving uses the proper "ingress.class":

```bash
kubectl patch configmap/config-network \
  -n knative-serving \
  --type merge \
  -p '{"data":{"ingress.class":"gateway-api.ingress.networking.knative.dev"}}'
```

##### Load balancer
Configuration so Knative serving uses nip.io for DNS. For `kind` the loadbalancer IP is `127.0.0.1`:

```bash
kubectl patch configmap/config-domain \
  -n knative-serving \
  --type merge \
  -p '{"data":{"127.0.0.1.nip.io":""}}'
```

##### (OPTIONAL) Deploy a sample hello world app:
```bash
cat <<-EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: helloworld-go
spec:
  template:
    spec:
      containers:
      - image: gcr.io/knative-samples/helloworld-go
        env:
        - name: TARGET
          value: Go Sample v1
EOF
```

### Install net-gateway-api
```bash
ko apply -f config/
```

### Load tested environment versions
```
source ./hack/test-env.sh
```

### Install a supported implementation
#### Istio
```bash
# gateway-api CRD must be installed before Istio.
echo ">> Installing Gateway API CRDs"
kubectl apply -f config/100-gateway-api.yaml

echo ">> Bringing up Istio"
curl -sL https://istio.io/downloadIstioctl | sh -
"$HOME"/.istioctl/bin/istioctl install -y --set values.gateways.istio-ingressgateway.type=NodePort --set values.global.proxy.clusterDomain="${CLUSTER_SUFFIX}"

echo ">> Deploy Gateway API resources"
kubectl apply -f ./third_party/istio/gateway/
```

#### Contour
```bash
echo ">> Bringing up Contour"
kubectl apply -f "https://raw.githubusercontent.com/projectcontour/contour-operator/${CONTOUR_VERSION}/examples/operator/operator.yaml"

# wait for operator deployment to be Available
kubectl wait deploy --for=condition=Available --timeout=120s -n "contour-operator" -l '!job-name'

echo ">> Deploy Gateway API resources"
ko resolve -f ./third_party/contour/gateway/ | \
  sed 's/LoadBalancerService/NodePortService/g' | \
  kubectl apply -f -
```

### (OPTIONAL) For testing purpose (Istio)
You can use port-forwarding to make requests to ingress from your machine:

```bash
kubectl port-forward  -n istio-system $(kubectl get pod -n istio-system -l "app=istio-ingressgateway" --output=jsonpath="{.items[0].metadata.name}") 8080:8080

curl -v -H "Host: helloworld-go.default.127.0.0.1.nip.io" http://localhost:8080
```

---

To learn more about Knative, please visit our
[Knative docs](https://github.com/knative/docs) repository.

If you are interested in contributing, see [CONTRIBUTING.md](./CONTRIBUTING.md)
and [DEVELOPMENT.md](./DEVELOPMENT.md).
