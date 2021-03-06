package limiter

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aofei/air"
	"github.com/patrickmn/go-cache"
)

// BodySizeGasConfig is a set of configurations for the `BodySizeGas`.
type BodySizeGasConfig struct {
	MaxBytes                 int64
	ErrRequestEntityTooLarge error

	Skippable func(*air.Request, *air.Response) bool
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
			if bsgc.Skippable != nil && bsgc.Skippable(req, res) {
				return next(req, res)
			}

			if req.ContentLength > bsgc.MaxBytes {
				res.Status = http.StatusRequestEntityTooLarge
				return bsgc.ErrRequestEntityTooLarge
			}

			req.Body = &maxBytesBody{
				bsgc:    bsgc,
				reqBody: req.Body,
				res:     res,
			}

			return next(req, res)
		}
	}
}

// maxBytesReader is similar to the `io.LimitReader` but is intended for
// limiting the size of incoming request bodies.
type maxBytesBody struct {
	sync.Mutex

	bsgc    BodySizeGasConfig
	reqBody io.ReadCloser
	res     *air.Response
	cl      int64
}

// Read implements the `io.Reader`.
func (mbb *maxBytesBody) Read(b []byte) (n int, err error) {
	mbb.Lock()
	defer mbb.Unlock()

	if rl := mbb.bsgc.MaxBytes - mbb.cl; rl > 0 {
		if int64(len(b)) > rl {
			b = b[:rl]
		}

		n, err = mbb.reqBody.Read(b)
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
	return mbb.reqBody.Close()
}

// RateGasConfig is a set of configurations for the `RateGas`.
type RateGasConfig struct {
	MaxRequests        int64
	ResetInterval      time.Duration
	UseClientAddress   bool
	ErrTooManyRequests error

	Skippable func(*air.Request, *air.Response) bool

	counterCache *cache.Cache
}

// RateGas returns an `air.Gas` that is used to limit request's rate based on
// the rgc.
func RateGas(rgc RateGasConfig) air.Gas {
	if rgc.ErrTooManyRequests == nil {
		rgc.ErrTooManyRequests = errors.New(
			http.StatusText(http.StatusTooManyRequests),
		)
	}

	rgc.counterCache = cache.New(rgc.ResetInterval, time.Minute)

	return func(next air.Handler) air.Handler {
		return func(req *air.Request, res *air.Response) error {
			if rgc.Skippable != nil && rgc.Skippable(req, res) {
				return next(req, res)
			}

			if rgc.MaxRequests <= 0 || rgc.ResetInterval <= 0 {
				return next(req, res)
			}

			host := ""
			if rgc.UseClientAddress {
				host = req.ClientHost()
			} else {
				host = req.RemoteHost()
			}

			_, e, ok := rgc.counterCache.GetWithExpiration(host)
			if !ok || e.Before(time.Now()) {
				rgc.counterCache.SetDefault(host, int64(0))
				e = time.Now().Add(rgc.ResetInterval)
			}

			count, err := rgc.counterCache.IncrementInt64(host, 1)
			if err != nil {
				return err
			}

			remaining := rgc.MaxRequests - count
			reached := remaining < 0
			if reached {
				remaining = 0
			}

			res.Header.Set(
				"X-RateLimit-Limit",
				strconv.FormatInt(rgc.MaxRequests, 10),
			)
			res.Header.Set(
				"X-RateLimit-Remaining",
				strconv.FormatInt(remaining, 10),
			)
			res.Header.Set(
				"X-RateLimit-Reset",
				strconv.FormatInt(e.Unix(), 10),
			)

			if reached {
				res.Status = http.StatusTooManyRequests
				return rgc.ErrTooManyRequests
			}

			return next(req, res)
		}
	}
}
