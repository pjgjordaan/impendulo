package db

import (
	"bytes"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool/javac"
	"github.com/godfried/impendulo/user"
	"labix.org/v2/mgo/bson"
	"reflect"
	"strconv"
	"testing"
)

func TestSetup(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := getSession()
	if err != nil{
		t.Error(err)
	}
	defer s.Close()
}

func TestRemoveFile(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := getSession()
	if err != nil{
		t.Error(err)
	}
	defer s.Close()
	f, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = AddFile(f)
	if err != nil {
		t.Error(err)
	}
	err = RemoveFileById(f.Id)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": f.Id}
	f, err = GetFile(matcher, nil)
	if f != nil || err == nil {
		t.Error("File not deleted")
	}
}

func TestGetFile(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := getSession()
	if err != nil{
		t.Error(err)
	}
	defer s.Close()
	f, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = AddFile(f)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": f.Id}
	dbFile, err := GetFile(matcher, nil)
	if err != nil {
		t.Error(err)
	}
	if !f.Equals(dbFile) {
		t.Error("Files not equivalent", f.String() == dbFile.String(), bytes.Equal(f.Data, dbFile.Data))
	}
}

func TestGetSubmission(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := getSession()
	if err != nil{
		t.Error(err)
	}
	defer s.Close()
	sub := project.NewSubmission(bson.NewObjectId(), "user", project.FILE_MODE, 1000)
	err = AddSubmission(sub)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": sub.Id}
	dbSub, err := GetSubmission(matcher, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(sub, dbSub) {
		t.Error("Submissions not equivalent")
	}
}

func TestGetResult(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := getSession()
	if err != nil{
		t.Error(err)
	}
	defer s.Close()
	file, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = AddFile(file)
	if err != nil {
		t.Error(err)
	}
	res := javac.NewResult(file.Id, fileData)
	err = AddResult(res)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": res.GetId()}
	dbRes, err := GetToolResult(res.GetName(), matcher, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(res, dbRes) {
		t.Error("Results not equivalent")
	}
}

func TestGetUserById(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := getSession()
	if err != nil{
		t.Error(err)
	}	
	defer s.Close()
	u := user.New("uname", "pword")
	err = AddUser(u)
	if err != nil {
		t.Error(err)
	}
	found, err := GetUserById("uname")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(u, found) {
		t.Error("Users not equivalent", u, found)
	}
}

func TestGetTest(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := getSession()
	if err != nil{
		t.Error(err)
	}
	defer s.Close()
	test := project.NewTest(bson.NewObjectId(), "name", "user", "pkg", testData, fileData)
	err = AddTest(test)
	if err != nil {
		t.Error(err)
	}
	found, err := GetTest(bson.M{"_id": test.Id}, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(test, found) {
		t.Error("Tests don't match", test, found)
	}
}

func TestCount(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := getSession()
	if err != nil{
		t.Error(err)
	}	
	defer s.Close()
	num := 100
	n, err := Count(PROJECTS, bson.M{})
	if err != nil {
		t.Error(err)
	}
	if n != 0 {
		t.Errorf("Invalid count %q, should be %q", n, 0)
	}
	for i := 0; i < num; i++ {
		var s int = i / 10
		err = AddProject(project.NewProject("name"+strconv.Itoa(s), "user", "lang", fileData))
		if err != nil {
			t.Error(err)
		}
	}
	n, err = Count(PROJECTS, bson.M{})
	if err != nil {
		t.Error(err)
	}
	if n != num {
		t.Errorf("Invalid count %q, should be %q", n, num)
	}
	n, err = Count(PROJECTS, bson.M{"name": "name0"})
	if err != nil {
		t.Error(err)
	}
	if n != 10 {
		t.Errorf("Invalid count %q, should be %q", n, 10)
	}

}

var fileInfo = bson.M{project.TIME: 1000, project.TYPE: project.SRC, project.MOD: "c", project.NAME: "Triangle.java", project.FTYPE: "java", project.PKG: "triangle", project.NUM: 1000}

var fileData = []byte(`[Faust:] "I, Johannes Faust, do call upon thee, Mephistopheles!"

[Faust:]
O growing Moon, didst thou but shine
A last time on this pain of mine
Behind this desk how oft have I
At midnight seen thee rising high
O'er book and paper I bend
Thou didst appear, o mournful friend

[Mephistopheles:]
I am the spirit that ever denies!
And justly so: for all that is born
Deserves to be destroyed in scorn
Therefore 'twere best if nothing were created
Destruction, sin, wickedness - plainly stated
All of which you as evil have classified
That is my element - there I abide

[Manager: ]
Scatter the stars with a lavish hand
Water, fire, tavern wall
Birds and beasts, all within command
Thus in our narrow booth today
Creation's ample scope display
Wander swiftly, observing well
From the Heavens, to the World, to Hell!

The World of Spirits is not barred to thee!

[Mephistopheles:] "Now then, Faustus. What wouldst thou have Mephisto do?"
[Faust:]
"I charge thee, Mephisto, wait upon me while I live... to do whatever Faustus shall command. Be it to make the moon drop from outer sphere, or the ocean to overwhelm the world. Go bear these tidings to great Lucifer: say he surrenders up his soul. So that he shall spare him four and twenty years, letting him live in all voluptiousness, having thee ever to attend on me. To give me whatsoever I shall ask."

[Mephistopheles:] "I will."

[Faust:]
Sublime spirit, thou hast given me all
All for which I besought thee, not in vain
Didst thou reveal thy countenance in the fire
Thou hast given me Nature for a kingdom
With the power to enjoy and feel
Only a visit of chilling bewilderment
Thou [then me?] bringest all the living creatures
And taught me to know my brothers in the Air
In the deep waters and in the silent coverts
When through the forest the storm rages
Uprooting the giant pines which in their fall
Crushing, drag down neighboring boughs and trunks
Whose [growingly?] hollow thunder shake the hills
Then thou dost lead me to a sheltering cave
And revealest me to myself and layest bare
The deep mysterious miracle of my Nature
And when the pure moon rises into sight
Soothingly above me, then about me hover
Creeping from rocky walls and dewy thickets
Silver shadows, phantoms of a bygone world
Which allay the austere joy of meditation

Now fully do I realize that Man
Can never possess perfection
With this ecstasy which brings me near and nearer
To the Gods

[Margarete: ]
My mother the harlot put me to death
My father the scoundrel ate my flesh
My dear little sister laid all my bones
In a dark shaded place under the stones
Then I changed into a wood-bird
Fluttering away
Fly away

[Mephistopheles:]
Mankind, that foolish Cosmos
Always acts as incomplete
He thought himself to Be
I am part of that part which was the Absolute
A part of that Darkness which gave birth to Light
The arrogant Light which would dispute
Ancient rank of Mother Night
Therefore I hope it won't be long before
With matter it shall perish evermore!

[Manager: ]
Scatter the stars with a lavish hand
Water, fire, tavern wall
Birds and beasts, all within command
Thus in our narrow booth today
Creation's ample scope display
Wander swiftly, observing well
From the Heavens to the World

The World of Spirits is not barred to thee!

[Faust:] "So, still I seek the force, the reason governing life's flow, and not just its external show."
[Mephistopheles:] "The governing force? The reason? Some things cannot be known; they are beyond your reach even when shown."
[Faust:] "Why should that be so?"
[Mephistopheles:] "They lie outside the boundaries that words can address; and man can only grasp those thoughts which language can express."
[Faust:] "What? Do you mean that words are greater yet than man?"
[Mephistopheles:] "Indeed they are."
[Faust:] "Then what of longing? Affection, pain or grief? I can't describe these, yet I know they are in my breast. What are they?""
[Mephistopheles:] "Without substance, as mist is."
[Faust:] "In that case man is only air as well!"`)

var testData = []byte(`Szénizotóp, szénizotóp,
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
