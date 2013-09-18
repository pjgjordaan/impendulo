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

package diff

import (
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

//Diff
func Diff(orig, change string) (ret string, err error) {
	origName := filepath.Join(util.BaseDir(),
		fmt.Sprint(&orig)+fmt.Sprint(&change))
	err = util.SaveFile(origName, []byte(orig))
	if err != nil {
		return
	}
	defer os.Remove(origName)
	args := []string{config.Config(config.DIFF), "-u", origName, "-"}
	execRes := tool.RunCommand(args, strings.NewReader(change))
	ret = string(execRes.StdOut)
	return
}

//Diff2HTML
func Diff2HTML(diff string) (ret template.HTML, err error) {
	if diff == "" {
		ret = template.HTML("<h4 class=\"text-success\">Files equivalent.<h4>")
		return
	}
	args := []string{config.Config(config.DIFF2HTML)}
	execRes := tool.RunCommand(args, strings.NewReader(diff))
	if execRes.HasStdErr() {
		err = fmt.Errorf("Could not generate html: %q",
			string(execRes.StdErr))
	} else if execRes.Err != nil {
		err = execRes.Err
	}
	ret = template.HTML(string(execRes.StdOut))
	return
}

//SetHeader
func SetHeader(diff, orig, change string) string {
	i := strings.Index(diff, "@@")
	if i == -1 || i >= len(diff) {
		return ""
	}
	diff = diff[i:]
	header := "--- " + orig + "\n" + "+++ " + change + "\n"
	return header + diff
}
