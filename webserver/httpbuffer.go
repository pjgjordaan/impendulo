//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

package webserver

import (
	"bytes"
	"net/http"
	"sync"
)

type (
	//Buffer is a http.ResponseWriter which buffers all the data and headers.
	HttpBuffer struct {
		bytes.Buffer
		resp    int
		headers http.Header
		once    sync.Once
	}
)

//Header implements the Header method of http.ResponseWriter
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
