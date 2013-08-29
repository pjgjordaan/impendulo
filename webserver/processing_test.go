package webserver
/*
import (
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/util"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"
)

func TestSubmitArchive(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	
	go processing.Serve(10)
	p := project.NewProject("Test", "user", "java", []byte{})
	err := db.AddProject(p)
	if err != nil {
		t.Error(err)
	}
	vals := url.Values{"project": []string{p.Id.Hex()}}
	fHeaders := []*multipart.FileHeader{&multipart.FileHeader{Filename: "/home/godfried/dev/go/src/github.com/godfried/impendulo/webserver/testArchive.zip"}}
	multiVals := &multipart.Form{Value: map[string][]string{"archive": []string{"archive"}}, File: map[string][]*multipart.FileHeader{"archive": fHeaders}}
	req := &http.Request{Method: "POST", Form: vals, MultipartForm: multiVals}
	store := sessions.NewCookieStore(util.CookieKeys())
	sess, err := store.Get(req, "test")
	if err != nil {
		t.Error(err)
	}
	ctx := NewContext(sess)
	err = SubmitArchive(req, ctx)
	if err != nil {
		t.Error(err)
	}
	processing.Shutdown()
}
*/