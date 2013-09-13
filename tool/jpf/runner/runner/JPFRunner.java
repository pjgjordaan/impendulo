package runner;

import gov.nasa.jpf.Config;
import gov.nasa.jpf.JPF;
import gov.nasa.jpf.JPFConfigException;
import gov.nasa.jpf.JPFException;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.OutputStream;
import java.io.PrintStream;
import java.util.HashMap;
import java.util.Map;
import java.util.Map.Entry;

public class JPFRunner {

	/**
	 * @param args
	 * @throws FileNotFoundException
	 */
	public static void main(String[] args) throws FileNotFoundException {
		if (args.length != 4) {
			throw new IllegalArgumentException("Expected 3 arguments.");
		}
		if (!new File(args[0]).exists()) {
			throw new FileNotFoundException("Could not find config file "
					+ args[0]);
		}
		Exception e = run(createConfig(args[0], args[1], args[2], args[3]));
		if (e != null) {
			System.err.println(e.getMessage());
		}
	}

	public static Exception run(Config config) {
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
		try {

			JPF jpf = new JPF(config);
			jpf.run();
			return null;
		} catch (JPFConfigException cx) {
			return cx;
		} catch (JPFException jx) {
			return jx;
		} finally {
			System.setOut(out);
			System.setErr(err);
		}
	}

	public static Config createConfig(String configName, String target,
			String targetLocation, String outFile) {
		Config config = JPF.createConfig(new String[] { configName });
		config.setProperty("target", target);
		config.setProperty("report.xml.file", outFile);
		config.setProperty("classpath",
				targetLocation + ";" + config.getProperty("classpath"));
		Map<String, String> defualt = DefaultConfig();
		for (Entry<String, String> cfg : defualt.entrySet()) {
			if (!config.containsKey(cfg.getKey())) {
				config.put(cfg.getKey(), cfg.getValue());
			}
		}
		return config;
	}

	public static Map<String, String> DefaultConfig() {
		Map<String, String> ret = new HashMap<String, String>();
		ret.put("report.publisher", "xml");
		ret.put("report.xml.class", "runner.ImpenduloPublisher");
		ret.put("report.xml.start", "jpf,sut");
		ret.put("report.xml.transition", "");
		ret.put("report.xml.constraint", "constraint,snapshot");
		ret.put("report.xml.property_violation", "error,snapshot");
		ret.put("report.xml.show_steps", "true");
		ret.put("report.xml.show_method", "true");
		ret.put("report.xml.show_code", "true");
		ret.put("report.xml.finished", "result,statistics");
		ret.put("search.class", "gov.nasa.jpf.search.DFSearch");
		ret.put("search.depth_limit", "1000");
		return ret;
	}
}
