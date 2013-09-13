package webserver

/*import (
	"bytes"
	"fmt"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/user"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func TestAddProject(t *testing.T) {
	go Run()
	baseUrl := "http://localhost:8080/"
	requests := []postHolder{
		postHolder{baseUrl + "addproject?name=project&lang=java", true},
		postHolder{baseUrl + "addproject?name=project&lang=", false},
		postHolder{baseUrl + "addproject?name=&lang=java", false},
		postHolder{baseUrl + "addproject?name=&lang=", false},
	}
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	u := user.New("user", "password")
	err := db.AddUser(u)
	if err != nil {
		t.Error(err)
	}
	err = postForm(baseUrl+"login", url.Values{"username": {"user"}, "password": {"password"}})
	if err != nil {
		t.Error(err)
	}
	for _, ph := range requests {
		//err = post(ph.url, []byte("Some data"))
		if ph.valid && err != nil {
			t.Error(err)
		} else if !ph.valid && err == nil {
			t.Error(fmt.Errorf("Expected error for %s.", ph.url))
		}
	}
}

func postForm(url string, values url.Values) error {
	return handleResponse(http.PostForm(url, values))
}

func post(url string, file []byte) error {
	buff := bytes.NewBuffer(file)
	return handleResponse(http.Post(url, "multipart/form-data", buff))
}

func handleResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	vals, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(vals))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Post failed with status %s.", resp.Status)
	}
	return nil
}
*/
