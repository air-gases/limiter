package limiter

import (
	"errors"

	"github.com/aofei/air"
)

// BodySizeGasConfig is a set of configurations for the `BodySizeGas()`.
type BodySizeGasConfig struct {
	MaxBytes int64
	Error413 error
}

// BodySizeGas returns an `air.Gas` that is used to limit ervery request's body
// size based on the bsgc.
func BodySizeGas(bsgc BodySizeGasConfig) air.Gas {
	if bsgc.Error413 == nil {
		bsgc.Error413 = errors.New("Request Entity Too Large")
	}

	return func(next air.Handler) air.Handler {
		return func(req *air.Request, res *air.Response) error {
			if req.ContentLength > bsgc.MaxBytes {
				res.Status = 413
				return bsgc.Error413
			}

			return next(req, res)
		}
	}
}
