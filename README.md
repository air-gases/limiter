# Limiter

A useful gas that used to limit every request for the web applications built
using [Air](https://github.com/sheng/air).

## Installation

Open your terminal and execute

```bash
$ go get github.com/air-gases/limiter
```

done.

> The only requirement is the [Go](https://golang.org), at least v1.8.

## Usage

The following application will limit the body size of all requests to within 1
`MB`.

```go
package main

import (
	"github.com/air-gases/limiter"
	"github.com/sheng/air"
)

func main() {
	air.Gases = []air.Gas{
		limiter.BodySizeGas(limiter.BodySizeGasConfig{
			MaxBytes: 1 << 20,
		}),
	}
	air.GET("/", func(req *air.Request, res *air.Response) error {
		return res.String("You are within the limits.")
	})
	air.Serve()
}
```

## Community

If you want to discuss this gas, or ask questions about it, simply post
questions or ideas [here](https://github.com/air-gases/limiter/issues).

## Contributing

If you want to help build this gas, simply follow
[this](https://github.com/air-gases/limiter/wiki/Contributing) to send pull
requests [here](https://github.com/air-gases/limiter/pulls).

## License

This gas is licensed under the Unlicense.

License can be found [here](LICENSE).
