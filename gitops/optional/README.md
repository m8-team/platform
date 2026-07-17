# Optional GitOps Applications

This directory contains Argo CD `Application` manifests for platform dependencies that are not installed by default.

Apply an optional application only when the corresponding platform feature needs it.

```bash
kubectl apply -f gitops/optional/<component>/application.yaml
```

Optional applications target paths under `gitops/components/system`.
