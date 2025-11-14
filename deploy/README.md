# Deployment

This repository includes a Helm chart for Kubernetes deployment under `deploy/charts/helloworld`.

## Deploy with Helm

Prerequisites:

- Kubernetes cluster and `kubectl` configured
- Helm 3 installed
- Container image published to a registry (default: `ghcr.io/berndonline/k8s/go-helloworld`)

Install or upgrade the chart:

```bash
cd deploy/charts/helloworld
helm upgrade --install helloworld . \
  --namespace helloworld \
  --create-namespace \
  --set image.repository=ghcr.io/berndonline/k8s/go-helloworld \
  --set image.tag="latest"
```

### Examples

Expose via Envoy Gateway (Gateway API HTTPRoute):

```bash
cd deploy/charts/helloworld
helm upgrade --install helloworld . \
  --namespace helloworld \
  --create-namespace \
  --set httpRoute.enabled=true \
  --set httpRoute.parentRefs[0].name=eg \
  --set httpRoute.parentRefs[0].namespace=envoy-gateway-system \
  --set httpRoute.hostnames[0]=my.example.com \
  --set image.repository=ghcr.io/berndonline/k8s/go-helloworld \
  --set image.tag="latest"
```

This assumes an Envoy Gateway instance listening on the `GatewayClass`/`Gateway` referenced by `parentRefs`. Update `hostnames` and `matches` as needed for your cluster.

### Options

- Image
  - `--set image.repository=...`
  - `--set image.tag=...` (leave empty to use chart `appVersion`)
  - `--set image.pullPolicy=IfNotPresent`

- Gateway API / HTTPRoute
  - `--set httpRoute.enabled=true`
  - `--set httpRoute.parentRefs[0].name=eg`
  - `--set httpRoute.parentRefs[0].namespace=envoy-gateway-system`
  - `--set httpRoute.hostnames[0]=my.example.com`
  - Override `httpRoute.matches` for custom path matching (defaults to `PathPrefix /`).
  - Supply an entire array of rule definitions via `httpRoute.rules` if you need custom filters/backends.

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

- DynamoDB backing store / AWS STS
  - `--set dynamodb.enabled=true`
  - `--set dynamodb.tableName=my-dynamodb-table`
  - `--set dynamodb.region=eu-west-1`
  - `--set dynamodb.roleArn=arn:aws:iam::123456789012:role/MyHelloWorldRole`
  - Optionally set a custom session name: `--set dynamodb.roleSessionName=helloworld`
  - When using IRSA, also provide the projected token path (usually `/var/run/secrets/eks.amazonaws.com/serviceaccount/token`) via `--set dynamodb.webIdentityTokenFile=...`
  - To have the chart project a service account token with a custom audience, enable `--set dynamodb.serviceAccountTokenProjection.enabled=true` and optionally tune:
    - `--set dynamodb.serviceAccountTokenProjection.mountPath=/var/run/secrets/eks.amazonaws.com/serviceaccount`
    - `--set dynamodb.serviceAccountTokenProjection.tokenFile=token`
    - `--set dynamodb.serviceAccountTokenProjection.audience=sts.amazonaws.com`
    - `--set dynamodb.serviceAccountTokenProjection.expirationSeconds=3600`
    - `--set dynamodb.serviceAccountTokenProjection.volumeName=aws-token`
  - Ensure the pod uses an IAM role that is permitted to assume `roleArn` (for example by enabling `serviceAccount.create` and binding it through IAM Roles for Service Accounts)

- Kafka producer
  - `--set kafka.enabled=true`
  - `--set kafka.topic=content-created`
  - `--set kafka.brokers[0]=my-cluster-kafka-bootstrap.kafka:9092` (point at your Strimzi bootstrap service; repeat the index for additional brokers if needed)
  - Optional client identifier: `--set kafka.clientId=helloworld`
  - Create the topic ahead of time (for Strimzi apply `deploy/strimzi/kafka-topic.yaml` in the Kafka namespace)

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
helm uninstall helloworld -n helloworld
```
