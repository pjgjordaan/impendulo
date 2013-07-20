package util

import (
	"bytes"
	"encoding/json"
	"testing"
)


func TestReadJson(t *testing.T) {
	//This only
	jmap := map[string]interface{}{"A": "2 3", "B": " Hallo ", "C": "", "D":"''"}
	marshalled, err := json.Marshal(jmap)
	if err != nil {
		t.Error(err)
	}
	reader := bytes.NewBuffer(marshalled)
	res, err := ReadJson(reader)
	if err != nil {
		t.Error(err)
	}
	for k, v := range jmap {
		if res[k] != v {
			t.Error(res[k], " != ", v)
		}
	}
}
