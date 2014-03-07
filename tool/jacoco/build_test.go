package jacoco

import (
	"encoding/xml"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"testing"
)

func TestProject(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "jacoco_build_test")
	resDir := filepath.Join(baseDir, "target")
	test := tool.NewTarget("UserTests.java", "testing", filepath.Join(baseDir, "test"), tool.JAVA)
	srcDir := filepath.Join(baseDir, "src")
	src := filepath.Join(srcDir, "kselect", "KSelect.java")
	err := util.SaveFile(src, srcData)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(baseDir)
		if err != nil {
			t.Error(err)
		}
	}()
	err = util.SaveFile(test.FilePath(), testData)
	if err != nil {
		t.Error(err)
	}
	p, err := NewProject("Test", srcDir, resDir, test)
	if err != nil {
		t.Error(err)
	}
	data, err := xml.Marshal(p)
	if err != nil {
		t.Error(err)
	}
	path := filepath.Join(baseDir, "build.xml")
	err = util.SaveFile(path, data)
	if err != nil {
		t.Error(err)
	}
	res := tool.RunCommand([]string{"ant", "-f", path}, nil)
	if res.Err != nil {
		t.Error(res.Err)
	}
	if !util.Exists(filepath.Join(resDir, "report", "html")) || !util.Exists(filepath.Join(resDir, "report", "report.xml")) {
		t.Error("No report created.\n", string(res.StdErr), string(res.StdOut))
	}
}

func TestTool(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "jacoco_build_test")
	test := tool.NewTarget("UserTests.java", "testing", filepath.Join(baseDir, "test"), tool.JAVA)
	srcDir := filepath.Join(baseDir, "src")
	src := filepath.Join(srcDir, "kselect", "KSelect.java")
	err := util.SaveFile(src, srcData)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(baseDir)
		if err != nil {
			t.Error(err)
		}
	}()
	err = util.SaveFile(test.FilePath(), testData)
	if err != nil {
		t.Error(err)
	}
	cov, err := New(baseDir, srcDir, test)
	if err != nil {
		t.Error(err)
	}
	r, err := cov.Run(bson.NewObjectId(), tool.NewTarget("KSelect.java", "kselect", srcDir, tool.JAVA))
	if err != nil {
		t.Error(err)
	}
	t.Log(string(r.GetReport().(*Report).HTML))
}

var srcData = []byte(`
package kselect;

public class KSelect {
	
	

	public int kselect(int k, int[] values) {
		//int kpos = 0;
		int pair1 = 0;
		int pair2 = 0;
		int[] order = values.clone();
		int cnt = 0;
		
		if (k > values.length/2) {
			return 0;
		}
		if (k > 0){
			
			for (int i = 0; i < values.length-2; i=i+2) {
				for(int j = 2; j < values.length; j=j+2) {
					if (values[i] < values[j]) {
						int tmp = values[i];
						int tmp1 = values[i+1];
						values[i] = values[j];
						values[i+1] = values[j+1];
						values[j] = tmp;
						values[j+1] = tmp1;
					} else if ((values[i] == values[j]) && values[i+1] > values[j+1]) {
						int tmp = values[i];
						int tmp1 = values[i+1];
						values[i] = values[j];
						values[i+1] = values[j+1];
						values[j] = tmp;
						values[j+1] = tmp1;
					}
				}
			}
			
			for (int i = 0; i < values.length; i=i+2) {
				cnt++;
				if (cnt == k) {
					pair1 = values[i];
					pair2 = values[i+1];
				}
				System.out.println(values[i] + "," + values[i+1]);
			}
			cnt = 0;
			for (int i = 0; i < order.length; i=i+2) {
				cnt++;
				if ((order[i] == pair1) && (order[i+1] == pair2)) {
					k = cnt;
				}
			}
			
		}
		
		
		return k;
	}
	
	public static void main(String[] args) {
		int []values = {3,1,4,1,5,9,2,6,5,3,5,8}; 
		//kselect(1, values);
	}
}
`)

var testData = []byte(`
package testing;

import junit.framework.TestCase;
import kselect.KSelect;

public class UserTests extends TestCase {

	public void testKselect() {
		KSelect k = new KSelect();
		int[] values = { 1, 4, 2, 3, 7, 1, 2, 1, 4, 2 };
		assertEquals("Expected 4th pair.", 4, k.kselect(2, values));
	}
}
`)
