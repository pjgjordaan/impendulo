//Copyright (C) 2013  The Impendulo Authors
//
//This library is free software; you can redistribute it and/or
//modify it under the terms of the GNU Lesser General Public
//License as published by the Free Software Foundation; either
//version 2.1 of the License, or (at your option) any later version.
//
//This library is distributed in the hope that it will be useful,
//but WITHOUT ANY WARRANTY; without even the implied warranty of
//MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
//Lesser General Public License for more details.
//
//You should have received a copy of the GNU Lesser General Public
//License along with this library; if not, write to the Free Software
//Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

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
	x := res[len(res)-1]["x"]
	expect := 1.0
	if x != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "x", x)
	}
	expect = 1.0
	y := res[len(res)-1]["y"]
	if y != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "y", y)
	}
	AddCoords(good, -1, 5)
	AddCoords(good, -1, 50)
	res = good["data"].([]map[string]float64)
	x = res[len(res)-1]["x"]
	expect = 2.0
	if x != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "x", x)
	}
	expect = 50.0
	y = res[len(res)-1]["y"]
	if y != expect {
		t.Errorf("Expected %f for %s got %f.", expect, "y", y)
	}
}
