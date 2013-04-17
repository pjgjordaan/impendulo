package utils

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestReadJSON(t *testing.T) {
	//This only
	jmap := map[string]interface{}{"A": "2 3", "B": " Hallo ", "C": "", "D": "''"}
	marshalled, err := json.Marshal(jmap)
	if err != nil {
		t.Error(err)
	}
	reader := bytes.NewBuffer(marshalled)
	res, err := ReadJSON(reader)
	if err != nil {
		t.Error(err)
	}
	for k, v := range jmap {
		if res[k] != v {
			t.Error(res[k], " != ", v)
		}
	}
}

func TestJSONValue(t *testing.T) {
	jmap := map[string]interface{}{"A": "2 3", "B": " Hallo ", "C": "", "D": "''"}
	for k, v := range jmap {
		if res, err := JSONValue(jmap, k); err != nil || res != v {
			t.Error(err, res, "!=", v)
		}
	}

}
