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

func CalcDiff(currHeader, nextHeader, currCode, nextCode string) (ret template.HTML){
	diff, err := Diff(currCode, nextCode)
	if err != nil{
		util.Log(err)
		return
	}
	diff = SetHeader(diff, currHeader, nextHeader)
	ret, err = Diff2HTML(diff)
	if err != nil{
		util.Log(err)
	}
	return
}	

func Diff(orig, change string) (ret string, err error) {
	origName := filepath.Join(util.BaseDir(),
		fmt.Sprint(&orig)+fmt.Sprint(&change))
	err = util.SaveFile(origName, []byte(orig))
	if err != nil {
		return
	}
	defer os.Remove(origName)
	args := []string{config.GetConfig(config.DIFF), "-u", origName, "-"}
	execRes := tool.RunCommand(args, strings.NewReader(change))
	ret = string(execRes.StdOut)
	return
}

func Diff2HTML(diff string) (ret template.HTML, err error) {
	args := []string{config.GetConfig(config.DIFF2HTML)}
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
		return "Files equivalent."
	}
	diff = diff[i:]
	header := "--- " + orig + "\n" + "+++ " + change + "\n"
	return header + diff
}
