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

func SetHeader(diff, orig, change string) string {
	i := strings.Index(diff, "@@")
	if i == -1 || i >= len(diff) {
		return ""
	}
	diff = diff[i:]
	header := "--- " + orig + "\n" + "+++ " + change + "\n"
	return header + diff
}
