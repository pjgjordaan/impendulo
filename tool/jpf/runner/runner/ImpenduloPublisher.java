package runner;

import gov.nasa.jpf.Config;
import gov.nasa.jpf.report.Reporter;
import gov.nasa.jpf.report.XMLPublisher;

import java.io.PrintWriter;

public class ImpenduloPublisher extends XMLPublisher {

	public ImpenduloPublisher(Config conf, Reporter reporter) {
		super(conf, reporter);
	}

	protected void openChannel() {
		out = new PrintWriter(System.out);
	}
}
