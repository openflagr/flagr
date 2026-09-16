# Flagr Helm chart

Minimal chart: Deployment + ClusterIP Service + health probes. Configure Flagr with `env` / `envFrom` — see [Environment variables](https://openflagr.github.io/flagr/flagr_env).

```bash
helm install flagr oci://ghcr.io/openflagr/flagr/charts/flagr --version 1.0.0 \
  --namespace flagr --create-namespace
kubectl -n flagr port-forward svc/flagr 18000:18000
curl -sS http://127.0.0.1:18000/api/v1/health
```

From a checkout: `helm install flagr ./helm --namespace flagr --create-namespace`.

Default is one replica, SQLite at `/data/flagr.sqlite` on an emptyDir. SQLite is **one writer** — do not raise `replicaCount`. For more eval traffic with SQLite, set `evalReplicas.replicaCount` (json_http readers of the primary). For MySQL/Postgres, raise `replicaCount` and leave `evalReplicas` at 0.

Scaling: https://openflagr.github.io/flagr/flagr_self_host#kubernetes-scaling
Self-hosting: https://openflagr.github.io/flagr/flagr_self_host
