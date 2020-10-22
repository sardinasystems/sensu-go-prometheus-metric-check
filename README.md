[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/sardinasystems/sensu-go-prometheus-metric-check)
![Go Test](https://github.com/sardinasystems/sensu-go-prometheus-metric-check/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/sardinasystems/sensu-go-prometheus-metric-check/workflows/goreleaser/badge.svg)

# sensu-go-prometheus-metric-check

## Table of Contents
- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Check definition](#check-definition)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

The sensu-go-prometheus-metric-check is a [Sensu Check][6] that queries Prometheus for alerting.

NOTE: for critical and warning levels [Nagios threshold format](http://nagios-plugins.org/doc/guidelines.html#THRESHOLDFORMAT) used.

## Files

- sensu-go-prometheus-metric-check

## Usage examples

```
sensu-go-prometheus-metric-check -H http://example.com:9090 -q scalar(up{instance="example.com:9100"}) -c 1: -w 1:	# scalar
sensu-go-prometheus-metric-check -H http://example.com:9090 -q up{instance="example.com:9100"} -c 1: -w 1:		# vector
```

## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add sardinasystems/sensu-go-prometheus-metric-check
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/sardinasystems/sensu-go-prometheus-metric-check].

### Check definition

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: sensu-go-prometheus-metric-check
  namespace: default
spec:
  command: sensu-go-prometheus-metric-check -H http://example.com:9090 -q rate(node_network_receive_bytes_total[1m]) -c 1e7: -w 1e6:
  subscriptions:
  - system
  runtime_assets:
  - sardinasystems/sensu-go-prometheus-metric-check
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-go-prometheus-metric-check repository:

```
go build
```

## Additional notes

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu-community/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/check-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/check-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/checks/
[7]: https://github.com/sensu-community/check-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-community/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
