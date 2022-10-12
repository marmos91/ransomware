# ransomware
A simple demonstration tool to simulate a ransomware attack

## Status
This project is currently being developed. Use at your own risk

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


### Encrypt a directory

Use an HEX key
```bash
$ ransomware encrypt --path ~/Downloads --key 6E6462656175746966756C70617373776F7264 --ext-blacklist txt,png,exe
```

Use a password
```bash
$ ransomware encrypt --path ~/Downloads --pass my-password --ext-blacklist txt,png,exe
```

### Decrypt a directory
Use an HEX key
```bash
$ ransomware decrypt --path ~/Downloads --key 6E6462656175746966756C70617373776F7264
```

Use a password
```bash
$ ransomware decrypt --path ~/Downloads --pass my-password
```
## How it works

COMING SOON