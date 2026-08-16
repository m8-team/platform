#!/usr/bin/env bash

set -euo pipefail

namespace=${1:-flink-1c}
gateway_name=${2:-${namespace}-sql-gateway}
default_brokers=(
  rc1a-an3ivliisaqgs0nb.mdb.yandexcloud.net:9091
  rc1d-93nrmjafghmd0gct.mdb.yandexcloud.net:9091
  rc1e-vrheh76ndbah9ciu.mdb.yandexcloud.net:9091
)

if ! command -v kubectl >/dev/null 2>&1; then
  echo "required command not found: kubectl" >&2
  exit 1
fi

kubectl -n "${namespace}" get secret flink-kafka-credentials >/dev/null
kubectl -n "${namespace}" rollout status "deployment/${gateway_name}" --timeout=2m
kubectl -n "${namespace}" exec "deployment/${gateway_name}" \
  -c sql-gateway -- test -r /etc/flink/secrets/kafka_client_jaas.conf

if [[ -n ${FLINK_KAFKA_BROKERS:-} ]]; then
  IFS=',' read -r -a brokers <<<"${FLINK_KAFKA_BROKERS}"
else
  brokers=("${default_brokers[@]}")
fi

for broker in "${brokers[@]}"; do
  echo "checking TCP connectivity to ${broker}"
  # The single-quoted program is evaluated by bash inside the Gateway pod.
  # shellcheck disable=SC2016
  kubectl -n "${namespace}" exec "deployment/${gateway_name}" \
    -c sql-gateway -- bash -ceu '
      broker=$1
      host=${broker%:*}
      timeout 10 openssl s_client \
        -connect "${broker}" \
        -servername "${host}" \
        -CAfile /usr/local/share/ca-certificates/Yandex/YandexInternalRootCA.crt \
        -verify_return_error </dev/null >/dev/null
    ' _ "${broker}"
done

echo "Kafka TLS connectivity and JAAS mount are available from ${gateway_name}"
