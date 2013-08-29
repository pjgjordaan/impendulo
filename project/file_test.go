package project

import (
	"reflect"
	"testing"
)

func TestParseName(t *testing.T) {
	correctFile := &File{Name: "File.java", Time: int64(1256030454696), Type: SRC, FileType: JAVA, Mod: "Saved", Num: 123, Package: "za.ac.sun.cs"}
	correct := "za_ac_sun_cs_File.java_1256030454696_123_c"
	incorrect := "za_ac_sun_cs_File.java_123_c"
	recvFile, err := ParseName(correct)
	if err != nil {
		t.Error(err)
	} else{
		correctFile.Id = recvFile.Id
		correctFile.SubId = recvFile.SubId
		if !reflect.DeepEqual(recvFile, correctFile) {
			t.Error(recvFile, "!=", correctFile)
		}
	}
	_, err = ParseName(incorrect)
	if err == nil {
		t.Error(incorrect, "is not a valid name")
	}

}
