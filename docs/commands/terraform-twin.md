# HAL Terraform Twin Command Spec

## Command
- `hal terraform twin`
- Aliases: `hal terraform bis`, `hal terraform dup`

## Purpose
Manage a second local Terraform Enterprise instance ("twin") that runs alongside the primary
deployment. Useful for multi-instance TFE replication labs, migration scenarios, and comparing
configuration across two independent TFE clusters.

## Related
- Parent namespace: [terraform.md](terraform.md)
- Primary TFE deployment: [terraform-deploy.md](terraform-deploy.md)

## Prerequisites
- HAL CLI is available in your local environment.
- The primary TFE stack should be running (`hal terraform create`) before enabling the twin.
- Shares the primary's PostgreSQL and MinIO containers — those must be healthy.

## Lifecycle Actions
`hal terraform twin` uses positional action arguments (or hidden flags):

| Action   | Description |
|----------|-------------|
| `status` | Show twin TFE container and API health (default) |
| `enable` | Deploy the twin Terraform Enterprise instance |
| `disable` | Destroy the twin instance and remove its local artifacts |
| `update` | Reconcile the twin Terraform Enterprise instance in place |

## Flags
```text
-e, --enable                          Deploy the twin Terraform Enterprise instance
    --disable                         Destroy the twin Terraform Enterprise instance and remove its local artifacts
-u, --update                          Reconcile the twin Terraform Enterprise instance in place
    --twin-tag string                 TFE container image tag for the twin instance
    --twin-image string               TFE container image name for the twin instance
    --twin-password string            Twin TFE encryption password (default "tfe_password")
    --twin-tfe-org string             TFE organization name for the twin instance (default "hal-bis")
    --twin-tfe-project string         TFE project name for the twin instance (default "Dave-bis")
    --twin-tfe-admin-username string  Initial twin TFE admin username (default matches primary)
    --twin-tfe-admin-email string     Initial twin TFE admin email (default matches primary)
    --twin-tfe-admin-password string  Initial twin TFE admin password (default matches primary)
    --twin-proxy-tag string           Nginx image tag for the twin ingress proxy
    --twin-proxy-image string         Nginx image name for the twin ingress proxy
    --twin-https-port int             Host HTTPS port exposed by the twin TFE ingress proxy (default 9443)
    --twin-hostname string            TLS hostname used by the twin TFE instance
    --twin-container-name string      Container name for the twin TFE core app (default "hal-tfe-bis")
    --twin-proxy-ip string            Static internal proxy IP on hal-net (default: auto-derived .249)
    --twin-db-password string         PostgreSQL password for the twin TFE backend (default "tfe_password")
    --twin-db-name string             Database name for the twin TFE schema (default "tfe_bis")
    --twin-minio-root-user string     MinIO root user for shared object storage (default "minioadmin")
    --twin-minio-root-password string MinIO root password for shared object storage (default "minioadmin")
    --twin-s3-bucket string           S3 bucket name for twin TFE objects in shared MinIO (default "tfe-bis-data")
-h, --help                            help for twin
```
- Global flags: `--debug`, `--dry-run`, `--verbose`

## Side Effects
- Deploys a second TFE application container (`hal-tfe-bis`) and its nginx ingress proxy on `hal-net`.
- Reuses the existing shared PostgreSQL and MinIO containers (creates a new database/bucket within them).
- `disable` removes the twin container and its ingress proxy but leaves the shared backing services intact.

## Example
```bash
# Check twin status
hal terraform twin

# Deploy the twin TFE instance
hal terraform twin enable

# Deploy on a custom port
hal terraform twin enable --twin-https-port 9443

# Reconcile the twin in place
hal terraform twin update

# Tear down the twin
hal terraform twin disable
```
