package limiter

import "github.com/aofei/air"

// BodySizeGasConfig is a set of configurations for the `BodySizeGas()`.
type BodySizeGasConfig struct {
	MaxBytes                 int64
	ErrRequestEntityTooLarge *air.Error
}

// BodySizeGas returns an `air.Gas` that is used to limit ervery request's body
// size based on the bsgc.
func BodySizeGas(bsgc BodySizeGasConfig) air.Gas {
	if bsgc.ErrRequestEntityTooLarge == nil {
		bsgc.ErrRequestEntityTooLarge = &air.Error{
			Code:    413,
			Message: "Request Entity Too Large",
		}
	}

	return func(next air.Handler) air.Handler {
		return func(req *air.Request, res *air.Response) error {
			if req.ContentLength > bsgc.MaxBytes {
				return bsgc.ErrRequestEntityTooLarge
			}

			return next(req, res)
		}
	}
}
