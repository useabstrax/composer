# Abstrax Composer Plugin

Install and run Composer with the correct PHP binary on an Abstrax server.

Binary: `abstrax-composer` → `abstrax composer …`  
Trust level: `official`  
OS: Debian/Ubuntu and RHEL-family.

Full user docs: [useabstrax.com/docs/plugins/official/composer](https://useabstrax.com/docs/plugins/official/composer)

## What it does

Abstrax installs PHP for projects; this plugin installs Composer itself.

- Downloads the stable phar from getcomposer.org and verifies the SHA-256 checksum
- Installs `/usr/local/lib/abstrax/composer/composer.phar` and a wrapper at `/usr/local/bin/composer`
- Optional default PHP when you are not in an Abstrax project
- `composer run --project=…` supplies path, user, and the project's PHP version

Pass Composer arguments through `abstrax composer run`. The plugin does not wrap every Composer subcommand.

## Install

```bash
sudo abstrax plugin install composer
sudo abstrax composer setup
abstrax composer version
```

Local build:

```bash
go build -o bin/abstrax-composer ./cmd/abstrax-composer
sudo cp bin/abstrax-composer /usr/local/lib/abstrax/plugins/
```

Tagged releases (`v*`) publish linux-amd64 / linux-arm64 archives and `plugin-manifest.json`.

## Quick start

```bash
sudo abstrax composer setup
abstrax composer status

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

Put Abstrax flags before the Composer command. `--verbose`, `--quiet`, and `--dry-run` are Abstrax flags; to pass them through, use `--`:

```bash
abstrax composer run --dry-run install          # Abstrax preview
abstrax composer run -- install --dry-run       # Composer --dry-run
```

## PHP resolution

1. `--php` on this invocation
2. `ABSTRAX_COMPOSER_PHP`
3. `--project` when the project runtime is PHP
4. Default in `/etc/abstrax/composer.json`
5. `php`

The global wrapper uses the configured default (or `php`). Project-correct PHP is `abstrax composer run --project=…`.

## Development

```bash
go test -race ./...
go vet ./...
go build -o bin/abstrax-composer ./cmd/abstrax-composer
```
