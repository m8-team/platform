#!/usr/bin/env bash

set -euo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
flink_root=$(cd -- "${script_dir}/.." && pwd)

for required_command in helm kubectl; do
  if ! command -v "${required_command}" >/dev/null 2>&1; then
    echo "required command not found: ${required_command}" >&2
    exit 1
  fi
done

work_dir=$(mktemp -d)
cleanup() {
  rm -rf -- "${work_dir}"
}
trap cleanup EXIT

mkdir -p "${work_dir}/operator" "${work_dir}/helm-cache"
cp "${flink_root}/shared/operator/Chart.yaml" \
  "${flink_root}/shared/operator/Chart.lock" \
  "${flink_root}/shared/operator/values.yaml" \
  "${work_dir}/operator/"
export HELM_REPOSITORY_CONFIG="${work_dir}/repositories.yaml"
export HELM_REPOSITORY_CACHE="${work_dir}/helm-cache"
helm repo add flink-operator-1-15 \
  https://archive.apache.org/dist/flink/flink-kubernetes-operator-1.15.0/ \
  >/dev/null
helm dependency build --skip-refresh "${work_dir}/operator" >/dev/null
helm lint "${work_dir}/operator"
helm template flink-operator "${work_dir}/operator" \
  --namespace flink-operator \
  --include-crds >"${work_dir}/operator.yaml"
kubectl apply --dry-run=client --validate=false \
  -f "${work_dir}/operator.yaml" >/dev/null

shopt -s nullglob
installation_overlays=("${flink_root}"/installations/*/overlays/*)
shopt -u nullglob

if [[ ${#installation_overlays[@]} -eq 0 ]]; then
  echo "no Flink installation overlays found" >&2
  exit 1
fi

for overlay in "${installation_overlays[@]}"; do
  overlay_name=$(basename -- "${overlay}")
  installation_name=$(basename -- "$(dirname -- "$(dirname -- "${overlay}")")")
  rendered_file="${work_dir}/${installation_name}-${overlay_name}.yaml"

  kubectl kustomize "${overlay}" >"${rendered_file}"
  kubectl apply --dry-run=client --validate=false \
    -f "${rendered_file}" >/dev/null

  runtime_images=$(
    awk '$1 == "image:" {print $2}' "${rendered_file}" | sort -u
  )
  runtime_image_count=$(awk 'NF {count++} END {print count+0}' <<<"${runtime_images}")
  if [[ ${runtime_image_count} -ne 1 ]]; then
    echo "${installation_name}/${overlay_name}: expected one runtime image, got ${runtime_image_count}" >&2
    exit 1
  fi
  if [[ ${runtime_images} != *@sha256:* ]]; then
    echo "${installation_name}/${overlay_name}: runtime image is not digest-pinned" >&2
    exit 1
  fi
done

kubectl kustomize "${flink_root}/argocd" >"${work_dir}/argocd.yaml"
kubectl apply --dry-run=client --validate=false \
  -f "${work_dir}/argocd.yaml" >/dev/null

while IFS= read -r -d '' secret_example; do
  kubectl apply --dry-run=client --validate=false \
    -f "${secret_example}" >/dev/null
done < <(find "${flink_root}/installations" -type f -name '*secret.example.yaml' -print0)

source_sql="${flink_root}/installations/flink-1c/sql/sources/price-types.sql"
view_sql="${flink_root}/installations/flink-1c/sql/views/price-types-normalized.sql"
job_sql="${flink_root}/installations/flink-1c/sql/jobs/price-types.sql"

grep -Fq "'connector' = 'upsert-kafka'" "${source_sql}"
# Backticks are literal Flink SQL identifier quotes.
# shellcheck disable=SC2016
grep -Fq 'PRIMARY KEY (`_IDRRef`) NOT ENFORCED' "${source_sql}"
grep -Fq "'key.format' = 'json'" "${source_sql}"
grep -Fq "'value.format' = 'json'" "${source_sql}"
grep -Fq "'value.fields-include' = 'EXCEPT_KEY'" "${source_sql}"
grep -Fq "WHEN 'AA==' THEN FALSE" "${view_sql}"
grep -Fq "WHEN 'AQ==' THEN TRUE" "${view_sql}"
grep -Fq 'FROM price_types_normalized' "${job_sql}"

while IFS= read -r -d '' sql_file; do
  final_line=$(awk 'NF {line=$0} END {print line}' "${sql_file}")
  if [[ ${final_line} != *';' ]]; then
    echo "SQL file does not end with a semicolon: ${sql_file}" >&2
    exit 1
  fi
done < <(find "${flink_root}/installations" -type f -name '*.sql' -print0)

while IFS= read -r -d '' base_kustomization; do
  if grep -Eq 'kafka-secret\.example|object-storage-secret\.example' \
    "${base_kustomization}"; then
    echo "example Secrets must not be included in Kustomize: ${base_kustomization}" >&2
    exit 1
  fi
done < <(find "${flink_root}/installations" -path '*/base/kustomization.yaml' -print0)

if command -v shellcheck >/dev/null 2>&1; then
  shellcheck "${script_dir}"/*.sh
fi

echo "Flink GitOps validation passed"
