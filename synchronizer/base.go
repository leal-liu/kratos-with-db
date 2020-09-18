package synchronizer

import (
	"time"

	"github.com/valyala/fasthttp"
)

type Base struct {
	host string
	port int
	uri  string
}

func (object *Base) SetHost(host string) *Base {
	object.host = host
	return object
}

func (object *Base) SetPort(port int) *Base {
	object.port = port
	return object
}

func (object *Base) SetUri(uri string) *Base {
	object.uri = uri
	return object
}

func (object *Base) Sync(timeout time.Duration) (code int, raw []byte, err error) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
	}()
	req.SetRequestURI(object.uri)
	if err = fasthttp.DoTimeout(req, res, timeout); nil != err {
		return
	}
	code = res.StatusCode()
	raw = make([]byte, len(res.Body()))
	copy(raw, res.Body())
	return
}
