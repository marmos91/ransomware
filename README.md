# ransomware
A simple demonstration tool to simulate a ransomware attack

## ⚠️ Disclaimer ⚠️
This software is made just for demonstration purposes. If you want to run it locally for tests, take care of what directories you decide to encrypt. The software is distributed in MIT license. Its use is free, however the author doesn't take responsibility for any illegal use of the code by 3rd parties.

## Setup

To setup the tool just run 

```bash
$ go install github.com/marmos91/ransomware
```

## Why 
In order to demonstrate the way ransomware works quickly and in a protected environment, **it is very useful to be able to restrict its operation within a directory**. This way the process takes much less time (the entire operating system does not need to be encrypted). Writing this tool in Go, also **allows the tool to be developed even in a non-Windows environment** (by far the most supported operating system by ransomware available online)

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

COMMANDS:
   create-keys, c  Generates a new random keypair and saves it to a file
   encrypt, e      Encrypts a directory
   decrypt, d      Decrypts a directory
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --silent       Runs the tool in silent mode (no logs) (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Create a keypair

First thing you need to do is to create a keypair. You can do it by running

```bash
$ ransomware create-keys --path ~/Desktop
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
   --path value, -p value  Runs the tool on a directory
   --publicKey value       Loads the provided RSA public key in PEM format
   --extBlacklist value    the extension to blacklist (default: ".enc")
   --extWhitelist value    the extension to whitelist
   --skipHidden            skips hidden folders (default: false)
   --dryRun                encrypts files without deleting originals (default: false)
   --encSuffix value       defines the suffix to add to encrypted files (default: ".enc")
   --help, -h              show help (default: false)
```

For example if you want to run the tool on the `~/Documents` folder run:
```bash
$ ransomware encrypt --publicKey ./pub.pem --path ~/Documents
```
This command provides the following options:

- `path`: the path to encrypt. This is required
- `publicKey`: the path of the publicKey PEM file created by the `create-keys` command
- `extBlacklist`: if provided, a comma-separated list of extension to skip. **This feature is useful, to exclude executable like `.exe` files**
- `extWhitelist`: if provided, a comma-separated list of extension to whitelist
- `skipHidden`: if set, skips hidden folders
- `dryRun`: just creates encrypted files without deleting originals
- `encSuffix`: defines a custom extension to set on encrypted files (default `.enc`)

### Examples
Just encrypt gif files on Desktop
```bash
$ ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --dryRun --extWhitelist .gif
```

Encrypt everything except `.csv` and `.pdf` files
```bash
$ ransomware encrypt --publicKey ./pub.pem --path ~/Desktop --dryRun --extBlacklist .csv,.pdf
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
   --help, -h              show help (default: false)
```
For example if you want to run the tool on the `~/Documents` folder run:

```bash
$ ransomware decrypt --privateKey ./priv.pem --path ~/Desktop/toEncrypt
```

This command provides the following options:

- `path`: the path to encrypt. This is required
- `privateKey`: the path of the privateKey PEM file created by the `create-keys` command
- `dryRun`: just creates decrypted files without deleting encrypted version
- `encSuffix`: defines a custom extension for encrypted files (default `.enc`)


## How it works

Coming soon...