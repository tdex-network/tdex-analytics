# 📊 tdex-analytics

[![Go Report Card](https://goreportcard.com/badge/github.com/tdex-network/tdex-analytics)](https://goreportcard.com/report/github.com/tdex-network/tdex-analytics)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/tdex-network/tdex-analytics)](https://pkg.go.dev/github.com/tdex-network/tdex-analytics)
[![Release](https://img.shields.io/github/release/tdex-network/tdex-analytics.svg)](https://github.com/tdex-network/tdex-analytics/releases/latest)

TDEX Analytics provides historical data about balances and prices of Liquidity Providers registered in the [tdex-registry](https://github.com/tdex-network/tdex-registry).

It periodically polls Liquidity Provider's registry and fetches all available markets and their market price and balances and fills a time-series DB with data that are accessible through gRPC & JSON HTTP APIs.

API specifications can be found in [here](https://github.com/tdex-network/tdex-analytics/blob/master/api-spec/protobuf/tdexa/v1).


## 🖥 Local Development

Below is a list of commands you will probably find useful for development.

### Requirements

* Go (^1.17.*)


### Build

Builds `tdexad` as static binary in the `./bin` folder

```bash
$ make build
```

### Build CLI

Builds `tdexa` as static binary in the `./bin` folder

```bash
$ make build-cli
```

### Run with docker-compose

Build docker image and runs the project with other dependencies with default configuration.

```bash
$ make dev
```

### Test

```bash
$ make testall
```

## 📄 Usage

- Configure CLI:
```
./bin/tdexa config
```

- List market's id's to be passed to prices/balances cmd's:
```
./bin/tdexa markets
```

- List balances for last hour:
```
./bin/tdexa balances --predefined_period 1
```

- List prices for last hour:
```
./bin/tdexa prices --predefined_period 1
```

### Release

Precompiled binaries are published with each [release](https://github.com/tdex-network/tdex-analytics/releases).

### License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/tdex-network/tdex-analytics/blob/master/LICENSE) file for details.
