---
title: Contributing
description: How to contribute to Flagr — issues, pull requests, and review.
---

# Contributing

Thanks for helping improve Flagr. This page is the contributor process: issues, pull requests, and review.

Local setup (including Windows): [Develop Flagr](index.md#develop-flagr). Build, test, and code layout: **[AGENTS.md](https://github.com/openflagr/flagr/blob/main/AGENTS.md)** (`make help` from the repo root). Tests: [Testing](flagr_testing.md).

Please follow the [Code of Conduct](https://github.com/openflagr/flagr/blob/main/CODE_OF_CONDUCT.md).

## Ways to contribute

- **Issues** — bug reports and feature ideas on [GitHub Issues](https://github.com/openflagr/flagr/issues). Search existing issues first; the [issue template](https://github.com/openflagr/flagr/blob/main/.github/ISSUE_TEMPLATE.md) asks for expected vs current behavior, repro steps, and your environment (`flagr version`).
- **Documentation** — user-facing pages in `docs/` (this site).
- **Code** — Go server (`pkg/`), Vue UI (`browser/flagr-ui/`), and tests.

For a large change, open an issue first so the approach can be discussed.

## Pull requests

1. Fork the repo and branch off `main`.
2. Keep the change focused: one problem per PR.
3. Fill in the [pull request template](https://github.com/openflagr/flagr/blob/main/.github/PULL_REQUEST_TEMPLATE.md).
4. Run the checks that match what you changed (see **Before commit / push** in [AGENTS.md](https://github.com/openflagr/flagr/blob/main/AGENTS.md)).
5. Open a PR against `main`.

Maintainers review PRs. Questions and requested changes are part of that review.

## Helm chart

The official chart lives in **`helm/`** (not `charts/flagr`). It is a small Deployment + ClusterIP Service; Flagr config is `env` / `envFrom` against [flagr_env.md](flagr_env.md).

```bash
make helm-lint       # helm lint --strict + helm template
make helm-unittest   # helm-unittest plugin v1.1.2 (CI installs it)
```

Kind `helm install` + `helm test` run in `.github/workflows/helm.yml` only (path-filtered on `helm/**`). Do not add `helm.yml` as a required GitHub check while `on.paths` skips Go PRs.

**Release rule:** every Flagr GitHub Release PR bumps `helm/Chart.yaml` `appVersion` to the new Flagr tag and bumps chart `version` patch (even if templates are unchanged), so a later OCI publish does not reuse a chart version.

## License

Flagr is Apache 2.0. Contributions are licensed under the same terms.
