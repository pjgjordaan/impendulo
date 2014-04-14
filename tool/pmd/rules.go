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
	"fmt"

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
func (r *Rules) RuleArray() []string {
	a := make([]string, len(r.Rules))
	i := 0
	for rule := range r.Rules {
		a[i] = rule
		i++
	}
	return a
}

//String
func (r *Rules) String() string {
	s := "Rules:"
	for k, v := range r.Rules {
		s += fmt.Sprintf("\n\tRule %s Used %t", k, v)
	}
	return s
}

//NewRules creates a bew Rules struct from a set of rules for a given project.
//Each rule in the set is checked against the available rules.
func NewRules(projectId bson.ObjectId, rules map[string]bool) (*Rules, error) {
	s, e := RuleSet()
	if e != nil {
		return nil, e
	}
	for r := range rules {
		if _, ok := s[r]; !ok {
			delete(rules, r)
		}
	}
	return &Rules{
		Id:        bson.NewObjectId(),
		ProjectId: projectId,
		Rules:     rules,
	}, nil
}

//RuleSet loads the available rules from a json file which can be set
//via the config file.
func RuleSet() (map[string]*Rule, error) {
	if ruleSet != nil {
		return ruleSet, nil
	}
	cp, e := config.PMD_CFG.Path()
	if e != nil {
		return nil, e
	}
	c, e := os.Open(cp)
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(util.ReadBytes(c), &ruleSet); e != nil {
		return nil, e
	}
	return ruleSet, nil
}

//DefaultRules creates a default Rules struct from the available rules.
func DefaultRules(projectId bson.ObjectId) (*Rules, error) {
	s, e := RuleSet()
	if e != nil {
		return nil, e
	}
	r, e := NewRules(projectId, make(map[string]bool))
	if e != nil {
		return nil, e
	}
	for k, rule := range s {
		if rule.Default {
			r.Rules[k] = true
		}
	}
	return r, nil
}
