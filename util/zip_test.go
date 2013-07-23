package util

import(
	"testing"
"errors"
"bytes"
)
func TestZip(t *testing.T) {
	files := map[string][]byte{"readme.txt": []byte("This archive contains some text files."), "gopher.txt": []byte("Gopher names:\nGeorge\nGeoffrey\nGonzo"), "todo.txt": []byte("Get animal handling licence.\nWrite more examples.")}
	zipped, err := Zip(files)
	if err != nil {
		t.Error(err)
	}
	unzipped, err := UnzipToMap(zipped)
	if err != nil {
		t.Error(err)
	}
	if len(files) != len(unzipped) {
		t.Error(errors.New("Zip error; invalid size"))
	}
	for k, v := range files {
		if !bytes.Equal(v, unzipped[k]) {
			t.Error(errors.New("Zip error, values not equal."))
		}
	}

}
