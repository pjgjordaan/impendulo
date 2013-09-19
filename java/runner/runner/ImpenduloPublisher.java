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
import gov.nasa.jpf.report.Reporter;
import gov.nasa.jpf.report.Statistics;
import gov.nasa.jpf.report.XMLPublisher;

/**
 * This class extends XMLPublisher so that we can easily customise the XML
 * output generated. See
 * http://godoc.org/github.com/godfried/impendulo/tool/jpf#Report for more
 * information.
 * 
 * @author godfried
 * 
 */
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
