# Strimzi Integration

This folder contains sample manifests to provision the Kafka resources that the application depends on when running in a Strimzi-managed cluster.

## Prerequisites

- Strimzi operator (and a Kafka cluster) already deployed in the target namespace
- Namespace-scoped Strimzi custom resources enabled (`Kafka`, `KafkaTopic`, `KafkaUser`, ...)

## Provision Topic and User

1. Adjust the namespace, Kafka cluster name, and topic name inside the manifests if needed.
2. Apply the sample manifests:

```bash
kubectl apply -f deploy/strimzi/kafka-topic.yaml
kubectl apply -f deploy/strimzi/kafka-user.yaml
```

The Strimzi operator will:

- create the `helloworld` topic (feel free to change the spec for partitions/retention)
- create a mutual-TLS `KafkaUser` named `helloworld`
- project the user credentials into a secret named `helloworld` that exposes `ca.crt`, `user.crt`, and `user.key`

## Deploy the Application

Point the Helm chart at the Strimzi-managed brokers and mount the generated secret:

```bash
cd deploy/charts/helloworld
helm upgrade --install helloworld . \
  --namespace helloworld \
  --create-namespace \
  --set kafka.enabled=true \
  --set kafka.topic=helloworld \
  --set kafka.brokers[0]=my-kafka-kafka-bootstrap.kafka:9093 \
  --set kafka.tls.enabled=true \
  --set kafka.tls.secretName=helloworld
```

The default TLS mount path (`/var/run/secrets/kafka`) along with the Strimzi key names (`ca.crt`, `user.crt`, `user.key`) matches what Strimzi creates. If you customise the secret layout, override `kafka.tls.mountPath`, `kafka.tls.caFile`, `kafka.tls.certFile`, or `kafka.tls.keyFile` accordingly.

> The application disables Kafka auto-topic-creation, so ensure the Strimzi `KafkaTopic` resource exists (or change the writer settings if you prefer broker-side auto creation).
