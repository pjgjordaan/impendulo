package web

import (
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/context"
	"github.com/godfried/impendulo/httpbuf"
	"github.com/godfried/impendulo/util"
	"net/http"
)

var store sessions.Store

func init() {
	store = sessions.NewCookieStore(util.CookieKeys())
}

type handler func(http.ResponseWriter, *http.Request, *context.Context) error

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sess, err := store.Get(req, "impendulo")
	if err != nil {
		util.Log(err)
	}
	ctx := context.NewContext(sess)
	buf := new(httpbuf.Buffer)
	err = h(buf, req, ctx)
	if err != nil {
		util.Log(err)
	}
	if err = ctx.Save(req, buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf.Apply(w)
}
