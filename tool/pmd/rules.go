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
func RuleSet() (ret map[string]*Rule, err error) {
	if ruleSet != nil {
		return ruleSet, nil
	}
	cfgPath, err := config.Config(config.PMD_CFG)
	if err != nil {
		return ruleSet, err
	}
	cfg, err := os.Open(cfgPath)
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
