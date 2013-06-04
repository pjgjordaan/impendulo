package project

import (
	"reflect"
	"testing"
)

func TestParseName(t *testing.T) {
	correctVals := map[string]interface{}{PKG: "za.ac.sun.cs", NAME: "File.java", TIME: int64(123000012312123), NUM: 123, MOD: "c", TYPE: SRC}
	correct := "za_ac_sun_cs_File.java_123000012312123_123_c"
	incorrect := "za_ac_sun_cs_File.java_123_c"
	vals, err := ParseName(correct)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(correctVals, vals) {
		t.Error(correctVals, "!=", vals)
	}
	_, err = ParseName(incorrect)
	if err == nil {
		t.Error(incorrect, "is not a valid name")
	}

}
