# eroero

[![Latest Release](https://img.shields.io/github/release/fawni/eroero.svg)](https://github.com/fawni/eroero/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/fawni/eroero/build.yml?logo=github&branch=master)](https://github.com/fawni/eroero/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/fawni/eroero)](https://goreportcard.com/report/github.com/fawni/eroero)

A tiny downloader for erome. 💄

## Installation

### Binaries

Download a binary from the [releases](https://github.com/fawni/eroero/releases)
page.

### Build from source

Go 1.17 or higher required. ([install instructions](https://golang.org/doc/install.html))

    go install github.com/fawni/eroero@latest

## Usage

```
eroero <album id>
```

`eroero -h` for more information.

### Flags

- `-o`, `--output`: output files to a specific directory _(default: current directory)_

## License

[ISC](LICENSE)
