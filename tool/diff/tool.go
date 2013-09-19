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

//Package diff adds diff creation to Impendulo. This allows us to calculate the diff
//between two source files, convert it to HTML and display the result.
//See http://www.gnu.org/software/diffutils/manual/html_node/Invoking-diff.html#Invoking-diff
//for more information.
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

//Diff calculates and returns the diff between orig and change.
func Diff(orig, change string) (ret string, err error) {
	//Load diff executable
	exec, err := config.Binary(config.DIFF)
	if err != nil {
		return
	}
	//Store one string temporarily on disk since we can only pipe one
	//argument to diff.
	origName := filepath.Join(util.BaseDir(), fmt.Sprint(&orig)+fmt.Sprint(&change))
	err = util.SaveFile(origName, []byte(orig))
	if err != nil {
		return
	}
	defer os.Remove(origName)
	args := []string{exec, "-u", origName, "-"}
	execRes := tool.RunCommand(args, strings.NewReader(change))
	ret = string(execRes.StdOut)
	return
}

//Diff2HTML converts a diff to HTML and returns the HTML.
func Diff2HTML(diff string) (ret template.HTML, err error) {
	//If there is no diff we don't need to run the script.
	if diff == "" {
		ret = template.HTML("<h4 class=\"text-success\">Files equivalent.<h4>")
		return
	}
	//Load the script
	script, err := config.Script(config.DIFF2HTML)
	if err != nil {
		return
	}
	//Execute it and convert the result to HTML.
	execRes := tool.RunCommand([]string{script}, strings.NewReader(diff))
	if execRes.HasStdErr() {
		err = fmt.Errorf("Could not generate html: %q", string(execRes.StdErr))
	} else if execRes.Err != nil {
		err = execRes.Err
	}
	ret = template.HTML(string(execRes.StdOut))
	return
}

//SetHeader adds a header to a diff string.
func SetHeader(diff, orig, change string) string {
	i := strings.Index(diff, "@@")
	if i == -1 || i >= len(diff) {
		return ""
	}
	diff = diff[i:]
	header := "--- " + orig + "\n" + "+++ " + change + "\n"
	return header + diff
}
