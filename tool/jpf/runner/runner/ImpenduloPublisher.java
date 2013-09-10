package runner;

import gov.nasa.jpf.Config;
import gov.nasa.jpf.report.Reporter;
import gov.nasa.jpf.report.Statistics;
import gov.nasa.jpf.report.XMLPublisher;

public class ImpenduloPublisher extends XMLPublisher {

	public ImpenduloPublisher(Config conf, Reporter reporter) {
		super(conf, reporter);
	}

	@Override
	protected void publishStatistics() {
		Statistics stat = reporter.getStatistics();
		out.println("  <statistics>");
		out.println("    <elapsed-time>"
				+ String.valueOf(reporter.getElapsedTime()) + "</elapsed-time>");
		out.println("    <new-states>" + stat.newStates + "</new-states>");
		out.println("    <visited-states>" + stat.visitedStates
				+ "</visited-states>");
		out.println("    <backtracked-states>" + stat.backtracked
				+ "</backtracked-states>");
		out.println("    <end-states>" + stat.endStates + "</end-states>");
		out.println("    <max-memory unit=\"MB\">" + (stat.maxUsed >> 20)
				+ "</max-memory>");
		out.println("  </statistics>");
	}

}
