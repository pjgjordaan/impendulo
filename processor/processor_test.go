package processor

import (
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/tool/junit"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"

	"testing"
)

func TestProcessFile(t *testing.T) {
	db.Setup(db.TEST_CONN)
	db.DeleteDB(db.TEST_DB)
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	p := project.New("Triangle", "User", "Java", "A description")
	if e := db.Add(db.PROJECTS, p); e != nil {
		t.Error(e)
	}
	s := project.NewSubmission(p.Id, "student", project.FILE_MODE, p.Time+100)
	if e := db.Add(db.SUBMISSIONS, s); e != nil {
		t.Error(e)
	}
	f := &project.File{bson.NewObjectId(), s.Id, "Triangle.java", "triangle", project.SRC, s.Time + 100, srcBytes, bson.M{}, []*project.Comment{}}
	if e := db.Add(db.FILES, f); e != nil {
		t.Error(e)
	}
	dataBytes, e := util.ZipMap(dataMap)
	if e != nil {
		t.Errorf("Could not zip map %q", e)
	}
	target := &tool.Target{Name: "Triangle", Package: "triangle", Ext: "java"}
	test := &junit.Test{bson.NewObjectId(), p.Id, "AllTests.java", "testing", p.Time + 50, junit.DEFAULT, target, testBytes, dataBytes}
	if e := db.Add(db.TESTS, test); e != nil {
		t.Error(e)
	}
	ut := &junit.Test{bson.NewObjectId(), p.Id, "UserTests.java", "testing", p.Time + 150, junit.USER, target, userTestBytes, dataBytes}
	if e := db.Add(db.TESTS, ut); e != nil {
		t.Error(e)
	}
	tf := &project.File{bson.NewObjectId(), s.Id, "UserTests.java", "testing", project.TEST, s.Time + 200, userTestBytes, bson.M{}, []*project.Comment{}}
	if e := db.Add(db.FILES, tf); e != nil {
		t.Error(e)
	}
	proc, e := NewFileProcessor(s.Id)
	if e != nil {
		t.Error(e)
		return
	}
	if e = proc.Process(f.Id); e != nil {
		t.Error(e)
	}
	if e = proc.Process(tf.Id); e != nil {
		t.Error(e)
	}
}

var srcBytes = []byte(`
package triangle;
public class Triangle {
	public int maxpath(int[][] triangle) {
		int height = triangle.length - 2;
		for (int i = height; i >= 1; i--) {
			for (int j = 0; j <= i; j++) {
				triangle[i][j] += triangle[i + 1][j + 1] > triangle[i + 1][j] ? triangle[i + 1][j + 1]
						: triangle[i + 1][j];
			}
		}
		return triangle[0][0];
	}
}
`)

var testBytes = []byte(`
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
import triangle.Triangle;

public class AllTests {

	private static class FileTest extends TestCase {

		protected int inputData[][] = null;

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
				int h = (int) t.nval;
				inputData = new int[h][];
				for (int i = 0; i < h; i++) {
					inputData[i] = new int[i + 1];
					for (int j = 0; j <= i; j++) {
						t.nextToken();
						inputData[i][j] = (int) t.nval;
					}
				}
				t.nextToken();
				outputValue = (int) t.nval;
			} catch (Exception e) {
				e.printStackTrace();
				brokenTest = true;
			}
		}

		public void testMaxPath() {
			assertFalse(brokenTest);
			ExecutorService executor = Executors.newSingleThreadExecutor();
			int answer = -1;
			try {
				Future<Integer> future = executor
						.submit(new Callable<Integer>() {

							@Override
							public Integer call() throws Exception {
								return new Triangle().maxpath(inputData);
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
			testMaxPath();
		}
	}

	public static Test suite() {
		String location = System.getProperty("data.location");
		if (location == null || !new File(location).exists()) {
			location = "src" + File.separator + "testing" + File.separator
					+ "data";
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

var dataMap = map[string][]byte{
	"data/0001.txt": []byte("5 \n 1 \n 2 3 \n 4 5 6 \n 7 8 9 10 \n 11 12 13 14 15 \n 35"),
	"data/0002.txt": []byte("2 \n 3 \n 3 3 \n 6"),
	"data/0003.txt": []byte("1 \n 9 \n 9"),
	"data/0004.txt": []byte("4 \n 6 \n 6 7 \n 3 7 1 \n 8 1 1 1 \n 23"),
	"data/0005.txt": []byte("2 \n 2552 \n 8988 2808 \n 11540"),
}

var userTestBytes = []byte(`
package testing;

import junit.framework.TestCase;
import triangle.Triangle;

public class UserTests extends TestCase {

	public void testKselect() {
		Triangle t = new Triangle();
		int[][] values = { {6}, {6, 3}, {2, 9, 3}};
		assertEquals("Expected 21.", 21, t.maxpath(values));
	}
}
`)
