# HAL Terraform SAML Command Spec

## Command
- `hal terraform saml`

## Purpose
Deploy Authentik as a shared Identity Provider and wire up Terraform Enterprise SAML SSO.
Optionally configures SCIM provisioning from Authentik to TFE for group-based team management.

The same Authentik stack is reused by `hal vault oidc` — it is only torn down when no other
product has registered against it.

## Related
- Parent namespace: [terraform.md](terraform.md)
- Shared IdP: also used by `hal vault oidc` (Authentik is a singleton on `hal-net`)

## Prerequisites
- HAL CLI is available in your local environment.
- TFE must be running (`hal terraform create`) before enabling SAML SSO.

## Lifecycle Actions
`hal terraform saml` uses positional action arguments (or hidden flags):

| Action   | Description |
|----------|-------------|
| `status` | Show Authentik stack health and TFE SAML state (default) |
| `enable` | Start Authentik, provision demo users/groups, configure TFE SAML |
| `disable` | Remove TFE SAML and tear down Authentik if no other product uses it |
| `update` | Re-provision TFE SAML providers (keeps Authentik running) |

## Demo Users
Authentik is seeded with two demo identities:

| Username | Password  | Authentik Group | TFE Team |
|----------|-----------|-----------------|----------|
| alice    | password  | admins          | admins   |
| bob      | password  | devs            | devs     |

## Flags
```text
-e, --enable                          Start Authentik and configure TFE SAML SSO
-d, --disable                         Remove TFE SAML and tear down Authentik if unused
-u, --update                          Re-provision TFE SAML providers
    --scim                            Also configure SCIM provisioning from Authentik to TFE
    --sync                            With --scim: re-push all group membership to TFE without full re-provision
    --authentik-image string          Authentik container image
    --authentik-tag string            Authentik image tag
    --tfe-url string                  TFE base URL (default: https://tfe.localhost:8443 for primary)
    --tfe-org string                  TFE organization name (default "hal-org")
    --tfe-token string                TFE admin API token (auto-bootstrapped if omitted)
    --tfe-admin-username string       TFE admin username
    --tfe-admin-email string          TFE admin email
    --tfe-admin-password string       TFE admin password
-h, --help                            help for saml
```
- Global flags: `--debug`, `--dry-run`, `--verbose`

## Side Effects
- Starts the Authentik IdP containers on `hal-net` (shared singleton — also used by `hal vault oidc`).
- Configures TFE SAML SSO provider pointing to Authentik.
- With `--scim`: also wires SCIM endpoint so Authentik group changes automatically sync TFE teams.
- `disable` tears down Authentik only if `hal vault oidc` is not also using it.

## Example
```bash
# Check SAML status
hal terraform saml

# Enable SAML SSO
hal terraform saml enable

# Enable SAML SSO with SCIM provisioning
hal terraform saml enable --scim

# Re-sync group membership without full re-provision
hal terraform saml update --scim --sync

# Disable TFE SAML (tears down Authentik if unused)
hal terraform saml disable
```
