# HAL Vault PKI Command Spec

## Command
- `hal vault pki`

## Purpose
Manage Vault PKI secrets engines — provisions a Root CA and an Intermediate CA, and optionally
deploys a working certificate demo using cert-manager on KinD or Vault's ACME endpoint with Caddy.

## Related
- Parent namespace: [vault.md](vault.md)

## Prerequisites
- HAL CLI is available in your local environment.
- Vault must be running (`hal vault create`) before enabling PKI.
- The `--k8s` and `--acme` demos additionally require a KinD cluster (created automatically).

## Lifecycle Actions
`hal vault pki` uses positional action arguments (or hidden flags):

| Action   | Description |
|----------|-------------|
| `status` | Show mount state, CA readiness, and demo health (default) |
| `enable` | Mount PKI engines, generate Root CA + Intermediate CA |
| `disable` | Disable PKI engines and remove all PKI resources |
| `update` | Reconcile PKI engines (recreate CAs) |

## Flags
```text
-e, --enable                         Enable PKI engines and generate Root CA + Intermediate CA
-d, --disable                        Disable PKI engines and remove all PKI resources
-u, --update                         Reconcile PKI engines (recreate CAs)
    --root-mount string              Vault mount path for the Root CA (default "pki-root")
    --int-mount string               Vault mount path for the Intermediate CA (default "pki-int")
    --root-ttl string                Max TTL for the Root CA (default "43800h" — 5 years)
    --int-ttl string                 Max TTL for the Intermediate CA (default "17520h" — 2 years)
    --allowed-domains string         Allowed domains for 'hal-role' (default "hal.local,cluster.local,svc.cluster.local")
    --max-cert-ttl string            Maximum TTL for leaf certs issued by 'hal-role' (default "24h")
    --k8s                            Deploy cert-manager + nginx web demo on KinD (enable/update only)
    --acme                           Deploy Vault ACME endpoint + Caddy demo on KinD (enable/update only)
    --force                          With --k8s/--acme update: also rebuild Root CA and Intermediate CA from scratch
    --kind-node-image string         KinD node image (default "kindest/node:v1.31.1")
    --cert-manager-version string    Jetstack cert-manager Helm chart version (empty = latest)
    --vault-pki-web-backend-image    Demo backend container image name for cert-manager/--k8s demo (default "nginx")
    --vault-pki-web-backend-tag      Demo backend container image tag (default "alpine")
    --vault-pki-caddy-image string   Caddy container image name for ACME/--acme demo (default "caddy")
    --vault-pki-caddy-tag string     Caddy container image tag (default "alpine")
    --acme-cert-ttl string           TTL for certs issued to Caddy via ACME (default "5m" — short for visible auto-renewal)
-h, --help                           help for pki
```
- Global flags: `--debug`, `--dry-run`, `--verbose`

## Side Effects
- Mounts `pki-root` and `pki-int` secrets engines in Vault and generates CA chain.
- With `--k8s`: creates a KinD cluster and deploys cert-manager + a signed nginx demo.
- With `--acme`: creates a KinD cluster and deploys a Caddy instance that auto-renews its certificate via Vault's ACME endpoint.
- `hal vault pki disable` unmounts both PKI engines and removes all generated certs/CAs.

## Example
```bash
# Check PKI status
hal vault pki

# Enable PKI engines (Root CA + Intermediate CA)
hal vault pki enable

# Enable PKI and deploy a cert-manager demo on KinD
hal vault pki enable --k8s

# Enable PKI and deploy an ACME/Caddy demo on KinD
hal vault pki enable --acme

# Reconcile CAs and rebuild the KinD demo from scratch
hal vault pki update --k8s --force

# Tear down PKI
hal vault pki disable
```
