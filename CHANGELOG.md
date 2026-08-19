# Changelog

All notable changes to the Abstrax Composer plugin are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-08-19

### Added

- **Action IDs** - Plugin metadata commands include `action` values such as `plugin.composer.run` and `plugin.composer.self_update` for Abstrax `--action` dispatch.

### Changed

- **Help and usage** - Usage text shows `abstrax composer …` instead of the `abstrax-composer` binary name.

## [0.1.0] - 2026-08-19

First release of the official Abstrax Composer plugin (`abstrax-composer` → `abstrax composer …`).

### Added

- **`composer setup`** - Download the latest stable Composer phar from getcomposer.org, verify SHA-256, install to `/usr/local/lib/abstrax/composer/composer.phar`, and write a global wrapper at `/usr/local/bin/composer`.
- **`composer self-update`** - Repeat the verified download to update the installed phar.
- **`composer remove`** - Remove the managed wrapper and phar. `--purge` also removes `/etc/abstrax/composer.json`.
- **`composer status`** - Report install state, wrapper/phar paths, Composer version, and resolved PHP (with source).
- **`composer configure --php`** - Server-wide default PHP binary for non-project invocations; rewrites the wrapper when Composer is already installed.
- **`composer run`** - Run Composer in the current directory, or with `--project` / `--path` / `--php` / `--user`. Abstrax flags come first; colliding Composer flags go after `--`.
- **`composer diagnose`** - Check PHP, Composer, git, unzip, and common PHP extensions.
- **`composer auth`** - Manage GitHub OAuth and HTTP basic credentials in a user's `~/.config/composer/auth.json` without printing secrets.
- **PHP resolution** - `--php`, `ABSTRAX_COMPOSER_PHP`, project runtime, config default, then `php`.
- **Machine output** - `--json` and `--json-stream` matching Abstrax core.
- **Plugin metadata** - Protocol v1 `plugin metadata`; `requires_abstrax >=0.1.0`.
