package jacoco

import (
	"encoding/xml"
	"io/ioutil"

	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"path/filepath"
	"testing"
	"time"
)

func TestProject(t *testing.T) {
	tmp, e := ioutil.TempDir("", "")
	if e != nil {
		panic(e)
	}
	baseDir := filepath.Join(tmp, "jacoco_build_test")
	resDir := filepath.Join(baseDir, "target")
	test := tool.NewTarget("UserTests.java", "testing", filepath.Join(baseDir, "test"), tool.JAVA)
	srcDir := filepath.Join(baseDir, "src")
	src := filepath.Join(srcDir, "kselect", "KSelect.java")
	if e := util.SaveFile(src, srcData); e != nil {
		t.Error(e)
	}
	//defer os.RemoveAll(baseDir)
	if e = util.SaveFile(test.FilePath(), userTest); e != nil {
		t.Error(e)
	}
	p, e := NewProject("Test", srcDir, resDir, test)
	if e != nil {
		t.Error(e)
	}
	data, e := xml.Marshal(p)
	if e != nil {
		t.Error(e)
	}
	path := filepath.Join(baseDir, "build.xml")
	if e = util.SaveFile(path, data); e != nil {
		t.Error(e)
	}
	r, e := tool.RunCommand([]string{"ant", "-f", path}, nil, 30*time.Second)
	if e != nil {
		t.Error(e)
	}
	if !util.Exists(filepath.Join(resDir, "report", "html")) || !util.Exists(filepath.Join(resDir, "report", "report.xml")) {
		t.Error("No report created.\n", string(r.StdErr), string(r.StdOut))
	}
}

func TestTool(t *testing.T) {
	if e := testTool("UserTests.java", "testing", userTest, nil); e != nil {
		t.Error(e)
	}
	if e := testTool("AllTests.java", "testing", normalTest, testData); e != nil {
		t.Error(e)
	}
}

func testTool(n, pkg string, t []byte, d map[string][]byte) error {
	tmp, e := ioutil.TempDir("", "")
	if e != nil {
		return e
	}
	baseDir := filepath.Join(tmp, "jacoco_build_test")
	test := tool.NewTarget(n, pkg, filepath.Join(baseDir, "test"), tool.JAVA)
	srcDir := filepath.Join(baseDir, "src")
	src := filepath.Join(srcDir, "kselect", "KSelect.java")
	if e := util.SaveFile(src, srcData); e != nil {
		return e
	}
	//defer os.RemoveAll(baseDir)
	if e := util.SaveFile(test.FilePath(), t); e != nil {
		return e
	}
	z, e := util.ZipMap(d)
	if e != nil {
		return e
	}
	if e = util.Unzip(test.PackagePath(), z); e != nil {
		return e
	}
	target := &tool.Target{Name: "KSelect", Ext: "java", Package: "kselect"}
	cov, e := New(baseDir, srcDir, test, target, bson.NewObjectId())
	if e != nil {
		return e
	}
	_, e = cov.Run(bson.NewObjectId(), tool.NewTarget("KSelect.java", "kselect", srcDir, tool.JAVA))
	return e
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

var userTest = []byte(`
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

var normalTest = []byte(`
package testing;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.StreamTokenizer;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

import junit.framework.Test;
import junit.framework.TestCase;
import junit.framework.TestSuite;
import kselect.KSelect;



public class AllTests {

	private static class FileTest extends TestCase {

		protected int inputData[] = null;
		protected int inputK = 0;
		protected long outputValue = 0;
		protected boolean brokenTest = false;

		protected String testCase;

		public FileTest(String s, String testCase) {
			super(s);
			this.testCase = testCase;
			try {
				BufferedReader r = new BufferedReader(new FileReader(getName()));
				StreamTokenizer t = new StreamTokenizer(r);
				t.parseNumbers();
				t.nextToken();
				inputK = (int) t.nval;
				t.nextToken();
				int n = (int) t.nval;
				inputData = new int[2 * n];
				for (int i = 0; i < 2 * n; i++) {
					t.nextToken();
					inputData[i] = (int) t.nval;
				}
				t.nextToken();
				outputValue = (int) t.nval;
			} catch (Exception e) {
				e.printStackTrace();
				brokenTest = true;
			}
		}

		public void testSelect() {
			assertFalse(brokenTest);
			ExecutorService executor = Executors.newSingleThreadExecutor();
			int answer = -1;
			try {
				Future<Integer> future = executor
						.submit(new Callable<Integer>() {

							@Override
							public Integer call() throws Exception {
								return new KSelect().kselect(inputK, inputData);
							}

						});
				answer = future.get(30, TimeUnit.SECONDS);
			} catch (TimeoutException te) {
				fail(String.format("Test %s took too long.", testCase));
			} catch (ExecutionException e) {
				fail(String.format("Test %s failed with message %s.", testCase,
						e.getMessage()));
			} catch (InterruptedException e) {
				fail(String.format("Test %s interrupted.", testCase));
			}
			if (answer != outputValue)
				System.out.println(answer + " " + outputValue);
			assertTrue(String.format(
					"Wrong answer %d , should be %d in testCase %s.", answer,
					outputValue, testCase), answer == outputValue);
		}

		public void runTest() {
			testSelect();
		}
	}

	public static Test suite() {
		String location = System.getProperty("data.location");
		if (location == null || !new File(location).exists()) {
			location = "src" + File.separator + "testing" + File.separator
					+ "data";
                        //throw new Exception("no location" + location);
		} else if(location != null){
                        //throw new Exception("found location" + location);
                }
		TestSuite suite = new TestSuite("Test for Triangle");
		File f = new File(location);
		String s[] = f.list();
		for (int i = 0; i < s.length; i++) {
			String n = s[i];
			if (n.endsWith(".txt")) {
				suite.addTest(new FileTest(location + File.separator + n, n));
			}
		}
		return suite;
	}

}`)

var testData = map[string][]byte{
	"data/0001.txt": []byte("-2\n1\n911 911\n0"),
	"data/0002.txt": []byte("-1\n1\n911 911\n1"),
	"data/0003.txt": []byte("0\n3\n789 789\n123 123\n456 456\n0"),
	"data/0004.txt": []byte("0\n5\n1730944 1043380\n2038072 1286400\n2207048 1459547\n2138646 1426646\n1917940 1151782\n0"),
	"data/0005.txt": []byte("-3\n3\n789 789\n123 123\n456 456\n2"),
}
