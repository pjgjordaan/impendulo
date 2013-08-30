package webserver
/*
import (
	"code.google.com/p/gorilla/sessions"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/processing"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/user"
	"github.com/godfried/impendulo/util"
	"mime/multipart"
	"net/http"
	//"net/url"
	"testing"
	"strconv"
	"os"
	"bytes"
	"path/filepath"
	"io"
	"fmt"
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
	u := user.New("username", "password")
	err = db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	zf, err := util.ZipMap(makeZipMap())
	if err != nil {
		t.Error(err)
	}
	tempFile := "/tmp/processing_test/temp.zip"
	err = util.SaveFile(tempFile, zf)
	if err != nil {
		t.Error(err)
	}
	//defer os.RemoveAll(tempFile)
	uri := fmt.Sprintf("/submitarchive?project=%s&user=%s",p.Id.Hex(), u.Name)
	req, err := archiveRequest(uri, tempFile)
	if err != nil {
		t.Error(err)
	}	
	store := sessions.NewCookieStore(util.CookieKeys())
	sess, err := store.Get(req, "test")
	if err != nil {
		t.Error(err)
	}	
	ctx := NewContext(sess)
	ctx.AddUser(u.Name)
	err = SubmitArchive(req, ctx)
	if err != nil {
		t.Error(err)
	}
	processing.Shutdown()
}

func archiveRequest(uri, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("archive", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}
	fmt.Println(body.String())
	return http.NewRequest("POST", uri, body)
}

func makeZipMap()(zipMap map[string][]byte){
	name := "_za.ac.sun.ac.za.Triangle_src_triangle_Triangle.java_"
	time := 1256033823717
	num := 8583
	zipMap = make(map[string][]byte)
	for i := 0; i < 10; i++{
		t := strconv.Itoa(time+i*100)
		n := strconv.Itoa(num+i)
		zipMap[name+t+"_"+n+"_c"] = fileData
	}
	return
}

var fileData = []byte(`
package triangle;
public class Triangle {
	public int maxpath(int[][] triangle) {
		int height = triangle.length - 2;
		for (int i = height; i >= 1; i--) {
			for (int j = 0; j <= i; j++) {
				triangle[i][j] += triangle[i + 1][j + 1] > triangle[i + 1][j] ? triangle[i + 1][j + 1]
						: triangle[i + 1][j];
			}
		}
		return triangle[0][0];
	}
}
`)
*/