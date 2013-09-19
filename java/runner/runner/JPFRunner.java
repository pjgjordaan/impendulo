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

package runner;

import gov.nasa.jpf.Config;
import gov.nasa.jpf.JPF;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.OutputStream;
import java.io.PrintStream;
import java.util.HashMap;
import java.util.Map;
import java.util.Map.Entry;

/**
 * This class is used to run JPF on a provided Java file. See
 * http://godoc.org/github.com/godfried/impendulo/tool/jpf#Tool for more
 * information.
 * 
 * @author godfried
 * 
 */
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
		try {
			run(createConfig(args[0], args[1], args[2], args[3]));
		} catch (Exception e) {
			System.err.println(e.getMessage());
		}
	}

	/**
	 * Here we run JPF with a provided Config.
	 * 
	 * @param config
	 *            the JPF configuration to be used for this execution.
	 */
	public static void run(Config config) throws Exception {
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
		} finally {
			System.setOut(out);
			System.setErr(err);
		}
	}

	/**
	 * This method creates a Config from a provided config file.
	 * 
	 * @param configName
	 *            the config file to load.
	 * @param target
	 *            the target of this config.
	 * @param targetLocation
	 *            the location of the target.
	 * @param outFile
	 *            the output file to be published to.
	 * @return A new Config.
	 */
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

	/**
	 * Loads a default configuration
	 * 
	 * @return
	 */
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
