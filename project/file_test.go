package project

import (
	"reflect"
	"testing"
)

func TestParseName(t *testing.T) {
	correctFile := &File{Name: "File.java", Time: int(123000012312123), Type: SRC, FileType: JAVA, Mod: "Saved", Num: 123, Package: "za.ac.sun.cs"}
	correct := "za_ac_sun_cs_File.java_123000012312123_123_c"
	incorrect := "za_ac_sun_cs_File.java_123_c"
	recvFile, err := ParseName(correct)
	if err != nil {
		t.Error(err)
	}
	correctFile.Id = recvFile.Id
	correctFile.SubId = recvFile.SubId
	if !reflect.DeepEqual(recvFile, correctFile) {
		t.Error(recvFile, "!=", correctFile)
	}
	_, err = ParseName(incorrect)
	if err == nil {
		t.Error(incorrect, "is not a valid name")
	}

}
