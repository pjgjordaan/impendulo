package testing;

import java.io.File;
import java.io.IOException;
import java.io.OutputStream;
import java.io.PrintStream;
import java.security.InvalidParameterException;

import org.apache.tools.ant.Project;
import org.apache.tools.ant.taskdefs.optional.junit.FormatterElement;
import org.apache.tools.ant.taskdefs.optional.junit.JUnitTask;
import org.apache.tools.ant.taskdefs.optional.junit.JUnitTest;

public class TestRunner {

	public static void main(String[] args) {
		if (args.length != 2) {
			throw new InvalidParameterException("Expected 2 arguments.");
		}
		PrintStream out = System.out;
		PrintStream err = System.err;
		System.setOut(new PrintStream(new OutputStream() {
			@Override
			public void write(int b) throws IOException {
			}
		}));
		System.setErr(new PrintStream(new OutputStream() {
			@Override
			public void write(int b) throws IOException {
			}
		}));
		String testExec = args[0];
		String[] split = testExec.split("\\.");
		String testName = split[split.length - 1];
		String dataLocation = args[1];
		System.setProperty("data.location", dataLocation);
		Project project = new Project();
		JUnitTask task;
		try {
			task = new JUnitTask();
			project.setProperty("java.io.tmpdir",
					System.getProperty("java.io.tmpdir"));
			task.setProject(project);
			JUnitTask.SummaryAttribute sa = new JUnitTask.SummaryAttribute();
			sa.setValue("on");
			task.setFork(false);
			task.setPrintsummary(sa);
			FormatterElement formater = new FormatterElement();
			FormatterElement.TypeAttribute type = new FormatterElement.TypeAttribute();
			type.setValue("xml");
			formater.setType(type);
			task.addFormatter(formater);
			JUnitTest test = new JUnitTest(testExec);
			test.setOutfile(testName + "_junit");
			test.setTodir(new File(dataLocation));
			task.addTest(test);
			task.execute();
		} catch (Exception e) {
			try {
				err.write(e.getMessage().getBytes());
			} catch (IOException e1) {
			}
		} finally {
			System.setOut(out);
			System.setErr(err);
			System.exit(0);
		}
	}

}
