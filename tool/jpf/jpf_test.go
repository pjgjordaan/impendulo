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

package jpf

import (
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	location := filepath.Join(os.TempDir(), "Racer")
	os.Mkdir(location, util.DPERM)
	defer os.RemoveAll(location)
	target := tool.NewTarget("Racer.java", "", location, tool.JAVA)
	err := util.SaveFile(target.FilePath(), srcFile)
	if err != nil {
		t.Errorf("Could not save file %q", err)
	}
	jpfConfig := NewConfig(bson.NewObjectId(), jpfBytes)
	jpf, err := New(jpfConfig, location)
	if err != nil {
		t.Errorf("Could not load jpf %q", err)
	}
	_, err = jpf.Run(bson.NewObjectId(), target)
	if err != nil {
		t.Errorf("Expected success, got %q", err)
	}
}

var srcFile = []byte(`
public class Racer implements Runnable {

     int d = 42;

     public void run () {
          doSomething(1001);                   // (1)
          d = 0;                               // (2)
     }

     public static void main (String[] args){
          Racer racer = new Racer();
          Thread t = new Thread(racer);
          t.start();

          doSomething(1000);                   // (3)
          int c = 420 / racer.d;               // (4)
          System.out.println(c);
     }
     
     static void doSomething (int n) {
          // not very interesting..
          try { Thread.sleep(n); } catch (InterruptedException ix) {}
     }
}
`)

var jpfBytes = []byte(`
listener=gov.nasa.jpf.listener.PreciseRaceDetector
`)
