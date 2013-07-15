package diff

import(
	"strings"
	"github.com/godfried/impendulo/tool"
"github.com/godfried/impendulo/config"
"github.com/godfried/impendulo/util"
"os"
"path/filepath"
"html/template"
"fmt"
)

func Diff(orig, change string)(ret string, err error){	
	origName := filepath.Join(util.BaseDir(), fmt.Sprint(&orig)+fmt.Sprint(&change))
	err = util.SaveFile(origName, []byte(orig))
	if err != nil{
		return
	}
	defer os.Remove(origName)
	args := []string{config.GetConfig(config.DIFF), "-u", origName, "-"}
	stdout, stderr, _ := tool.RunInputCommand(args, strings.NewReader(change))
	fmt.Println(string(stdout), string(stderr))
	ret = string(stdout)
	return
}

func Diff2HTML(diff string) (ret template.HTML, err error){
	args := []string{config.GetConfig(config.DIFF2HTML)}
	stdout, stderr, err := tool.RunInputCommand(args, strings.NewReader(diff))
	if stderr != nil && len(stderr) > 0{
		err = fmt.Errorf("Could not generate html: %q", string(stderr))
	}
	ret = template.HTML(string(stdout))
	return
}

func SetHeader(diff, orig, change string)string{
	i := strings.Index(diff, "@@")
	if i == -1 || i >= len(diff){
		return "Files equivalent."
	}
	diff = diff[i:]
	header := "--- "+orig+"\n"+"+++ "+change+"\n"
	return header+diff
}