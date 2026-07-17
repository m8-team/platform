---
title: "Platform Release Catalog"
---

# Platform Release Catalog

## Format

Release catalog entries are `PlatformRelease` resources:

```yaml
apiVersion: installer.m8.io/v1alpha1
kind: PlatformRelease
metadata:
  name: 1.0.0
spec:
  kubernetes:
    minVersion: "1.30.0"
    maxVersion: "1.34.x"
  signature:
    algorithm: cosign-bundle-v1
    value: sha256:...
    keyRef: m8-release-signing-key
  components:
    cilium:
      version: "1.17.0"
      chart:
        repository: oci://registry.m8.io/charts/cilium
        version: "1.17.0"
        digest: sha256:...
      images:
        - repository: registry.m8.io/cilium/cilium
          version: "1.17.0"
          digest: sha256:...
```

The example release is:

- `catalog/releases/1.0.0.yaml`

## Verification

MVP verification enforces:

- release schema validation;
- no floating component versions;
- chart and image digest format;
- non-empty signature unless `--allow-unsigned-release` is passed.

Production verification must add:

- Cosign signature verification;
- SBOM presence and checksum;
- vulnerability metadata policy checks;
- Kubernetes compatibility checks;
- supported upgrade path checks;
- offline bundle completeness checks.

## Prohibited Inputs

- `latest`
- `main`
- `master`
- charts without digest;
- images without digest;
- unsigned production release catalogs.

## Air-Gapped Use

`bundle export` must copy every artifact referenced by `PlatformRelease`, keep original digest metadata, write bundle manifest and checksums, then sign the bundle. `bundle import` must verify bundle signature and push artifacts to private registries before `plan` or `install`.

