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

package db

import (
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/junit"
	"labix.org/v2/mgo/bson"

	"reflect"
	"testing"
)

func TestJUnitTest(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, e := Session()
	if e != nil {
		t.Error(e)
	}
	defer s.Close()
	test := junit.NewTest(bson.NewObjectId(), "name", junit.DEFAULT, &tool.Target{}, junitData, junitData)
	if e = AddJUnitTest(test); e != nil {
		t.Error(e)
	}
	if v, e := JUnitTest(bson.M{"_id": test.Id}, nil); e != nil {
		t.Error(e)
	} else if !reflect.DeepEqual(test, v) {
		t.Error("Tests don't match", test, v)
	}
}

var junitData = []byte(`Szénizotóp, szénizotóp,
süss fel!

Szénizotópmalom karjai járnak
új Nanováros fényeinél,
járnak és járnak és szintetizálnak,
éljen a Haladás, éljen a Fény!

Hidrogénhíd tör a tiszta jövõbe!
Elõre, elõre!
Héjakra, gyûrûkre, mezonmezõre!
Elõre, elõre!
Hallgatag szénmedence népe,
elõre, mind elõre!

De tûnt idõ, te merre bolyongsz az anyagban?
Visszatérsz-e még a nyüzsgõ szálakon?
Rétegek, halmazok, iramló pályák,
ez vagyok én, és ez itt az otthonom.

Róka hasa telelõ, mélyén folyik az idõ,
alvó libalegelõ, zúgó libalegelõ.
Kádam vizén a hajó, bentrõl szól egy rádió -
éjjel anya hallható, nappal apa hogyha jó az adó.

Róka hasa telelõ, felhõn gurul az idõ,
halkan rezeg a mezõ mélyén valami erõ.
Este leesik a hó, csend van, kiköt a hajó.
Lámpás téli kikötõ mélyén molekula nõ idebenn.

Mint ahogy látjuk, apró, molekuláris gépek azok,
amelyek ezt a mozgást végzik.
Hangsúlyozom mégegyszer, a molekulák szintjén
egy sejtben nyolcmilliárd fehérjemolekula fordul elõ
és ezek a parányi kis gépek végzik összehangoltan a mozgásokat.

Fordul a gép!

A vágtató ló mozgása esetén az izomrostokban fehérjék,
az aktin- és miozinszálak egymásra csúszása idézi elõ a mozgást tulajdonképpen,
és akkor is, amikor felemelem a kezemet, az izmaimban, az izomsejtekben
ezek a fehérjeszálak csúsznak egymásba.

Fordul a gép,
Folyik el az élet.

És ezek a másodlagos kölcsönhatások szobahõmérsékleten, tehát az élet hõmérsékletén
örökösen felhasadnak a hõmozgás energiája folytán, de csak egy-egy kötés hasad fel,
tehát maga a szerkezet fennmarad egységesen, ugyanakkor bizonyos elemei képesek
elég jelentõs atomi szintû mozgásokra.
Ennek a következménye az az elõzõ ábrán szemléltetett
nyüzsgés, mozgás, amit láttunk.
Tehát a fehérjék térszerkezete örökös nyüzsgésben van
szobahõmérsékleten, és ez a fajta flexibilitás teszi lehetõvé azt, hogy a fehérjék, mint
molekuláris gépek, atomi mozgások végrehajtására képesek.`)
