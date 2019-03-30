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
// size based on the bsgc. It prevents clients from accidentally or maliciously
// sending a large request and wasting server resources.
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

			req.Body = &maxBytesBody{
				bsgc: bsgc,
				req:  req,
				res:  res,
			}

			return next(req, res)
		}
	}
}

// maxBytesReader is similar to the `io.LimitReader` but is intended for
// limiting the size of incoming request bodies.
type maxBytesBody struct {
	bsgc BodySizeGasConfig
	req  *air.Request
	res  *air.Response
	cl   int64
}

// Read implements the `io.Reader`.
func (mbb *maxBytesBody) Read(b []byte) (n int, err error) {
	if rl := mbb.bsgc.MaxBytes - mbb.cl; rl > 0 {
		if int64(len(b)) > rl {
			b = b[:rl]
		}

		n, err = mbb.req.Body.Read(b)
	} else {
		return 0, mbb.bsgc.ErrRequestEntityTooLarge
	}

	mbb.cl += int64(n)
	if err == nil && mbb.bsgc.MaxBytes-mbb.cl <= 0 {
		if mbb.res.Written {
			mbb.res.Status = http.StatusRequestEntityTooLarge
		}

		err = mbb.bsgc.ErrRequestEntityTooLarge
	}

	return
}

// Close implements the `io.Closer`.
func (mbb *maxBytesBody) Close() error {
	return nil
}
