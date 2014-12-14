package lc

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util/convert"

	"strings"

	"time"
)

var NoCountsError = errors.New("no counts for wc")
var javaComments = []string{`\/\/`, `\/\*\*`, `\/\*`, `\*`}

func Lines(data string) (int64, error) {
	return lines(data, javaComments)
}

func lines(data string, comments []string) (int64, error) {
	cs := ""
	for _, c := range comments {
		cs += `/^\s*` + c + `/d;`
	}
	es := `/^\s*$/d`
	r, e := tool.RunCommand([]string{`/bin/sed`, cs + es}, strings.NewReader(data), 10*time.Minute)
	if e != nil {
		return -1, e
	}
	if r.StdErr != nil && len(r.StdErr) > 0 {
		return -1, fmt.Errorf("error %s running sed", string(r.StdErr))
	}
	if r, e = tool.RunCommand([]string{"wc", "-l"}, bytes.NewReader(r.StdOut), 10*time.Minute); e != nil {
		return -1, e
	}
	sp := strings.Split(string(r.StdOut), " ")
	if len(sp) == 0 {
		return -1, NoCountsError
	}
	return convert.Int64(strings.TrimSpace(sp[0]))
}

func LinesB(data []byte) (int64, error) {
	return Lines(string(data))
}
