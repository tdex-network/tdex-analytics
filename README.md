# Tdex-Analytics
Tdex-Analytics-Daemon provides historical data about balances and prices of various
Liquidity Provider's market's.<br>
It periodically polls Liquidity Provider's registry and fetches all available 
markets and their prices and balances.<br>
These background jobs fills time-series DB with data that are accessible through gRPC api's.<br>
Api specification can be found in [here](https://github.com/tdex-network/tdex-analytics/blob/master/api-spec/protobuf/tdexa/v1/analytics.proto#L1). <br>
List of available Liquidity Provider's can be found in [here](https://github.com/tdex-network/tdex-registry).

### Usage 
Build binaries:
```
make build
```
Start tdex-analytics daemon:
```
make dev
```
Config CLI:
```
./bin/tdexa config
```
List market's id's to be passed to prices/balances cmd's:
```
./bin/tdexa markets
```
List balances for last hour:
```
./bin/tdexa balances --predefined_period 1
```
List prices for last hour:
```
./bin/tdexa prices --predefined_period 1
```
In-depth documentation for installing and using the tdex-analytics is available at docs.tdex.network (TODO)

### Release

Precompiled binaries are published with each [release](https://github.com/tdex-network/tdex-analytics/releases).

### License

This project is licensed under the MIT License - see the[LICENSE](https://github.com/tdex-network/tdex-analytics/blob/master/LICENSE) file for details.