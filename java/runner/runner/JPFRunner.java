//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package runner;

import gov.nasa.jpf.Config;
import gov.nasa.jpf.JPF;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.OutputStream;
import java.io.PrintStream;
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
		Config config = new Config(new String[] {});
		config.setProperty("target", target);
		config.setProperty("report.xml.file", outFile);
		config.setProperty("classpath",
				targetLocation + ";" + config.getProperty("classpath"));
		addDefaults(config);
		addFileConfig(config, configName);
		return config;
	}

	private static void addFileConfig(Config config, String configName) {
		Config fileConfig = new Config(configName);
		for(Entry<Object, Object> fileProperty : fileConfig.entrySet()){
			config.put(fileProperty.getKey(), fileProperty.getValue());
		}
	}

	/**
	 * Loads a default configuration
	 * 
	 * @return
	 */
	public static void addDefaults(Config config) {
		config.put("report.publisher", "xml");
		config.put("report.xml.class", "runner.ImpenduloPublisher");
		config.put("report.xml.start", "jpf,sut");
		config.put("report.xml.transition", "");
		config.put("report.xml.constraint", "constraint,snapshot");
		config.put("report.xml.property_violation", "error,snapshot");
		config.put("report.xml.show_steps", "true");
		config.put("report.xml.show_method", "true");
		config.put("report.xml.show_code", "true");
		config.put("report.xml.finished", "result,statistics");
		config.put("search.class", "gov.nasa.jpf.search.DFSearch");
		config.put("search.depth_limit", "1000");
		config.put("search.multiple_errors", "true");
	}
}
