package tool

import (
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
)

var fb = &Tool{bson.NewObjectId(), "findbugs", JAVA, "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "warning_count", WARNS, []string{JAVA, "-jar"}, []string{"-textui", "-low"}, bson.M{}, PKG_PATH}
var javac = &Tool{bson.NewObjectId(), COMPILE, JAVA, JAVAC, WARNS, ERRS, []string{}, []string{"-implicit:class"}, bson.M{CP: ""}, FILE_PATH}

func TestGetArgs(t *testing.T) {
	fbExp := []string{"java", "-jar", "/home/disco/apps/findbugs-2.0.2/lib/findbugs.jar", "-textui", "-low", "here"}
	res := fb.GetArgs("here")
	if !reflect.DeepEqual(fbExp, res) {
		t.Error("Arguments not computed correctly", res)
	}
	compExp := []string{JAVAC, "-implicit:class", CP, "there", "here"}
	res = javac.GetArgs("here")
	if reflect.DeepEqual(compExp, res) {
		t.Error("Arguments not computed correctly", res)
	}
	javac.setFlagArgs(map[string]string{CP: "there"})
	res = javac.GetArgs("here")
	if !reflect.DeepEqual(compExp, res) {
		t.Error("Arguments not computed correctly", res)
	}

}
