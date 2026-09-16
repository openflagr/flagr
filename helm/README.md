# Flagr Helm chart

Minimal chart: Deployment + ClusterIP Service + health probes. Configure Flagr with `env` / `envFrom` — see [Environment variables](https://openflagr.github.io/flagr/flagr_env).

```bash
helm install flagr oci://ghcr.io/openflagr/flagr/charts/flagr --version 1.0.0 \
  --namespace flagr --create-namespace
kubectl -n flagr port-forward svc/flagr 18000:18000
curl -sS http://127.0.0.1:18000/api/v1/health
```

From a checkout: `helm install flagr ./helm --namespace flagr --create-namespace`.

| Mode | Values | Pods |
|------|--------|------|
| SQLite (default) | — | 1 writer |
| SQLite HA | `evalReplicas.replicaCount` | 1 writer + N json_http readers of the primary |
| GitOps | `gitops.enabled` + `gitops.flagsURL` | N identical json_http pods (GitHub raw URL) |
| MySQL / Postgres | `replicaCount` + `env` | N pods on one DB |

`evalReplicas` and `gitops` are mutually exclusive. Details: [Kubernetes scaling](https://openflagr.github.io/flagr/flagr_self_host#kubernetes-scaling).

## Releasing

[`cd_helm.yml`](../.github/workflows/cd_helm.yml) packages and pushes to `oci://ghcr.io/openflagr/flagr/charts/flagr`, and fails if that chart `version` already exists on GHCR. Bump `version` (chart) on every chart change, and on every Flagr release bump `appVersion` to the new Flagr tag plus at least a `version` patch — even when templates are unchanged.

**One-time:** make the GHCR package public (Package settings → Change visibility). Anonymous `helm install oci://…` returns 401 until then.
