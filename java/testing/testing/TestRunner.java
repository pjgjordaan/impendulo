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

/**
 * This class is used to run a JUnit Test on a Java class. The results of the
 * Test is stored in a XML file. We use ant libraries to run the Test and
 * generate the XML file. See
 * http://godoc.org/github.com/godfried/impendulo/tool/junit#Tool for more
 * information.
 * 
 * @author godfried
 * 
 */
public class TestRunner {
	public static void main(String[] args) {
		if (args.length != 2) {
			throw new InvalidParameterException("Expected 2 arguments.");
		}
		// Disable output pipes.
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
		// Setup arguments
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
