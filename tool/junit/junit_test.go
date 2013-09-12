package junit

import (
	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	location := filepath.Join(os.TempDir(), "Triangle")
	srcLocation := filepath.Join(location, "triangle")
	testLocation := filepath.Join(location, "testing")
	os.Mkdir(location, util.DPERM)
	defer os.RemoveAll(location)
	os.Mkdir(srcLocation, util.DPERM)
	os.Mkdir(testLocation, util.DPERM)
	err := util.Copy(location, config.Config(config.TESTING_DIR))
	if err != nil {
		t.Errorf("Could not copy directory %q", err)
	}
	target := tool.NewTarget("Triangle.java",
		tool.JAVA, "triangle", location)
	err = util.SaveFile(target.FilePath(), validFile)
	if err != nil {
		t.Errorf("Could not save file %q", err)
	}
	dataBytes, err := util.ZipMap(dataMap)
	if err != nil {
		t.Errorf("Could not zip map %q", err)
	}
	test := NewTest(bson.NewObjectId(), "AllTests.java", "user", "testing", testBytes, dataBytes)
	junit, err := New(test, location)
	if err != nil {
		t.Errorf("Expected success, got %q", err)
	}
	_, err = junit.Run(bson.NewObjectId(), target)
	if err != nil {
		t.Errorf("Expected success, got %q", err)
	}
	err = util.SaveFile(target.FilePath(), invalidFile)
	if err != nil {
		t.Errorf("Could not save file %q", err)
	}
	_, err = junit.Run(bson.NewObjectId(), target)
	if err == nil {
		t.Errorf("Expected error.")
	}
	target = tool.NewTarget("File.java",
		tool.JAVA, "", location)
	_, err = junit.Run(bson.NewObjectId(), target)
	if err == nil {
		t.Error("Expected error")
	}
}

var validFile = []byte(`
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

var invalidFile = []byte(`
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
