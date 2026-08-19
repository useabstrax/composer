# Abstrax Composer Plugin

Official Abstrax CLI plugin for **installing and running Composer**.

Binary: `abstrax-composer` → `abstrax composer …`  
Trust level: `official`  
OS: Debian/Ubuntu and RHEL-family (Rocky/Alma/RHEL).

## What it does

Abstrax installs PHP for projects; this plugin installs **Composer itself**, keeps it up to date, and runs it with the right PHP binary and user.

- Downloads the stable Composer phar from getcomposer.org and verifies the SHA-256 checksum
- Installs it globally at `/usr/local/bin/composer` (wrapper) and `/usr/local/lib/abstrax/composer/composer.phar`
- Optional default PHP when you are not in an Abstrax project (`php8.2` instead of unversioned `php`)
- `composer run` works in any directory; `--project` supplies path, user, and the project's PHP version

It does not wrap every Composer subcommand. Pass Composer arguments through `abstrax composer run`.

## Install (local)

```bash
cd plugins/composer
go build -o bin/abstrax-composer ./cmd/abstrax-composer
sudo cp bin/abstrax-composer /usr/local/lib/abstrax/plugins/
abstrax composer version
abstrax-composer plugin metadata | jq .
```

Release builds (linux-amd64 / linux-arm64 archives + `plugin-manifest.json`) are produced by the GitHub Actions release workflow when you push a `v*` tag.

## Quick start

```bash
sudo abstrax composer setup
abstrax composer status

# Optional: use a versioned PHP when not in an Abstrax project
sudo abstrax composer configure --php=php8.2

abstrax composer run install --no-dev --optimize-autoloader
abstrax composer run --project=example.com install --no-dev
```

## Commands

| Command | Root? | Description |
|---------|-------|-------------|
| `composer setup` | yes | Download, verify, and install Composer globally |
| `composer self-update` | yes | Replace the phar with the latest stable release |
| `composer remove` | yes | Remove the wrapper and phar (`--purge` also removes config) |
| `composer status` | no | Installed?, paths, Composer version, resolved PHP |
| `composer configure` | yes (writes) | Show or set the default PHP binary |
| `composer run [args…]` | no* | Run Composer with the resolved PHP binary |
| `composer diagnose` | no | Check PHP, Composer, git, unzip, and extensions |
| `composer auth` | yes (writes) | Show or update `auth.json` for a user |

\*Running as root without `--user` / `--project` / `--allow-root` is refused (or dropped to `SUDO_USER` when sudo was used).

Globals: `--json`, `--json-stream`, `--yes`, `--dry-run`, `--verbose`, `--quiet`, `--no-color`.

## PHP resolution

1. `--php` on this invocation
2. `ABSTRAX_COMPOSER_PHP`
3. `--project` and the project runtime is PHP → versioned CLI (`php8.5`, Remi path)
4. Default in `/etc/abstrax/composer.json`
5. `php`

The global `/usr/local/bin/composer` wrapper uses the configured default (or `php`). Project-correct PHP is `abstrax composer run --project=…` or `$ABSTRAX_CLI_PHP composer …` in deploy hooks.

## `composer run`

Put Abstrax flags before the Composer command:

```bash
abstrax composer run install
abstrax composer run install --no-dev --optimize-autoloader
abstrax composer run --project=myapp install --no-dev
abstrax composer run --php=php8.2 update
abstrax composer run --path=/srv/app --user=deploy install
```

`--verbose`, `--quiet`, and `--dry-run` are Abstrax flags. To pass those through to Composer, put them after `--`:

```bash
abstrax composer run --dry-run install          # Abstrax preview
abstrax composer run -- install --dry-run       # Composer --dry-run
```

## Config (`/etc/abstrax/composer.json`)

```json
{
  "version": 1,
  "php": "php8.2"
}
```

Omit `php` (or set `--php=php`) to use the unversioned `php` binary.

## Auth

```bash
sudo abstrax composer auth --user=deploy --github-token=ghp_…
abstrax composer auth --user=deploy
sudo abstrax composer auth --user=deploy --http-basic-host=repo.packagist.com \
  --username=token --password=secret
sudo abstrax composer auth --user=deploy --remove=github
```

Credentials are written to `~/.config/composer/auth.json` (mode `0600`). Tokens are never printed in full.
