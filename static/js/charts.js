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

function timeChart(fileName, resultName, chartData) {
    if (chartData === null){
	return;
    }
    var lineData = d3.nest()
	.key(function(d) { return d.name; })
	.entries(chartData);
    
    var m = [80, 100, 80, 100];
    var w = 900 - m[1] - m[3];
    var h = 400 - m[0] - m[2];

    var colour = d3.scale.category10() 
        .domain(d3.keys(chartData[0]).filter(function(key) { return key == "name"; }));  

    var dotCol = function(d) { 
	return colour(d.name); 
    };    
    
    var lineCol = function(d) { 
        return colour(d.key);
    }
    var y = d3.scale.linear()
	.domain(d3.extent(chartData, function(d){return d.y;}))
	.range([h, 0]);
    
    var x = d3.time.scale()
	.domain(d3.extent(chartData, function(d){return d.x;}))
	.range([0, w]);  

    var loadLink = function(d) {
	return "displayresult?time="+d.x+
	    "&resultname="+resultName+
	    "&filename="+fileName;
    };

    var loadDate = function(d,i) { 
	return x(new Date(d.x)); 
    };

    var loadY = function(d) {
	return y(d.y);
    }
    
    var hideTooltip = function(d) {
	d3.select("#tooltip").classed("hidden", true);
	d3.select(this)
	    .attr("fill", dotCol);
    };

    var line = d3.svg.line()
	.interpolate("linear")
    	.x(loadDate)
	.y(loadY);
    
    var xAxis = d3.svg.axis()
	.scale(x)
	.ticks(7)
	.tickSize(-h)
	.orient("bottom")
	.tickSubdivide(true);
    
    var yAxis = d3.svg.axis()
	.scale(y)
	.ticks(5)
	.orient("left");   
    
    var zoom = d3.behavior.zoom()
	.x(x)
	.y(y)
	.on("zoom", function(){
	    chart.select(".x.axis").call(xAxis);
	    chart.select(".y.axis").call(yAxis);
	    chart.selectAll(".line")
		.attr("class", "line")
		.attr("d", function(d) { return line(d.values); })
		.style("stroke", lineCol);
	    chartBody.selectAll(".link")
	    	.attr("xlink:href", loadLink)
		.attr("class", "link")
		.select(".dot")
		.attr("class", "dot")
		.attr("fill", dotCol)
    		.attr("cx", loadDate)
		.attr("cy", loadY)
		.attr("r", 4)
		.on("mouseover", showTooltip)
		.on("mouseout", hideTooltip);
	});
    
    var chart = d3.select("#chart")
	.append("svg:svg")
	.attr("width", w + m[1] + m[3])
	.attr("height", h + m[0] + m[2])
	.append("svg:g")
	.attr("transform", "translate(" + m[3] + "," + m[0] + ")")
	.call(zoom);

    chart.append("svg:rect")
	.attr("width", w)
	.attr("height", h)
	.attr("class", "plot");
    
    chart.append("svg:g")
	.attr("class", "x axis")
	.attr("transform", "translate(0," + h + ")")
	.call(xAxis);

    chart.append("text")
        .attr("x", w/2 )
        .attr("y",  h+40)
        .style("text-anchor", "middle")
        .text("Time");
    
    chart.append("svg:g")
	.attr("class", "y axis")
	.attr("transform", "translate(-25,0)")
	.call(yAxis);
    
    chart.append("svg:clipPath")
	.attr("id", "clip")
	.append("svg:rect")
	.attr("x", -10)
	.attr("y", -10)
	.attr("width", w+20)
	.attr("height", h+20);

    var chartBody = chart.append("g")
	.attr("clip-path", "url(#clip)");

    chartBody.selectAll(".line")
	.data(lineData, function(d) { return d.key; })
	.enter().append("path")
	.attr("class", "line")
	.attr("d", function(d) { return line(d.values); })
	.style("stroke", lineCol);
    
    chartBody.selectAll(".link")
	.data(chartData)
	.enter()
	.append("svg:a")
	.attr("xlink:href", loadLink)
	.attr("class", "link")
	.append("svg:circle")
	.attr("class", "dot")
	.attr("fill", dotCol)
    	.attr("cx", loadDate)
	.attr("cy", loadY)
	.attr("r", 4)
	.on("mouseover", showTooltip)
	.on("mouseout", hideTooltip);

    var legend = chart.append("g")
	.attr("class", "legend")
    	.attr("height", 100)
	.attr("width", 100)
	.attr('transform', 'translate(-20,50)');  
    
    legend.selectAll('rect')
	.data(lineData)
	.enter()
	.append("rect")
	.attr("x", w+35)
	.attr("y", function(d, i){ 
	    return i *  20;
	})
	.attr("width", 10)
	.attr("height", 10)
	.style("fill", lineCol);
    
    legend.selectAll('text')
	.data(lineData)
	.enter()
	.append("text")
	.attr("x", w+50)
	.attr("y", function(d, i){ 
	    return i *  20 + 9;
	})
	.text(function(d) {
            return d.key;
	});

}

function showTooltip(d){
    d3.select(this)
	.attr("fill", "black");
    var xVal = new Date(+d.x).toLocaleTimeString();
    var xPos = parseFloat(d3.select(this).attr("cx"));
    var yPos = parseFloat(d3.select(this).attr("cy"));
    d3.select("#tooltip")
	.style("left", xPos + "px")
	.style("top", yPos + "px")						
	.select("#title")
	.text(d.name+": "+d.y);
    d3.select("#tooltip")
	.select("#x")
	.text(xVal);
    d3.select("#tooltip").classed("hidden", false);
}
