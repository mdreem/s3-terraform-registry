package schema

import "io"

type ProxyResponse struct {
	Body          io.ReadCloser
	ContentLength int64
	ContentType   string
}
