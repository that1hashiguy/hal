# HAL Creds Command Spec

## Base Command
- Command: `hal creds`
- Purpose: Inspect active lab credentials across running HAL services
- Default behavior: runs `hal creds status`

## Subcommands
- `hal creds status`
  - Display credentials for all currently running lab services
  - Only services that are actively running are shown

## Notes
- Credentials are read from live containers and cached state files — no credentials are stored by HAL beyond what the products themselves persist.
- For Vault in `prod` mode, the unseal key and root token are read from `~/.hal/vault-prod/init.json` (written by `hal vault create --mode prod`).

## Sources
- Namespace: `cmd/creds/creds.go`
