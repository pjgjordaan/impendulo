package web

import (
	"net/http"
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/httpbuf"
	"github.com/godfried/impendulo/context"
	"os"
)

var store sessions.Store

func init(){
	store = sessions.NewCookieStore([]byte(os.Getenv("KEY")))
}

type handler func(http.ResponseWriter, *http.Request, *context.Context) error

func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//create the context
	sess, err := store.Get(req, "impendulo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := context.NewContext(req, sess)
	//run the handler and grab the error, and report it
	buf := new(httpbuf.Buffer)
	err = h(buf, req, ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//save the session
	if err = ctx.Session.Save(req, buf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//apply the buffered response to the writer
	buf.Apply(w)
}

