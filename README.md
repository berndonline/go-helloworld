# go-helloworld

Simple Go REST API service demonstrating best practices for HTTP routing, metrics, tracing, and basic auth/JWT. It exposes:
- Root handler returning a message and hostname
- REST API under `/api` with basic auth (`/v1`) and JWT (`/v2`)
- Prometheus metrics on `/metrics` (internal port)
- Optional reverse proxy routes under `/proxy`

This repo includes a Helm chart for Kubernetes deployment.

## Deploy with Helm

Prerequisites:
- Kubernetes cluster and `kubectl` configured
- Helm 3 installed
- Container image published to a registry (default: `ghcr.io/berndonline/k8s/go-helloworld`)

Add or update values as needed and install:

```bash
cd deploy/charts/helloworld
helm upgrade --install helloworld . \
  --namespace default \
  --create-namespace \
  --set image.repository=ghcr.io/berndonline/k8s/go-helloworld \
  --set image.tag="v0.0.1"
```

### Options

- Image
  - `--set image.repository=...`
  - `--set image.tag=...` (leave empty to use chart `appVersion`)
  - `--set image.pullPolicy=IfNotPresent`

- Ingress
  - `--set ingress.enabled=true`
  - `--set ingress.hosts[0].host=my.example.com`
  - `--set ingress.hosts[0].paths[0].path=/helloworld(/|$)(.*)`
  - `--set ingress.className=nginx`

- Metrics (Prometheus Operator / kube-prometheus-stack)
  - `--set metrics.enabled=true`
  - Optional RBAC (if your Prometheus requires it within the app namespace):
    - `--set metrics.rbacEnabled=true`
    - `--set metrics.prometheusServiceAccount=prometheus-k8s`
    - `--set metrics.prometheusNamespace=monitoring`
  - Optional alert rules (edit `values.yaml` under `metrics.rules`)

- Tracing (Jaeger agent sidecar)
  - `--set tracing.enabled=true`
  - `--set tracing.collectorArgs={"--reporter.grpc.host-port=dns:///jaeger-collector-headless.observability:14250","--reporter.type=grpc"}`

- Resources and Security
  - Defaults are set in `values.yaml` (requests/limits)
  - Container and pod security contexts are enabled by default; override via:
    - `--set containerSecurityContext.readOnlyRootFilesystem=false`
    - `--set podSecurityContext.runAsNonRoot=true`

- Service Account
  - `--set serviceAccount.create=true` to create a SA for the deployment
  - `--set serviceAccount.name=my-sa` to use an existing SA

- Pod Disruption Budget
  - `--set pdb.enabled=true` (set `pdb.minAvailable` as needed)

### Uninstall

```bash
helm uninstall helloworld -n default
```

## Local Development

Build the container from repo root:

```bash
docker build -f build/docker/Dockerfile -t ghcr.io/berndonline/k8s/go-helloworld:dev .
```

Run locally:

```bash
docker run --rm -p 8080:8080 -p 9100:9100 ghcr.io/berndonline/k8s/go-helloworld:dev
```

## Repository Layout

- `cmd/helloworld`: app entrypoint
- `internal/app`: server, api, auth, mongo, metrics, tracing, proxy
- `web/static`: static assets
- `build/docker/Dockerfile`: container build
- `deploy/charts/helloworld`: Helm chart
