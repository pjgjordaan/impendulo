package junit

import (
	"testing"
	"os"
	"labix.org/v2/mgo/bson"
	"github.com/godfried/impendulo/util"
	"github.com/godfried/impendulo/project"
	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/config"
	"path/filepath"
)

func TestRun(t *testing.T){
	location := filepath.Join(os.TempDir(), "Triangle")
	srcLocation := filepath.Join(location, "triangle")
	testLocation := filepath.Join(location, "testing")
	dataLocation := filepath.Join(testLocation, "data")
	os.Mkdir(location, util.DPERM)
	defer os.RemoveAll(location)
	os.Mkdir(srcLocation, util.DPERM)
	os.Mkdir(testLocation, util.DPERM)
	err := util.Copy(location, config.GetConfig(config.TESTING_DIR))
	if err != nil{
		t.Errorf("Could not copy directory %q", err)
	}
	target := tool.NewTarget("Triangle.java", 
		project.JAVA, "triangle", location)
	testTarget := tool.NewTarget("AllTests.java", 
		project.JAVA, "testing", location)
	err = util.SaveFile(target.FilePath(), file)
	if err != nil{
		t.Errorf("Could not save file %q", err)
	}
	err = util.SaveFile(testTarget.FilePath(), test)
	if err != nil{
		t.Errorf("Could not save file %q", err)
	}
	for name, data := range testData{
		err = util.SaveFile(filepath.Join(dataLocation, name), data)
		if err != nil{
			t.Errorf("Could not save file %q", err)
		}
	}
	junit := New(location, location, dataLocation)
	_, err = junit.Run(bson.NewObjectId(), testTarget)
	if err != nil{
		t.Errorf("Expected success, got %q", err)
	}
	err = util.SaveFile(target.FilePath(), file2)
	if err != nil{
		t.Errorf("Could not save file %q", err)
	}
	_, err = junit.Run(bson.NewObjectId(), testTarget)
	if err == nil{
		t.Errorf("Expected error.")
	}
	target = tool.NewTarget("File.java", 
		project.JAVA, "", location)
	_, err = junit.Run(bson.NewObjectId(), testTarget)
	if err == nil{
		t.Error("Expected error")
	}
}

var file = []byte(`
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


var file2 = []byte(`
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

var test = []byte(`
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

var testData = map[string][]byte{
	"0001.txt": []byte("5 \n 1 \n 2 3 \n 4 5 6 \n 7 8 9 10 \n 11 12 13 14 15 \n 35"),
	"0002.txt": []byte("2 \n 3 \n 3 3 \n 6"),
	"0003.txt": []byte("1 \n 9 \n 9"),
	"0004.txt": []byte("4 \n 6 \n 6 7 \n 3 7 1 \n 8 1 1 1 \n 23"),
	"0005.txt": []byte("2 \n 2552 \n 8988 2808 \n 11540"),
}