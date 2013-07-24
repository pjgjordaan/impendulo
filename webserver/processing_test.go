package webserver

import(
	"github.com/godfried/impendulo/util"
	"code.google.com/p/gorilla/sessions"
	"mime/multipart"
	"net/http"
	"net/url"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"testing"

)

func TestDoArchive(t *testing.T){
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	go processing.Serve()
	p := project.NewProject("Test", "user", "java")
	err := db.AddProject(p)
	if err != nil{
		t.Error(err)
	}
	println("a")
	vals := url.Values{"project":[]string{p.Id.Hex()}}
	fHeaders := []*multipart.FileHeader{&multipart.FileHeader{Filename: "/home/godfried/dev/go/src/github.com/godfried/impendulo/webserver/testArchive.zip"}}
	multiVals := &multipart.Form{Value: map[string][]string{"archive":[]string{"archive"}},File: map[string][]*multipart.FileHeader{"archive":fHeaders}}
	req := &http.Request{Method: "POST", Form: vals, MultipartForm: multiVals}
	println("a")
	store := sessions.NewCookieStore(util.CookieKeys())
	println("b")
	sess, err := store.Get(req, "test")
	if err != nil{
		t.Error(err)
	}
	println("c")
	ctx := NewContext(sess)
	msg,err := doArchive(req, ctx)
	if err != nil{
		t.Error(err, msg)
	}
}