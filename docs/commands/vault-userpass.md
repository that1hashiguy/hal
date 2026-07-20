# HAL Vault Userpass Command Spec

## Command
- `hal vault userpass`
- Alias: `hal vault up`

## Purpose
Configure Vault's userpass auth method with a demo user and token metadata policy.
Useful for quickly demonstrating username/password-based Vault authentication in local labs.

## Related
- Parent namespace: [vault.md](vault.md)

## Prerequisites
- HAL CLI is available in your local environment.
- Vault must be running (`hal vault create`) before enabling userpass auth.

## Lifecycle Actions
`hal vault userpass` uses positional action arguments (or hidden flags):

| Action   | Description |
|----------|-------------|
| `status` | Show whether userpass auth is enabled and list demo users (default) |
| `enable` | Enable userpass auth method and create a demo user |
| `disable` | Disable userpass auth method and remove its policy |
| `update` | Reconcile userpass auth method configuration |

## Flags
```text
-e, --enable     Enable userpass auth method and create a demo user
-d, --disable    Disable userpass auth method and remove its policy
-u, --update     Reconcile userpass auth method configuration
-h, --help       help for userpass
```
- Global flags: `--debug`, `--dry-run`, `--verbose`

## Side Effects
- Enables or disables the `userpass/` auth mount in Vault.
- Creates a demo user with an associated policy on enable.

## Example
```bash
# Check userpass status
hal vault userpass

# Enable userpass auth and create demo user
hal vault userpass enable

# Reconcile configuration
hal vault userpass update

# Disable and remove
hal vault userpass disable
```
