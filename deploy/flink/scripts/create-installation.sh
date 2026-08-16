#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 flink-<name>" >&2
  exit 1
fi

installation_name=$1
if [[ ! ${installation_name} =~ ^flink-[a-z0-9]([-a-z0-9]*[a-z0-9])?$ ]]; then
  echo "installation name must be a DNS label beginning with flink-" >&2
  exit 1
fi

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
flink_root=$(cd -- "${script_dir}/.." && pwd)
source_installation="${flink_root}/installations/flink-1c"
target_installation="${flink_root}/installations/${installation_name}"
source_application="${flink_root}/argocd/installations/flink-1c.application.yaml"
target_application="${flink_root}/argocd/installations/${installation_name}.application.yaml"
argo_kustomization="${flink_root}/argocd/kustomization.yaml"

if [[ -e ${target_installation} || -e ${target_application} ]]; then
  echo "installation or Argo CD Application already exists: ${installation_name}" >&2
  exit 1
fi

mkdir -p "${target_installation}"
cp -R "${source_installation}/base" "${target_installation}/base"
cp -R "${source_installation}/overlays" "${target_installation}/overlays"

find "${target_installation}" -type f -exec \
  sed -i.bak "s/flink-1c/${installation_name}/g" {} +
find "${target_installation}" -type f -name '*.bak' -delete

sed -i.bak \
  's/username="1c-cr-cdc-consumer"/username="<KAFKA_USERNAME>"/' \
  "${target_installation}/base/kafka-secret.example.yaml"
rm -f -- "${target_installation}/base/kafka-secret.example.yaml.bak"

for sql_group in sources views sinks jobs; do
  mkdir -p "${target_installation}/sql/${sql_group}"
  touch "${target_installation}/sql/${sql_group}/.gitkeep"
done

cat >"${target_installation}/sql/kustomization.yaml" <<EOF
apiVersion: kustomize.config.k8s.io/v1alpha1
kind: Component

configMapGenerator:
  - name: ${installation_name}-sql

generatorOptions:
  annotations:
    argocd.argoproj.io/sync-wave: "0"
EOF

cat >"${target_installation}/README.md" <<EOF
# Flink installation: ${installation_name}

This installation owns namespace \`${installation_name}\`, its FlinkDeployment,
SQL Gateway, Kafka settings, object-storage state prefix, resources, Secrets, and
SQL jobs. It shares only the runtime image and Kubernetes operator.

Before enabling its Argo CD Application, define:

- the runtime repository and digest in \`base/kustomization.yaml\`;
- JobManager/TaskManager sizing and slots in \`base/flink-deployment.yaml\`;
- unique object-storage paths containing \`/${installation_name}/\`;
- Kafka brokers, topics, principal, and ACLs in installation-owned SQL/Secrets;
- business SQL under \`sql/{sources,views,sinks,jobs}\`;
- production secret delivery without committing plaintext credentials.
EOF

cp "${source_application}" "${target_application}"
sed -i.bak "s/flink-1c/${installation_name}/g" "${target_application}"
rm -f -- "${target_application}.bak"

printf '  - installations/%s.application.yaml\n' "${installation_name}" \
  >>"${argo_kustomization}"

echo "created ${target_installation}"
echo "created ${target_application}"
echo "next: replace copied runtime, sizing, Kafka, storage, Secret, and SQL settings"
