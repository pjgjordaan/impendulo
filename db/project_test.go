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

package db

import (
	"bytes"
	"github.com/godfried/impendulo/project"
	"labix.org/v2/mgo/bson"
	"reflect"
	"testing"
)

func TestRemoveFile(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	f, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = Add(FILES, f)
	if err != nil {
		t.Error(err)
	}
	err = RemoveFileById(f.Id)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": f.Id}
	f, err = File(matcher, nil)
	if f != nil || err == nil {
		t.Error("File not deleted")
	}
}

func TestFile(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	f, err := project.NewFile(bson.NewObjectId(), fileInfo, fileData)
	if err != nil {
		t.Error(err)
	}
	err = Add(FILES, f)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": f.Id}
	dbFile, err := File(matcher, nil)
	if err != nil {
		t.Error(err)
	}
	if !f.Equals(dbFile) {
		t.Error("Files not equivalent", f.String() == dbFile.String(), bytes.Equal(f.Data, dbFile.Data))
	}
}

func TestSubmission(t *testing.T) {
	Setup(TEST_CONN)
	defer DeleteDB(TEST_DB)
	s, err := Session()
	if err != nil {
		t.Error(err)
	}
	defer s.Close()
	sub := project.NewSubmission(bson.NewObjectId(), "user", project.FILE_MODE, 1000)
	err = Add(SUBMISSIONS, sub)
	if err != nil {
		t.Error(err)
	}
	matcher := bson.M{"_id": sub.Id}
	dbSub, err := Submission(matcher, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(sub, dbSub) {
		t.Error("Submissions not equivalent")
	}
}

var fileInfo = bson.M{
	project.TIME: 1000,
	project.TYPE: project.SRC,
	project.NAME: "Triangle.java",
	project.PKG:  "triangle",
}

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
