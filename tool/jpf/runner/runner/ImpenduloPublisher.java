package runner;

import gov.nasa.jpf.Config;
import gov.nasa.jpf.report.Reporter;
import gov.nasa.jpf.report.XMLPublisher;

import java.io.ByteArrayOutputStream;
import java.io.PrintWriter;

public class ImpenduloPublisher extends XMLPublisher {

	ByteArrayOutputStream stream = new ByteArrayOutputStream();
	public ImpenduloPublisher(Config conf, Reporter reporter) {
		super(conf, reporter);
	}

	protected void openChannel() {
		out = new PrintWriter(stream);
	}
	public ByteArrayOutputStream getStream(){
		return stream;
	}
}
