package pmd

import (
	"encoding/json"
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
)

type (
	//Rule is the descriptive representation of a PMD rule.
	//It is used when a PMD configuration is chosen.
	Rule struct {
		Id          string
		Name        string
		Description string
		Default     bool
	}
	//Rules specifies the PMD rules configured for a specific project.
	//It only stores each rule's identifier.
	Rules struct {
		Id        bson.ObjectId   "_id"
		ProjectId bson.ObjectId   "projectid"
		Rules     map[string]bool "rules"
	}
)

var (
	ruleSet map[string]*Rule
)

//RuleArray extracts an array of pmd rule identifiers from a Rules struct.
func (this *Rules) RuleArray() []string {
	ret := make([]string, len(this.Rules))
	i := 0
	for r := range this.Rules {
		ret[i] = r
		i++
	}
	return ret
}

//NewRules creates a bew Rules struct from a set of rules for a given project.
//Each rule in the set is checked against the available rules.
func NewRules(projectId bson.ObjectId, rules map[string]bool) (ret *Rules, err error) {
	valid, err := RuleSet()
	if err != nil {
		return
	}
	for r := range rules {
		if _, ok := valid[r]; !ok {
			delete(rules, r)
		}
	}
	ret = &Rules{
		Id:        bson.NewObjectId(),
		ProjectId: projectId,
		Rules:     rules,
	}
	return
}

//RuleSet loads the available rules from a json file which can be set
//via the config file.
func RuleSet() (map[string]*Rule, error) {
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

//DefaultRules creates a default Rules struct from the available rules.
func DefaultRules(projectId bson.ObjectId) (rules *Rules, err error) {
	set, err := RuleSet()
	if err != nil {
		return
	}
	rules, err = NewRules(projectId, make(map[string]bool))
	if err != nil {
		return
	}
	for k, rule := range set {
		if rule.Default {
			rules.Rules[k] = true
		}
	}
	return
}
