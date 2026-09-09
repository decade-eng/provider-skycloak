# provider-skycloak

`provider-skycloak` is a [Crossplane](https://crossplane.io/) provider for
[Skycloak](https://skycloak.io/) (managed Keycloak), built with
[Upjet](https://github.com/crossplane/upjet) on top of the official
[sky-cloak/skycloak](https://registry.terraform.io/providers/sky-cloak/skycloak/latest)
Terraform provider. It exposes XRM-conformant managed resources for the
Skycloak public API.

The Terraform provider runs **in-process** (Terraform Plugin Framework over
protov6) — no Terraform CLI or provider binary at runtime. Because upstream
keeps its provider constructor in an `internal/` package, `make
provider-source` vendors the pinned upstream source into `third_party/`
(gitignored) and injects a one-file shim (`hack/xpshim.go.tmpl`) that legally
re-exports the constructor from inside the module boundary.

## Managed Resources

| API group | Kind | Terraform resource |
|---|---|---|
| `skycloak.crossplane.io` | `Cluster` | `skycloak_cluster` |
| `skycloak.crossplane.io` | `Realm` | `skycloak_realm` |
| `skycloak.crossplane.io` | `Domain` | `skycloak_domain` |
| `skycloak.crossplane.io` | `DomainRoute` | `skycloak_domain_route` |
| `skycloak.crossplane.io` | `CustomExtension` | `skycloak_custom_extension` |
| `skycloak.crossplane.io` | `ClusterExtension` | `skycloak_cluster_extension` |
| `skycloak.crossplane.io` | `Application` | `skycloak_application` |
| `skycloak.crossplane.io` | `ApplicationSecret` | `skycloak_application_secret` |
| `skycloak.crossplane.io` | `ApplicationRoleAssignment` | `skycloak_application_role_assignment` |
| `skycloak.crossplane.io` | `IdentityProvider` | `skycloak_identity_provider` |
| `skycloak.crossplane.io` | `SMTP` | `skycloak_smtp` |
| `skycloak.crossplane.io` | `LoginBranding` | `skycloak_login_branding` |
| `skycloak.crossplane.io` | `EmailBranding` | `skycloak_email_branding` |
| `skycloak.crossplane.io` | `ThemeAssignment` | `skycloak_theme_assignment` |
| `skycloak.crossplane.io` | `ClientThemeAssignment` | `skycloak_client_theme_assignment` |
| `skycloak.crossplane.io` | `CustomTheme` | `skycloak_custom_theme` |
| `skycloak.crossplane.io` | `ClusterSecurity` | `skycloak_cluster_security` |
| `skycloak.crossplane.io` | `ClusterMaintenanceWindow` | `skycloak_cluster_maintenance_window` |
| `skycloak.crossplane.io` | `CaptchaDomain` | `skycloak_captcha_domain` |
| `skycloak.crossplane.io` | `Export` | `skycloak_export` |
| `skycloak.crossplane.io` | `RealmExport` | `skycloak_realm_export` |
| `skycloak.crossplane.io` | `RealmImport` | `skycloak_realm_import` |
| `skycloak.crossplane.io` | `RealmGroup` | `skycloak_realm_group` |
| `skycloak.crossplane.io` | `RealmGroupMembership` | `skycloak_realm_group_membership` |
| `skycloak.crossplane.io` | `RealmRole` | `skycloak_realm_role` |
| `skycloak.crossplane.io` | `RealmUser` | `skycloak_realm_user` |
| `skycloak.crossplane.io` | `RealmUserRoleAssignment` | `skycloak_realm_user_role_assignment` |
| `skycloak.crossplane.io` | `SIEMDestination` | `skycloak_siem_destination` |
| `skycloak.crossplane.io` | `WebhookSubscription` | `skycloak_webhook_subscription` |

Namespaced variants of every kind are available under
`skycloak.m.crossplane.io` (`ProviderConfig` and `ClusterProviderConfig` live
there too).

## Authentication

The provider authenticates against the Skycloak API with a management API key
(minted in the Skycloak dashboard; stored as `SKYCLOAK_MANAGEMENT_KEY` in
Infisical at `/crossplane-system/crossplane-providers`). The key is stored in
a Kubernetes Secret as a JSON document:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: skycloak-creds
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: |
    {"api_key": "<SKYCLOAK_MANAGEMENT_KEY>"}
```

The optional `endpoint` and `api_version` provider attributes may be carried
in the same JSON (`{"api_key": "...", "endpoint": "...", "api_version": "..."}`).

Reference it from a ProviderConfig (cluster-scoped shown; namespaced
`ProviderConfig`/`ClusterProviderConfig` under `skycloak.m.crossplane.io`
are also generated):

```yaml
apiVersion: skycloak.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: skycloak-creds
      key: credentials
```

## Example

A custom domain on the staging cluster plus the realm route (the surfaces
issue 07 of the keycloak-migration needs):

```yaml
apiVersion: skycloak.crossplane.io/v1alpha1
kind: Domain
metadata:
  name: auth-internal-staging
spec:
  forProvider:
    clusterId: <skycloak-cluster-uuid>
    domain: internal.staging.decade.com
  providerConfigRef:
    name: default
---
apiVersion: skycloak.crossplane.io/v1alpha1
kind: DomainRoute
metadata:
  name: advisors-auth-internal-staging
spec:
  forProvider:
    clusterId: <skycloak-cluster-uuid>
    domainIdRef:
      name: auth-internal-staging
    realm: advisors
    allowAdminAccess: false
    corsAllowedOrigins:
      - https://advisor-web.internal.staging.decade.com
  providerConfigRef:
    name: default
```

After the Domain becomes Ready, create the DNS records returned in
`status.atProvider.dnsRecords` (TXT verification records + CNAME to
`status.atProvider.cnameTarget`) at the DNS provider; verification completes
out of band and is observable via `status.atProvider.verificationStatus`.

More examples are generated under [`examples-generated/`](examples-generated/).

## Developing

One-time setup (submodules + vendored provider source):

```console
make submodules
make provider-source
```

Run the code-generation pipeline:

```console
make generate
```

Run against a Kubernetes cluster (out of cluster):

```console
make run
```

Build, push, and install:

```console
make all
```

Build just the provider binary and the Crossplane package (xpkg) image:

```console
make build
```

Run the unit tests:

```console
make test
```

### Bumping the upstream Terraform provider

1. Update `TERRAFORM_PROVIDER_VERSION` (and `TERRAFORM_PROVIDER_COMMIT`, if
   pinning to a commit rather than the tag) in the `Makefile`.
2. Update `tfProviderVersion` in `config/provider.go` to match.
3. Run:

```console
make provider-source-clean generate
```

`make generate` re-fetches `config/schema.json` (via Terraform
`providers schema -json` against a filesystem mirror of the release
binaries), re-scrapes `config/provider-metadata.yaml` from the upstream docs,
and re-runs the upjet pipeline, controller-gen and angryjet. Review the
resulting diff of `package/crds/` for breaking CRD schema changes (CI runs
`make crddiff` on modified CRDs).

## Publishing

Packages are published to `ghcr.io/decade-eng/provider-skycloak` by the
`Publish Provider Package` workflow (workflow_dispatch with a version input),
mirroring the [crossplane-provider-cockroach](https://github.com/decade-eng/crossplane-provider-cockroach)
pipeline. The xpkg.upbound.io mirror is disabled (no credentials in this
org). Tags are created via the `Tag` workflow.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an issue.
