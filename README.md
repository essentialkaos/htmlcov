<p align="center"><a href="#readme"><img src=".github/images/card.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/w/htmlcov/ci"><img src="https://kaos.sh/w/htmlcov/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/r/htmlcov"><img src="https://kaos.sh/r/htmlcov.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/w/htmlcov/codeql"><img src="https://kaos.sh/w/htmlcov/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src=".github/images/license.svg"/></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#command-line-completion">Command-line completion</a> • <a href="#man-documentation">Man documentation</a> • <a href="#usage">Usage</a> • <a href="#ci-status">CI Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`htmlcov` is an utility for converting Go coverage profiles into HTML pages. It's just better version of `go tool cover -html=cover.out -o coverage.html` command.

![Screenshot](.github/images/screenshot1.png)

![Screenshot](.github/images/screenshot2.png)

### Installation

#### From source

To build the `htmlcov` from scratch, make sure you have a working Go 1.18+ workspace (_[instructions](https://go.dev/doc/install)_), then:

```
go install github.com/essentialkaos/htmlcov@latest
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux and macOS from [EK Apps Repository](https://apps.kaos.st/htmlcov/latest):

```bash
bash <(curl -fsSL https://apps.kaos.st/get) htmlcov
```

### Command-line completion

You can generate completion for `bash`, `zsh` or `fish` shell.

Bash:
```bash
sudo htmlcov --completion=bash 1> /etc/bash_completion.d/htmlcov
```

ZSH:
```bash
sudo htmlcov --completion=zsh 1> /usr/share/zsh/site-functions/htmlcov
```

Fish:
```bash
sudo htmlcov --completion=fish 1> /usr/share/fish/vendor_completions.d/htmlcov.fish
```

### Man documentation

You can generate man page using next command:

```bash
htmlcov --generate-man | sudo gzip > /usr/share/man/man1/htmlcov.1.gz
```

### Usage

<p align="center"><img src=".github/images/usage.svg"/></p>

### CI Status

| Branch | Status |
|--------|----------|
| `master` | [![CI](https://kaos.sh/w/htmlcov/ci.svg?branch=master)](https://kaos.sh/w/htmlcov/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/htmlcov/ci.svg?branch=develop)](https://kaos.sh/w/htmlcov/ci?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
