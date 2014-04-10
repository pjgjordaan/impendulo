//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

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
func Diff(orig, change string) (string, error) {
	//Load diff executable
	d, e := config.DIFF.Path()
	if e != nil {
		return "", e
	}
	b, e := util.BaseDir()
	if e != nil {
		return "", e
	}
	//Store one string temporarily on disk since we can only pipe one
	//argument to diff.
	p := filepath.Join(b, fmt.Sprint(&orig)+fmt.Sprint(&change))
	if e = util.SaveFile(p, []byte(orig)); e != nil {
		return "", e
	}
	defer os.Remove(p)
	a := []string{d, "-u", p, "-"}
	r, e := tool.RunCommand(a, strings.NewReader(change))
	if e != nil {
		return "", e
	}
	return string(r.StdOut), nil
}

//Diff2HTML converts a diff to HTML and returns the HTML.
func Diff2HTML(d string) (template.HTML, error) {
	//If there is no diff we don't need to run the script.
	if d == "" {
		return template.HTML("<h4 class=\"text-success\">Files equivalent.<h4>"), nil
	}
	//Load the script
	s, e := config.DIFF2HTML.Path()
	if e != nil {
		return "", e
	}
	//Execute it and convert the result to HTML.
	r, e := tool.RunCommand([]string{s}, strings.NewReader(d))
	if r.HasStdErr() {
		return "", fmt.Errorf("Could not generate html: %q", string(r.StdErr))
	} else if e != nil {
		return "", e
	}
	return template.HTML(string(r.StdOut)), nil
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
