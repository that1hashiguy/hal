# PQC PKI interoperability

This file records observed interoperability for the `hal vault pki --pqc` lab
tracked in [issue #110](https://github.com/hashimiche/hal/issues/110). The
observations below are dated results, not a claim that every platform or later
code revision has been tested. See the [PKI command guide](docs/commands/vault-pki.md)
for setup, operation, and the detailed test narrative.

## Runtime and version record

The default workload image is the official NGINX image pinned in
[`cmd/vault/defaults.go`](cmd/vault/defaults.go):

```text
docker.io/library/nginx@sha256:1dd3048a04f4b76ebd706c1bbb9df7d9d53b4f8253b32ce14467088c9b5ada0f
```

| Validation | Vault | Kubernetes and host | Workload and host TLS runtime |
| --- | --- | --- | --- |
| 2026-10-20, existing cluster | `2.2.0+ent.hsm` | Kubernetes v1.31.1; macOS arm64 | Pinned stock NGINX 1.31.3 with OpenSSL 3.5.7 and bundled njs 1.0.0; host OpenSSL 3.6.3 |
| 2026-10-20, fresh deployment | `2.2.0+ent` in HAL dev mode | New Kubernetes v1.36.1 KinD cluster; Podman on macOS arm64 | Pinned stock NGINX image; separate component versions were not recorded for this run |
| 2026-10-22, in-place rotation | Existing production Vault HTTPS deployment | Existing KinD cluster; versions were not separately recorded | Existing stock NGINX workloads; versions were not separately recorded |

Vault must be licensed Enterprise 2.2.0 or later and expose ML-DSA-65 PKI
support. HAL checks its version, license, and live PKI API schema before
provisioning. The host OpenSSL must support ML-DSA X.509 certificate validation.
The container image must provide the NGINX SSL variables, OpenSSL algorithms,
and bundled HTTP njs module used by the lab. `--pqc-image` and
`--kind-node-image` allow image overrides; an override needs its own
interoperability run.

## Observed behavior

- On the existing v1.31.1 cluster, Vault-issued ML-DSA-65 client and server
  identities authenticated between the two NGINX pods. The same TLS 1.3
  connection negotiated `X25519MLKEM768`. The browser-to-frontend hop used a
  separate conventional ECDSA identity.
- Stock NGINX accepted a valid client and rejected missing, untrusted,
  expired, wrong-EKU, and absent-EKU clients. It also rejected an incorrect
  backend hostname, an untrusted backend root, and classical-only groups.
  Rejected requests did not reach the authenticated backend handler. NGINX may
  complete TLS and return HTTP 400 for certificate rejection instead of
  sending a TLS alert.
- On the fresh v1.36.1 KinD cluster, issuance and offline verification passed.
  A certificate-verified POST through host port 8093 confirmed both `mldsa65`
  handshake signatures and `X25519MLKEM768` on the same frontend-to-backend
  connection. The full CLI and stock NGINX positive and negative suite passed.
- On 2026-10-22, twelve certificate-verified requests over roughly twelve
  seconds succeeded during automatic rotation of ten-second certificates.
  Three frontend serials were observed, changing about every five seconds.
  Each request confirmed both `mldsa65` signatures, `X25519MLKEM768`, verified
  client identity, and frontend chain and hostname verification. Runtime logs
  confirmed recurring backend issuance; no CA rebuild was needed.

## Reproduce the integration check

With a running licensed Vault Enterprise 2.2.0+, host OpenSSL with ML-DSA X.509
support, `curl`, and a container engine, run from the repository root:

```sh
HAL_PQC_INTEGRATION=1 go test ./cmd/vault -run '^TestPQCIntegration$' -v -count=1
```

The test uses `VAULT_ADDR`, `VAULT_TOKEN`, and `VAULT_CACERT`, falling back to
HAL production state. `HAL_PQC_TEST_ENGINE` selects the engine (default
`podman`), and `HAL_PQC_TEST_IMAGE` selects an image instead of the pinned
default. It creates process-specific diagnostic Vault mounts and removes only
the mounts it created. It does not create or tear down KinD. Ordinary
`go test ./...` skips this live integration check unless
`HAL_PQC_INTEGRATION=1` is set.

## Limits of the evidence

- The dated runs above are historical. They do not establish that the latest
  commit has passed a fresh live Vault and Kubernetes interoperability run.
- The initial fresh-cluster rollout encountered stale container-runtime aliases
  after host image import. The node alias was repaired, and HAL was changed to
  let Kubernetes pull the image directly. The fixed code was not subsequently
  retested by creating another cluster from scratch.
- Interactive browser automation and destructive live teardown were not
  performed in these runs. The rotation run reused an existing deployment.
- No result is recorded here for other operating systems, Kubernetes versions,
  Vault versions, container engines, or image overrides. A passing Go unit
  suite or GitHub Actions run does not establish those runtime combinations.
