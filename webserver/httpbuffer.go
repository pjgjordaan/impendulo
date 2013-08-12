package webserver

import (
	"bytes"
	"net/http"
	"sync"
)

//Buffer is a http.ResponseWriter which buffers all the data and headers.
type HttpBuffer struct {
	bytes.Buffer
	resp    int
	headers http.Header
	once    sync.Once
}

//Header implements the header method of http.ResponseWriter
func (this *HttpBuffer) Header() http.Header {
	this.once.Do(func() {
		this.headers = make(http.Header)
	})
	return this.headers
}

//WriteHeader implements the WriteHeader method of http.ResponseWriter
func (this *HttpBuffer) WriteHeader(resp int) {
	this.resp = resp
}

//Apply takes an http.ResponseWriter and calls the required methods on it to
//output the buffered headers, response code, and data. It returns the number
//of bytes written and any errors flushing.
func (this *HttpBuffer) Apply(w http.ResponseWriter) (n int, err error) {
	if len(this.headers) > 0 {
		h := w.Header()
		for key, val := range this.headers {
			h[key] = val
		}
	}
	if this.resp > 0 {
		w.WriteHeader(this.resp)
	}
	n, err = w.Write(this.Bytes())
	return
}
