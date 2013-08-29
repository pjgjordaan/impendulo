package processing

import (
	//	"github.com/godfried/impendulo/config"
	"github.com/godfried/impendulo/db"
	"github.com/godfried/impendulo/project"
	//	"github.com/godfried/impendulo/tool"
	"github.com/godfried/impendulo/util"
	"labix.org/v2/mgo/bson"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupTests(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	s := getSubmission()
	err := db.AddSubmission(s)
	if err != nil {
		t.Error(err)
	}
	test, err := getTest(s.ProjectId)
	if err != nil {
		t.Error(err)
	}
	err = db.AddTest(test)
	if err != nil {
		t.Error(err)
	}
	dir := filepath.Join(os.TempDir(), s.Id.Hex())
	defer os.RemoveAll(dir)
	runners, err := SetupTests(test.ProjectId, dir)
	if runners == nil {
		t.Error("Could not set up tests properly")
	}

}

/*
func TestCompileTest(t *testing.T) {
	err := config.LoadConfigs("../config.txt")
	if err != nil {
		panic(err)
	}
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	s := getSubmission()
	err = db.AddSubmission(s)
	if err != nil {
		t.Error(err)
	}
	f := getFile(s.Id)
	err = db.AddFile(f)
	if err != nil {
		t.Error(err)
	}
	test, err := getTest()
	if err != nil {
		t.Error(err)
	}
	err = db.AddTest(test)
	if err != nil {
		t.Error(err)
	}
	dir := filepath.Join(os.TempDir(), f.Id.Hex())
	defer os.RemoveAll(dir)
	runners, err := SetupTests(test.ProjectId, dir)
	if err != nil {
		t.Error(err)
	}
	if runners == nil {
		t.Error("Could not set up tests properly")
	}
	ti, err := ExtractFile(f, dir)
	if err != nil {
		t.Error(err)
	}
	target := tool.NewTarget(runner.Project, test.Names[0], runner.Lang, runner.Package, runner.Dir)
	ok, err := runner.Compile(target, f, ti.Dir)
	if !ok || err != nil {
		t.Error(err)
	}
	f = getFile(s.Id)
	f.Data = errSrc
	err = db.AddFile(f)
	if err != nil {
		t.Error(err)
	}
	dir = filepath.Join(os.TempDir(), f.Id.Hex())
	defer os.RemoveAll(dir)
	runner = SetupTests(test.Project, test.Lang, dir)
	if runner == nil {
		t.Error("Could not set up tests properly")
	}
	ti, err = ExtractFile(f, dir)
	if err != nil {
		t.Error(err)
	}
	target = tool.NewTarget(runner.Project, test.Names[0], runner.Lang, runner.Package, runner.Dir)
	ok, err = runner.Compile(target, f, ti.Dir)
	if ok && err == nil {
		t.Error("Expected no compile")
	}

}

func TestExecute(t *testing.T) {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	s := getSubmission()
	err := db.AddSubmission(s)
	if err != nil {
		t.Error(err)
	}
	f := getFile(s.Id)
	err = db.AddFile(f)
	if err != nil {
		t.Error(err)
	}
	test, err := getTest()
	if err != nil {
		t.Error(err)
	}
	err = db.AddTest(test)
	if err != nil {
		t.Error(err)
	}
	dir := filepath.Join(os.TempDir(), s.Id.Hex())
	defer os.RemoveAll(dir)
	runner := SetupTests(test.Project, test.Lang, dir)
	if runner == nil {
		t.Error("Could not set up tests properly")
	}
	ti, err := ExtractFile(f, dir)
	if err != nil {
		t.Error(err)
	}
	err = runner.Execute(f, ti.Dir)
	if err != nil {
		t.Error(err)
	}
}
*/
func getSubmission() *project.Submission {
	p := project.NewProject("Triangle", "user", "java", []byte{})
	db.AddProject(p)
	return project.NewSubmission(p.Id, p.User, project.FILE_MODE, 1000)
}

func getTest(pId bson.ObjectId) (*project.Test, error) {
	testZip, err := util.ZipMap(map[string][]byte{"testing/EasyTests.java": testData})
	if err != nil {
		return nil, err
	}
	dataZip, err := util.ZipMap(map[string][]byte{"0001.etxt": testCase})
	if err != nil {
		return nil, err
	}
	return project.NewTest(pId, "EasyTests.java", "user", "testing", testZip, dataZip), nil
}

func getFile(subId bson.ObjectId) *project.File {
	info := bson.M{project.TIME: 1000, project.TYPE: project.SRC, project.MOD: 'c', project.NAME: "Triangle.java", project.FTYPE: "java", project.PKG: "triangle", project.NUM: 100}
	file, _ := project.NewFile(subId, info, srcData)
	return file
}

var errSrc = []byte(`package triangle;


public class {
	public int maxpath(int[] tri) {
		int l = tri.length;
		for (int i = l - 2; i >= 0; i--){
			for (int j = 0; j <= i; j++){
				tri[i][j] += tri[i + 1][j] > tri[i + 1][j + 1] ? tri[i + 1][j]
						:tri[i + 1][j + 1];
			}
		}
		return tri[0][0];
	}
}`)

var srcData = []byte(`package triangle;


public class Triangle {
	public int maxpath(int[][] tri) {
		int l = tri.length;
		for (int i = l - 2; i >= 0; i--){
			for (int j = 0; j <= i; j++){
				tri[i][j] += tri[i + 1][j] > tri[i + 1][j + 1] ? tri[i + 1][j]
						:tri[i + 1][j + 1];
			}
		}
		return tri[0][0];
	}
}`)

var testCase = []byte(`1
9
9
`)

var testData = []byte(`package testing;

import triangle.Triangle;
import junit.framework.Test;
import junit.framework.TestSuite;
import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.io.StreamTokenizer;

import junit.framework.TestCase;

public class EasyTests {

	public static long startTime = -1;

	public static long timeLimit = 2 * 60 * 1000;
	public final static String DATA_LOCATION_PROPERTY = "data.location";

	private static class FileTest extends TestCase {

		protected int inputData[][] = null;

		protected long outputValue = 0;

		protected boolean brokenTest = false;

		public FileTest(String s) {
			super(s);
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

		public void runTest() {
			if (startTime == -1) {
				startTime = System.currentTimeMillis();
			}
			else if (System.currentTimeMillis() - startTime > timeLimit) {
				assertTrue("Out of time", false);
			}
			else {
				assertFalse(brokenTest);
				Triangle tri = new Triangle();
				int answer = tri.maxpath(inputData);
				assertTrue("Wrong answer " + answer + ", should be " + outputValue, answer == outputValue);
			}
		}
	}

	public static Test suite() {
		TestSuite suite = new TestSuite("Test for triangle");
		String loc = System.getProperty(DATA_LOCATION_PROPERTY);
		File f = new File(loc);
		String s[] = f.list();
                for (int i = 0; i < s.length; i++) {
			String n = s[i];
			if (n.endsWith(".etxt")) {
				suite.addTest(new FileTest(loc + File.separator+ n));
			}
		}
		return suite;
	}

}`)
