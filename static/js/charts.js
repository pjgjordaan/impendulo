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

function timeChart(chartData) {
    if (chartData === null){
	return;
    }
    var m = [80, 100, 80, 100];
    var w = 900 - m[1] - m[3],
    h = 400 - m[0] - m[2],
    max = 0,
    min = Number.MAX_VALUE;
    
    var names = d3.keys(chartData)
	.filter(function(name){
	    return name !== "x";
	});

    for(var i in names){
	var tmp = d3.max(chartData[names[i]]); 
	max = tmp > max ? tmp : max; 
	tmp = d3.min(chartData[names[i]]); 
	min = tmp < min ? tmp : min;
    }
    
    var x = d3.time.scale()
	.domain([d3.min(chartData["x"]), d3.max(chartData["x"])])
	.range([0, w]);
    
    var y = d3.scale.linear()
	.domain([min, max])
	.range([h, 0]);
    
    var colours = d3.scale.category10() 
        .domain(names); 
    
    var chart = d3.select("#chart")
	.append("svg:svg")
	.attr("width", w + m[1] + m[3])
	.attr("height", h + m[0] + m[2])
	.append("svg:g")
	.attr("transform", "translate(" + m[3] + "," + m[0] + ")");
    
    var xAxis = d3.svg.axis()
	.scale(x)
	.tickSize(-h)
	.tickSubdivide(true);
    
    chart.append("svg:g")
	.attr("class", "x axis")
	.attr("transform", "translate(0," + h + ")")
	.call(xAxis);

    chart.append("text")
        .attr("x", w/2 )
        .attr("y",  h+40)
        .style("text-anchor", "middle")
        .text("Time");
    
    var yAxisLeft = d3.svg.axis()
	.scale(y)
	.ticks(4)
	.orient("left");
    
    chart.append("svg:g")
	.attr("class", "y axis")
	.attr("transform", "translate(-25,0)")
	.call(yAxisLeft);
    
    var line = d3.svg.line()
    	.x(function(d,i) { 
	    return x(new Date(chartData["x"][i])); 
	})
	.y(function(d) { 
	    return y(d); 
	});
    
    var dot = function(name, x,y){
	return {
	    "name": name,
	    "x":x, 
	    "y":y
	}
    };
    
    var colourIndex = 0;
    for(var index in names){
	var chartName = names[index];
	var col = colours(colours.domain()[colourIndex]);
	chart.append("svg:path")
	    .attr("d", line(chartData[chartName]))
	    .style("stroke", function() {
		return col;
	    });
	for(var k = 0; k < chartData[chartName].length; k++){
	    chart.append("svg:a")
		.text("here")
		.attr("xlink:href", "displayresult?currentIndex="+k+"&nextIndex="+(k+1))
		.append("svg:circle")
		.attr("name", chartName)
		.attr("x", chartData["x"][k])
		.attr("y", chartData[chartName][k])
		.attr("fill", col)
		.attr("col", col)
		.attr("cx", function(d, i) {
		    return x(chartData["x"][k]);
		})
		.attr("cy", function(d) {
		    return y(chartData[chartName][k]);
		})
		.attr("r", function(d) {
		    return 4;
		})
		.on("mouseover", function() {
		    d3.select(this)
			.attr("fill", "black");
		    var name = d3.select(this).attr("name");
		    var yVal = d3.select(this).attr("y");
		    var xVal = new Date(+d3.select(this).attr("x"))
			.toLocaleTimeString();
		    var xPos = parseFloat(d3.select(this).attr("cx"));
		    var yPos = parseFloat(d3.select(this).attr("cy"));
		    d3.select("#tooltip")
			.style("left", xPos + "px")
			.style("top", yPos + "px")						
			.select("#title")
			.text(name+": "+yVal);
		    d3.select("#tooltip")
			.select("#x")
			.text(xVal);
		    d3.select("#tooltip").classed("hidden", false);
		})
		.on("mouseout", function() {
		    d3.select("#tooltip").classed("hidden", true);
		    d3.select(this)
			.attr("fill", d3.select(this).attr("col"));
		});
	    
	}
	colourIndex ++;
	
    }
    var legend = chart.append("g")
	.attr("class", "legend")
    	.attr("height", 100)
	.attr("width", 100)
	.attr('transform', 'translate(-20,50)');  
    
    legend.selectAll('rect')
	.data(names)
	.enter()
	.append("rect")
	.attr("x", w+35)
	.attr("y", function(d, i){ 
	    return i *  20;
	})
	.attr("width", 10)
	.attr("height", 10)
	.style("fill", function(d, i) { 
            return colours(colours.domain()[i]);
	});
    
    legend.selectAll('text')
	.data(names)
	.enter()
	.append("text")
	.attr("x", w+50)
	.attr("y", function(d, i){ 
	    return i *  20 + 9;
	})
	.text(function(d) {
            return d;
	});

}
