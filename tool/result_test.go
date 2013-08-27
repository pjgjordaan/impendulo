package tool

import (
	"testing"
)

func TestAddCoords(t *testing.T) {
	bad := CreateChart("bad")
	bad["data"] = 5
	AddCoords(bad, 0, 0)
	bad["data"] = nil
	AddCoords(bad, 0, 0)
	AddCoords(nil, 0, 0)
	good := CreateChart("good")
	AddCoords(good, 1000, 1)
	res := good["data"].([]map[string]float64)
	x := res[len(res) -1]["x"]
	expect := 1.0
	if x != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "x", x)
	}
	expect = 1.0
	y := res[len(res) -1]["y"]
	if y != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "y", y)
	}
	AddCoords(good, -1, 5)
	AddCoords(good, -1, 50)
	res = good["data"].([]map[string]float64)
	x = res[len(res) -1]["x"]
	expect = 2.0
	if x != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "x", x)
	}
	expect = 50.0
	y = res[len(res) -1]["y"]
	if y != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "y", y)
	}
}
