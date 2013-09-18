package db

import (
	"github.com/godfried/impendulo/tool/junit"
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
)

func TestJUnitTest(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	test := junit.NewTest(bson.NewObjectId(), "name", "user", "pkg", junitData, junitData)
	err = AddJUnitTest(test)
	if err != nil {
		t.Error(err)
	}
	found, err := JUnitTest(bson.M{"_id": test.Id}, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(test, found) {
		t.Error("Tests don't match", test, found)
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
