package wc

import (
	"errors"
	"fmt"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util/convert"

	"strings"

	"time"
)

var NoCountsError = errors.New("no counts for wc")

func Lines(data string) (int64, error) {
	r, e := tool.RunCommand([]string{"wc", "-l"}, strings.NewReader(data), 10*time.Minute)
	if e != nil {
		return -1, e
	}
	if r.StdErr != nil && len(r.StdErr) > 0 {
		return -1, fmt.Errorf("error %s running wc", string(r.StdErr))
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
