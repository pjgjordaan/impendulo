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
