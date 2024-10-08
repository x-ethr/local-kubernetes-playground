# `local-kubernetes-playground`

Software engineers at ETHR previously used a variation of the following project as a playground for software
development, automation testing, research, and for demonstrating proof-of-concepts.

This playground was the motivation behind establishing `x-ethr` and its related open-source repositories.

> [!IMPORTANT]
> The following project requires an expansive amount of knowledge around development, kubernetes, and overall
> systems. While the guide can be followed step-by-step to produce a fully functioning cluster, there
> are [requirements](#requirements) that would otherwise be challenging for beginners to 1. understand, 2. setup, 3.
> debug.
>
> If requirements are correctly met, the entirety of this project can be deployed in under five minutes by simply
> following the [usage](#usage) section.
>
> Users of `local-kubernetes-playground` will involve themselves in the following disciplines:
> - Software Engineering
> - DevOps
> - Systems Administration
> - GitOps
> - Databases
> - Security

## Example

***The Playground's Deployed Service Mesh***

![istio-example](./.documentation/example-istio-service-mesh-3.png)

## Requirements

> [!IMPORTANT]
> Usage, requirements, and documentation was vetted on a Mac Studio, M1 Max 2022 on MacOS, Sonoma 14.5. Other systems
> are likely subject to incompatibilities.

###### System

- MacOS with Administrative Privileges
- [`go`](https://go.dev/doc/install)
- [`cloud-provider-kind`](https://github.com/kubernetes-sigs/cloud-provider-kind)
- [Homebrew](https://brew.sh)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [`istioctl`](https://istio.io/latest/docs/setup/getting-started/)
- [`ethr-cli`](https://github.com/x-ethr/ethr-cli)
- [`flux`](https://fluxcd.io/flux/get-started/)
- [`psql`](https://www.postgresql.org/download/)

###### Optional(s)

- [OpenLens](https://github.com/MuhammedKalkan/OpenLens) - Kubernetes UI Dashboard

## Usage

> [!NOTE]
> During the first minute or two, there may be a few warnings that surface. Due to Kubernetes reconciliation, all errors
> should resolve by minute three or four.

1. Install `kind`.
    ```bash
    go install sigs.k8s.io/kind@latest
    sudo install "$(go env --json | jq -r ".GOPATH")/bin/kind" /usr/local/bin
    ```
1. Create a cluster via `kind`.
    ```bash
    kind create cluster --config "configuration.yaml"
    kubectl config set-context "$(printf "%s-kind" "kind")"
    ```
2. Unable node(s).
    ```bash
    kubectl label node kind-control-plane node.kubernetes.io/exclude-from-external-load-balancers- 
    ```
3. Setup a local load-balancer (within its own private terminal session).
    ```bash
    go install sigs.k8s.io/cloud-provider-kind@latest
    sudo install "$(go env --json | jq -r ".GOPATH")/bin/cloud-provider-kind" /usr/local/bin
    cloud-provider-kind -v 9
    ```
4. Verify connectivity to the cluster.
    - If using OpenLens, select the `kind-kind` context.
5. Bootstrap.
    ```bash
    flux bootstrap github --repository "https://github.com/x-ethr/cluster-management" \
        --owner "x-ethr" \
        --private "false" \
        --personal "false" \
        --path "clusters/local" \
        --verbose
    ```
    - Requires a valid
      GitHub [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
      set as environment variables: `GITHUB_TOKEN`.
    - For users outside the `x-ethr` organization, fork, import, or copy
      the https://github.com/x-ethr/cluster-management repository; or use a customized Flux GitOps project.
6. Sync local cluster repository's `vendors`.
    ```bash
    git submodule update --remote --recursive
    ```
7. Add `kustomization.yaml` to new cluster directory (only applicable during first-time cluster setup).
    ```bash
    cat << EOF > ./vendors/cluster-management/clusters/local/kustomization.yaml
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources: []
    EOF
    ```
8. Optionally, update the `Kustomization.flux-system.spec.interval` (changes each time a local cluster is bootstrapped).
9. Push local changes to `vendors` submodules.
    ```bash
    git submodule foreach "git add . && git commit --message \"Git Submodule Update(s)\" && git push -u origin HEAD:main" 
    ```
10. Start the local registry.
     ```bash
     bash ./scripts/registry.bash
     ```
11. Wait for the various resources to reconcile successfully.

### Service-Mesh

*The following command will port-forward the gateway's configured port `80` and expose it on `localhost:8080`.*

```bash
kubectl port-forward --namespace development services/gateway 8080:80
```

###### Network Traffic

*In order to view tracing and network traffic, issue the following command(s)*:

```bash
for i in $(seq 1 250); do
    curl "http://localhost:8080/v1/test-service-1"
    curl "http://localhost:8080/v1/test-service-2"
    curl "http://localhost:8080/v1/test-service-2/alpha"
    
    curl "http://localhost:8080/v1/authentication"
done
```

###### Kiali

*The following command will expose the `kiali` service and open a browser to its dashboard.*

```bash
istioctl dashboard kiali
```

###### Tracing (Jaeger)

*The following command will expose the `jaeger` service and open a browser to its dashboard.*

```bash
istioctl dashboard jaeger
```

###### Istio & `istoctl`

*Useful `istoctl` command(s)*

```bash
kubectl -n istio-system logs --since=1h istiod-6bc5bc58b4-wvhmc --follow
```

###### Redis

*Useful `kubectl` command(s)*

**Logging**

```bash
kubectl --namespace caching logs --since=10m services/redis --follow
```

*Useful `redis-cli` command(s)*

```bash
redis-cli
```

**Add consumer to consumer group**

```bash
xadd demo-stream * name john email jdoe@test.com
xadd demo-stream * tom tom@test.com
```

## Contributions

Please see the [**Contributing Guide**](./CONTRIBUTING.md) file for additional details.

## Debugging

###### Ingress LB Address

```bash
kubectl get --namespace istio-system svc/ingress-gateway -o=jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

###### (upstream_reset_before_response_started{connection_termination})

Restart the Istio API Gateway Deployment

## External Reference(s)

- [Official Schema Store](https://github.com/SchemaStore/schemastore/tree/master/src/schemas/json)
    - [OpenAPI 3.1](https://raw.githubusercontent.com/OAI/OpenAPI-Specification/main/schemas/v3.1/schema.json)
    - [OpenAPI 3.0](https://raw.githubusercontent.com/OAI/OpenAPI-Specification/main/schemas/v3.0/schema.json)
- [AWS EKS, Crossplane, Flux Sample](https://github.com/aws-samples/eks-gitops-crossplane-flux/tree/main)
    - [Blog Reference](https://aws.amazon.com/blogs/containers/gitops-model-for-provisioning-and-bootstrapping-amazon-eks-clusters-using-crossplane-and-argo-cd/)
- [Istio By Example](https://istiobyexample.dev/grpc/)
- [Distributed Tracing](https://istio.io/latest/about/faq/distributed-tracing/)
- [Slog Guide](https://betterstack.com/community/guides/logging/logging-in-go/)
- [Tokens & Microservice(s)](https://fusionauth.io/articles/tokens/tokens-microservices-boundaries)
- [Golang Reverse Proxy](https://pkg.go.dev/net/http/httputil#ReverseProxy)
- [JWKS](https://github.com/coreos/go-oidc/blob/v3/oidc/jwks.go)
- [CoreOS OIDC](https://github.com/coreos/go-oidc/tree/v3)
- [Example of Microservice Communication](https://supertokens.com/static/6f6a1368b9082a0347063eed943d582b/78612/jwks-flow.png)
- [Redis Message Broker](https://semaphoreci.com/blog/redis-message-broker)
- [Kubernetes & Redis](https://www.dragonflydb.io/guides/redis-kubernetes)
- [Redis Streams](https://redis.io/docs/latest/develop/data-types/streams/)
