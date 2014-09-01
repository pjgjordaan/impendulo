package webutil

import (
	"fmt"

	"labix.org/v2/mgo/bson"

	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/util/convert"

	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

//File reads a file's name and data from a request form.
func File(r *http.Request, n string) (string, []byte, error) {
	f, h, e := r.FormFile(n)
	if e != nil {
		return "", nil, e
	}
	d, e := ioutil.ReadAll(f)
	if e != nil {
		return "", nil, e
	}
	return h.Filename, d, nil
}

//Strings retrieves a string value from a request form.
func Strings(r *http.Request, n string) ([]string, error) {
	if r.Form == nil {
		if e := r.ParseForm(); e != nil {
			return nil, e
		}
	}
	return r.Form[n], nil
}

func Bool(r *http.Request, n string) (bool, error) {
	v, e := String(r, n)
	if e != nil {
		return false, e
	}
	return convert.Bool(v)
}

//String retrieves a string value from a request form.
func String(r *http.Request, n string) (string, error) {
	v := r.FormValue(n)
	if strings.TrimSpace(v) != "" {
		return v, nil
	}
	if r.Form == nil {
		if e := r.ParseForm(); e != nil {
			return "", e
		}
	}
	vs, ok := r.Form[n]
	if !ok || len(vs) == 0 || vs[0] == "" {
		return "", fmt.Errorf("invalid value for %s", n)
	}
	return vs[0], nil
}

func Int(r *http.Request, n string) (int, error) {
	v, e := String(r, n)
	if e != nil {
		return -1, e
	}
	return convert.Int(v)
}

func Int64(r *http.Request, n string) (int64, error) {
	v, e := String(r, n)
	if e != nil {
		return -1, e
	}
	return convert.Int64(v)
}

func Id(r *http.Request, n string) (bson.ObjectId, error) {
	v, e := String(r, n)
	if e != nil {
		return "", e
	}
	return convert.Id(v)
}

func Index(r *http.Request, n string, maxSize int) (int, error) {
	i, e := Int(r, n)
	if e != nil {
		return -1, e
	}
	if i > maxSize {
		return 0, nil
	} else if i < 0 {
		return maxSize, nil
	}
	return i, nil
}

func ServePath(u *url.URL, src string) (string, error) {
	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}
	ext, e := filepath.Rel("/"+filepath.Base(src), path.Clean(u.Path))
	if e != nil {
		return "", e
	}
	sp := filepath.Join(src, ext)
	if util.IsDir(sp) && !strings.HasSuffix(u.Path, "/") {
		u.Path = u.Path + "/"
	}
	return sp, nil
}

func Credentials(r *http.Request) (string, string, error) {
	u, e := String(r, "user-id")
	if e != nil {
		return "", "", e
	}
	p, e := String(r, "password")
	if e != nil {
		return "", "", e
	}
	return u, p, nil
}
