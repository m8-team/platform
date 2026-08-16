SHELL := /bin/bash

FLINK_RUNTIME_REPOSITORY ?= cr.yandex/crpbq0asshi5u8nd7vks/docker/flink/flink-kafka
FLINK_RUNTIME_TAG ?= 2.2.1-local
FLINK_IMAGE_CONTEXT := shared/image
FLINK_OPERATOR_REPOSITORY := https://archive.apache.org/dist/flink/flink-kubernetes-operator-1.15.0/

.PHONY: validate render image-build operator-dependency sql-check

validate:
	scripts/validate.sh

render: operator-dependency
	helm template flink-operator shared/operator --namespace flink-operator >/dev/null
	@for overlay in installations/*/overlays/*; do \
		kubectl kustomize "$$overlay" >/dev/null; \
	done
	kubectl kustomize argocd >/dev/null

image-build:
	docker build --pull \
		--tag "$(FLINK_RUNTIME_REPOSITORY):$(FLINK_RUNTIME_TAG)" \
		"$(FLINK_IMAGE_CONTEXT)"
	scripts/check-sql-syntax.sh \
		"$(FLINK_RUNTIME_REPOSITORY):$(FLINK_RUNTIME_TAG)" \
		installations/flink-1c/sql

sql-check:
	scripts/check-sql-syntax.sh \
		"$(FLINK_RUNTIME_REPOSITORY):$(FLINK_RUNTIME_TAG)" \
		installations/flink-1c/sql

operator-dependency:
	helm repo add flink-operator-1-15 "$(FLINK_OPERATOR_REPOSITORY)" \
		--force-update >/dev/null
	helm dependency build --skip-refresh shared/operator
