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

Regex path (NGINX Ingress) example

```bash
cd deploy/charts/helloworld
helm upgrade --install helloworld . \
  --namespace helloworld \
  --create-namespace \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set ingress.hosts[0].host=my.example.com \
  --set ingress.hosts[0].paths[0].path='/helloworld(/|$)(.*)' \
  --set ingress.hosts[0].paths[0].pathType=ImplementationSpecific \
  --set image.repository=ghcr.io/berndonline/k8s/go-helloworld \
  --set image.tag="latest"
```

Prefix path (no regex) example

```bash
cd deploy/charts/helloworld
helm upgrade --install helloworld . \
  --namespace helloworld \
  --create-namespace \
  --set ingress.enabled=true \
  --set ingress.className=nginx \
  --set ingress.annotations."nginx\.ingress\.kubernetes\.io/rewrite-target"="/" \
  --set ingress.hosts[0].host=my.example.com \
  --set ingress.hosts[0].paths[0].path=/helloworld \
  --set ingress.hosts[0].paths[0].pathType=Prefix \
  --set image.repository=ghcr.io/berndonline/k8s/go-helloworld \
  --set image.tag="latest"
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
  - `--set ingress.hosts[0].paths[0].pathType=ImplementationSpecific` (defaults; use `Prefix` for non-regex)
  - `--set ingress.className=nginx`
  - Regex matching is enabled by default (`nginx.ingress.kubernetes.io/use-regex: "true"`)
  - For non-regex prefix matching, override rewrite target: `--set ingress.annotations."nginx\.ingress\.kubernetes\.io/rewrite-target"="/"`

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
  - `--set kafka.brokers[0]=broker-1.kafka:9092` (repeat index for multiple brokers)
  - Optional client identifier: `--set kafka.clientId=helloworld`
  - Enable mTLS with Strimzi secrets:
    - `--set kafka.tls.enabled=true`
    - `--set kafka.tls.secretName=helloworld` (the `KafkaUser` name)
    - `--set kafka.tls.mountPath=/var/run/secrets/kafka` *(default)*
    - Optionally override Strimzi key names via `kafka.tls.caFile`, `kafka.tls.certFile`, `kafka.tls.keyFile`

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
