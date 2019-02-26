package limiter

import (
	"errors"
	"net/http"

	"github.com/aofei/air"
)

// BodySizeGasConfig is a set of configurations for the `BodySizeGas`.
type BodySizeGasConfig struct {
	MaxBytes                 int64
	ErrRequestEntityTooLarge error
}

// BodySizeGas returns an `air.Gas` that is used to limit ervery request's body
// size based on the bsgc.
func BodySizeGas(bsgc BodySizeGasConfig) air.Gas {
	if bsgc.ErrRequestEntityTooLarge == nil {
		bsgc.ErrRequestEntityTooLarge = errors.New(
			http.StatusText(http.StatusRequestEntityTooLarge),
		)
	}

	return func(next air.Handler) air.Handler {
		return func(req *air.Request, res *air.Response) error {
			if req.ContentLength > bsgc.MaxBytes {
				res.Status = http.StatusRequestEntityTooLarge
				return bsgc.ErrRequestEntityTooLarge
			}

			return next(req, res)
		}
	}
}
