package pmd

import (
	"encoding/json"
	"fmt"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"strings"
)

type Tool struct {
	cmd   string
	rules string
}

func New(rules []string) *Tool {
	return &Tool{
		cmd:   config.Config(config.PMD),
		rules: strings.Join(rules, ","),
	}
}

func (this *Tool) Lang() string {
	return tool.JAVA
}

func (this *Tool) Name() string {
	return NAME
}

func (this *Tool) Run(fileId bson.ObjectId, ti *tool.TargetInfo) (res tool.ToolResult, err error) {
	outFile := filepath.Join(ti.Dir, "pmd.xml")
	args := []string{this.cmd, config.PMD, "-f", "xml", "-stress",
		"-shortnames", "-R", this.rules, "-r", outFile, "-d", ti.Dir}
	defer os.Remove(outFile)
	execRes := tool.RunCommand(args, nil)
	resFile, err := os.Open(outFile)
	if err == nil {
		//Tests ran successfully.
		data := util.ReadBytes(resFile)
		res, err = NewResult(fileId, data)
	} else if execRes.HasStdErr() {
		err = fmt.Errorf("Could not run pmd: %q.", string(execRes.StdErr))
	} else {
		err = execRes.Err
	}
	return
}

type Rule struct {
	Id          string
	Name        string
	Description string
	Default     bool
}

var ruleSet []Rule

type Rules struct {
	Id        bson.ObjectId "_id"
	ProjectId bson.ObjectId "projectid"
	Rules     []string      "rules"
}

func NewRules(projectId bson.ObjectId, rules []string) *Rules {
	return &Rules{
		Id:        bson.NewObjectId(),
		ProjectId: projectId,
		Rules:     rules,
	}
}

func RuleSet() ([]Rule, error) {
	if ruleSet != nil {
		return ruleSet, nil
	}
	cfg, err := os.Open(config.Config(config.PMD_RULES))
	if err == nil {
		data := util.ReadBytes(cfg)
		err = json.Unmarshal(data, &ruleSet)
	}
	return ruleSet, err
}

func DefaultRules(projectId bson.ObjectId) *Rules {
	set, err := RuleSet()
	rules := NewRules(projectId, make([]string, 0, len(set)))
	if err != nil {
		return rules
	}
	for _, r := range set {
		if r.Default {
			rules.Rules = append(rules.Rules, r.Id)
		}
	}
	return rules
}
