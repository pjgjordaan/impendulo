package runner;

import gov.nasa.jpf.Config;
import gov.nasa.jpf.JPF;
import gov.nasa.jpf.JPFConfigException;
import gov.nasa.jpf.JPFException;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.PrintStream;

public class JPFRunner {

	/**
	 * @param args
	 * @throws FileNotFoundException
	 */
	public static void main(String[] args) throws FileNotFoundException {
		if (args.length != 3) {
			throw new IllegalArgumentException("Expected 3 arguments.");
		}
		if (!new File(args[0]).exists()) {
			throw new FileNotFoundException("Could not find config file "
					+ args[0]);
		}
		Exception e = run(createConfig(args[0], args[1], args[2]));
		if (e != null) {
			System.err.println(e.getMessage());
		}
	}

	public static Exception run(Config config) {
		try {
			PrintStream tOut = System.out;
			PrintStream tErr = System.out;
			System.setOut(new PrintStream(new NullOutputStream()));
			System.setErr(new PrintStream(new NullOutputStream()));
			JPF jpf = new JPF(config);
			jpf.run();
			System.setOut(tOut);
			System.setErr(tErr);
			System.out.println(((ImpenduloPublisher) jpf.getReporter()
					.getPublishers().get(0)).getStream().toString());
			return null;
		} catch (JPFConfigException cx) {
			return cx;
		} catch (JPFException jx) {
			return jx;
		}
	}

	public static Config createConfig(String configName, String target,
			String targetLocation) {
		Config config = JPF.createConfig(new String[] { configName });
		config.setProperty("target", target);
		config.setProperty("report.publisher", "xml");
		config.setProperty("report.xml.class", "runner.ImpenduloPublisher");
		config.setProperty("classpath",
				targetLocation + ";" + config.getProperty("classpath"));
		config.setProperty("report.xml.start", "jpf,sut");
		config.setProperty("report.xml.transition", "");
		config.setProperty("report.xml.constraint", "constraint,snapshot");
		config.setProperty("report.xml.property_violation",
				"error,snapshot,trace");
		config.setProperty("report.xml.show_steps", "true");
		config.setProperty("report.xml.show_method", "true");
		config.setProperty("report.xml.show_code", "true");
		config.setProperty("report.xml.finished", "result,statistics");
		return config;
	}

}
