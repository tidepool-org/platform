# device-check-go

![](https://github.com/rinchsan/device-check-go/workflows/CI/badge.svg)
![](https://img.shields.io/github/release/rinchsan/device-check-go.svg?colorB=7E7E7E)
[![](https://pkg.go.dev/badge/github.com/rinchsan/device-check-go.svg)](https://pkg.go.dev/github.com/rinchsan/device-check-go)
[![](https://codecov.io/github/rinchsan/device-check-go/coverage.svg?branch=master)](https://codecov.io/github/rinchsan/device-check-go?branch=master)
[![](https://goreportcard.com/badge/github.com/rinchsan/device-check-go)](https://goreportcard.com/report/github.com/rinchsan/device-check-go)
[![](https://awesome.re/mentioned-badge.svg)](https://awesome-go.com/#third-party-apis)
[![](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

:iphone: iOS DeviceCheck SDK for Go - query and modify the per-device bits

## Installation

```bash
go get github.com/rinchsan/device-check-go
```

## Getting started

### Initialize SDK

```go
import "github.com/rinchsan/device-check-go"

cred := devicecheck.NewCredentialFile("/path/to/private/key/file") // You can create credential also from raw string/bytes
cfg := devicecheck.NewConfig("ISSUER", "KEY_ID", devicecheck.Development)
client := devicecheck.New(cred, cfg)
````

### Use DeviceCheck API

#### Query two bits

```go
var result devicecheck.QueryTwoBitsResult
if err := client.QueryTwoBits("DEVICE_TOKEN", &result); err != nil {
	switch {
	// Note that QueryTwoBits returns ErrBitStateNotFound error if no bits found
	case errors.Is(err, devicecheck.ErrBitStateNotFound):
		// handle ErrBitStateNotFound error
	default:
		// handle other errors
	}
}
```

#### Update two bits

```go
if err := client.UpdateTwoBits("DEVICE_TOKEN", true, true); err != nil {
	// handle errors
}
```

#### Validate device token

```go
if err := client.ValidateDeviceToken("DEVICE_TOKEN"); err != nil {
	// handle errors
}
```

## Apple documentation

- [iOS DeviceCheck API for Swift](https://developer.apple.com/documentation/devicecheck)
- [HTTP commands to query and modify the per-device bits](https://developer.apple.com/documentation/devicecheck/accessing_and_modifying_per-device_data)
