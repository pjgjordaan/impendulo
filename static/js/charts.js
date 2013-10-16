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

var DOT_RADIUS = 4;
var FOCUS_COLOUR = 'black';
var COLOURS = ['#1f77b4', '#ff7f0e', '#2ca02c',  '#d62728', '#9467bd', '#8c564b', '#e377c2', '#7f7f7f', '#bcbd22', '#17becf'];
var LAUNCH = 'Launches';
      
function timeChart(fileName, resultName, chartData, compare) {
    if (chartData === null){
	return;
    }
    var tools = chartData.filter(function(d){
	return !isLaunch(d);
    });
    var launches = chartData.filter(isLaunch);
    var lineData = d3.nest()
	.key(getKey)
	.entries(tools);  
    var m = [10, 150, 100, 100];
    var w = 1100 - m[1] - m[3];
    var h = 500 - m[0] - m[2];
    var mid = (d3.max(tools, getY)-d3.min(tools, getY))/2;
    var getColour = d3.scale.ordinal()
	.range(COLOURS) 
        .domain(d3.keys(chartData[0]).filter(function(key) { return key === 'name'; }));  
    var chartColour = function(d) { 
	return getColour(getKey(d)); 
    };    
    var y = d3.scale.linear()
	.domain(d3.extent(chartData, getY))
	.range([h, 0]);
    
    var x = d3.time.scale()
	.domain(d3.extent(chartData, getChartX))
	.range([0, w]);  

    var loadLink = function(d) {
	var href = 'displayresult?time='+d.x+
	    '&result='+resultName+
	    '&file='+fileName+
	    '&sid='+d.subId+
	    '&uid='+d.user;
	href = compare ? href + compare : href;
	return href;
    };

    var loadDate = function(d,i) { 
	return x(new Date(getChartX(d))); 
    };

    var loadY = function(d) {
	return y(d.y);
    }
    
    var hideTooltip = function(d) {
	var selected = d3.select(this);
	var attr = 'r';
	var val = DOT_RADIUS;
	if(isLaunch(d)){
	    var xPos = parseFloat(d3.select(this).attr('cx'));
	    var yPos = parseFloat(d3.select(this).attr('cy'));
	    attr = 'points';
	    val = function(d){
		return star(xPos, yPos);
	    };
	}
	selected
	    .transition()
            .duration(500)
            .ease('linear')
	    .attr('fill', chartColour)
	    .attr(attr, val);
	d3.select('#tooltip')
	    .transition()
            .duration(500)
            .ease('linear')
	    .style('opacity', 0);
    };

    var line = d3.svg.line()
	.interpolate('linear')
    	.x(loadDate)
	.y(loadY);
    
    var xAxis = d3.svg.axis()
	.scale(x)
	.ticks(7)
	.tickSize(-h)
	.orient('bottom')
	.tickSubdivide(true);
    
    var yAxis = d3.svg.axis()
	.scale(y)
	.ticks(5)
	.orient('left')
	.tickSubdivide(true);
    
    var zoom = d3.behavior.zoom()
	.x(x)
	.y(y)
	.on('zoom', function(){
	    var duration = 1000;
	    var ease = 'linear';
	    chart.select('.x.axis')
		.transition()
		.duration(duration)
		.ease(ease)
		.call(xAxis);
	    chart.select('.y.axis')
		.transition()
		.duration(duration)
		.ease(ease)
		.call(yAxis);
	    chart.selectAll('.line')
		.transition()
		.duration(duration)
	    	.ease(ease)
		.attr('d', function(d) { 
		    return line(d.values); 
		});
	    chartBody.selectAll('.dot')
		.transition()
		.duration(duration)
		.ease(ease)
		.attr('cx', loadDate)
		.attr('cy', loadY)
		.attr('r', DOT_RADIUS);
	    chartBody.selectAll('.launch')
		.transition()
		.duration(duration)
		.ease(ease)
		.attr('points', function(d){
		    return star(loadDate(d), y(mid));
		})
    		.attr('cx', loadDate)
		.attr('cy', y(mid));
	});
    
    var chart = d3.select('#chart')
	.append('svg:svg')
	.attr('width', w + m[1] + m[3])
	.attr('height', h + m[0] + m[2])
	.append('svg:g')
	.attr('transform', 'translate(' + m[3] + ',' + m[0] + ')')
	.call(zoom);

    chart.append('svg:rect')
	.attr('width', w)
	.attr('height', h)
	.attr('class', 'plot');
    
    chart.append('svg:g')
	.attr('class', 'x axis')
	.attr('transform', 'translate(0,' + h + ')')
	.call(xAxis);

    chart.append('text')
        .attr('x', w/2 )
        .attr('y',  h+40)
        .style('text-anchor', 'middle')
        .text('Time');
    
    chart.append('svg:g')
	.attr('class', 'y axis')
	.attr('transform', 'translate(-25,0)')
	.call(yAxis);
    
    chart.append('svg:clipPath')
	.attr('id', 'clip')
	.append('svg:rect')
	.attr('x', -10)
	.attr('y', -10)
	.attr('width', w+20)
	.attr('height', h+20);

    var chartBody = chart.append('g')
	.attr('clip-path', 'url(#clip)');

    chartBody.selectAll('.line')
	.data(lineData, getKey)
	.enter()
	.append('path')
	.attr('class', 'line')
	.attr('key', chartKey)
	.style('stroke', chartColour)
	.attr('d', function(d) { 
	    return line(d.values); 
	})
	.style('opacity', function(d){
	    return d.values[0].show ? 1.0 : 0.0;
	});

    chartBody.selectAll('.link')
	.data(tools)
	.enter()
	.append('svg:a')
	.attr('xlink:href', loadLink)
	.attr('class', 'link')
	.attr('key', chartKey)
	.style('opacity', function(d){
	    return d.show ? 1.0 : 0.0;
	})
	.append('svg:circle')
	.attr('class', 'dot')
	.attr('fill', chartColour)
    	.attr('cx', loadDate)
	.attr('cy', loadY)
	.attr('r', DOT_RADIUS)
	.on('mouseover', showTooltip)
	.on('mouseout', hideTooltip);


    chartBody.selectAll('.launch')
	.data(launches)
	.enter()
	.append('svg:polygon')
	.attr('class', 'launch')
	.attr('fill', chartColour)
    	.attr('points', function(d){
	    return star(loadDate(d), y(mid));
	})
    	.attr('cx', loadDate)
	.attr('cy', y(mid))
	.attr('key', chartKey)
	.style('opacity', function(d){
	    return 0.0;
	})
	.on('mouseover', showTooltip)
	.on('mouseout', hideTooltip);

    var legendData = d3.nest()
	.key(function(d){
	    return d.subId;
	})
	.entries(getUnique(chartData));
    
    var legend = chart.append('g')
	.attr('class', 'legend')
    	.attr('height', 100)
	.attr('width', 100)
	.attr('transform', 'translate(-20,50)');  
    
    var legendElements = legend.selectAll('g')
	.data(legendData)
	.enter()
	.append('g');
    legendElements.append('text')
	.attr('x', w+20)
	.attr('y', function(d, i){
	    var offset = i * (d.values.length + 1) * 25;
	    return (offset) + 13;
	})
	.attr('font-size','10px')
	.attr('class', 'clickable')
	.text(function(d){
	    return d.values[0].user;
	})
	.on('click', function(d, i){
	    for(var j = 0; j < d.values.length; j++){
		toggleVisibility(d.values[j]);
	    }
	    var opacity = d3.select(this).style('opacity');
	    opacity = opacity === '1' ? 0.3 : 1.0;
	    d3.select(this)
		.transition()
		.duration(500)
		.ease('linear')
		.style('opacity', opacity);
	});
    
    var elemData = function(d, i){
	var offset = (i * (d.values.length +1) * 25) + 25;
	return legendData[i].values.map(function(d){
	    d.offset = offset;
	    return d;
	});
    };

    legendElements.selectAll('.legendrect')
	.data(elemData)
	.enter()
	.append('rect')
	.attr('class', 'legendrect clickable')
	.attr('key', legendKey)
	.attr('x', w+30)
	.attr('y', function(d, i){ 
	    return i *  25 + d.offset;
	})
	.attr('width', 15)
	.attr('height', 15)
	.attr('showing', true)
	.style('fill', chartColour)
	.style('opacity', function(d){
	    return d.show ? 1.0 : 0.3;
	})
	.on('click', toggleVisibility);
    
    legendElements.selectAll('.legendtext')
	.data(elemData)
	.enter()
	.append('text')
	.attr('class', 'legendtext')
	.attr('key', legendKey)
	.attr('x', w+50)
	.attr('y', function(d, i){ 
	    return (i *  25) + 13 + d.offset;
	})
	.style('opacity', function(d){
	    return d.show ? 1.0 : 0.3;
	})
	.attr('font-size','10px')
	.text(getName);

}

function legendKey(d){
    return 'legend'+trimKey(d);
}

function chartKey(d){
    return 'chart'+trimKey(d);
}

function getUnique(arr){
   var u = {}, a = [];
   for(var i = 0, l = arr.length; i < l; ++i){
      if(u.hasOwnProperty(arr[i].key)) {
         continue;
      }
      a.push(arr[i]);
      u[arr[i].key] = 1;
   }
   return a;
}

function showTooltip(d){
    var xVal = new Date(getChartX(d)).toLocaleTimeString();
    var xPos = parseFloat(d3.select(this).attr('cx'));
    var yPos = parseFloat(d3.select(this).attr('cy'));
    var text = d.name+': '+d.y;
    var selected = d3.select(this)
    var attr = 'r';
    var val = 8;
    if(isLaunch(d)){
	text = d.name;
	attr = 'points';
	val = function(d){
	    return star(xPos, yPos, 2);
	};
    } 
    selected
	.transition()
        .duration(500)
        .ease('linear')
	.attr('fill', FOCUS_COLOUR)
	.attr(attr, val);
    
    var tooltip = d3.select('#tooltip')
	.style('left', xPos + 'px')
	.style('top', yPos + 'px');	
    tooltip.selectAll('.tooltip-line')
	.remove();
    tooltip
	.append('h5')
	.attr('class', 'tooltip-line')
	.text(d.user);
    tooltip
	.append('p')
	.attr('class', 'tooltip-line')
	.text(text);
    if(d.adjust > 0){
	var adjustment = ' (+'+new Date(d.adjust).toLocaleTimeString()+')'
	tooltip
	    .append('p')
	    .attr('class', 'tooltip-line')
	    .text(xVal + adjustment)
    } else{
	tooltip
	    .append('p')
	    .attr('class', 'tooltip-line')
	    .text(xVal);
    }
    tooltip
	.transition()
        .duration(500)	
	.ease('linear')
	.style('opacity', 1);
}

function star(x, y, scale)
{
    scale = scale || 1;
    var innerRadius = 2 * scale;
    var outerRadius = 10 * scale;
    var arms = 8;
    var results = '';  
    var angle = Math.PI / arms;
    for (var i = 0; i < 2 * arms; i++)
    {
	var r = (i & 1) === 0 ? outerRadius : innerRadius;
	var currX = x + Math.cos(i * angle) * r;
	var currY = y + Math.sin(i * angle) * r;
	if (i === 0)
	{
            results = currX + ',' + currY;
	}
	else
	{
            results += ', ' + currX + ',' + currY;
	}
    }
    return results;
}

function toggleVisibility(d){
    var key = trimKey(d);
    var chartOpacity = d3.select('[key=chart'+key+']').style('opacity');
    var legendOpacity = 1.0;
    if(chartOpacity === '0'){
	chartOpacity = 1.0;
    }else{
	chartOpacity = 0.0;
	legendOpacity = 0.3;
    }
    d3.selectAll('[key=legend'+key+']')
	.transition()
        .duration(500)
        .ease('linear')
	.style('opacity', legendOpacity);
    d3.selectAll('[key=chart'+key+']')
	.transition()
        .duration(500)
        .ease('linear')
	.style('opacity', chartOpacity);
}

function getChartX(d){
    return +(d.x + d.adjust);
}

function getActualX(d){
    return +d.x;
}

function getY(d){
    return +d.y;
}

function getName(d){
    return d.name;
}

function isLaunch(d){
    return endsWith(d.name, LAUNCH);
}

function getKey(d) {
    return d.key; 
}

function trimKey(d) {
    return replaceAll(getKey(d), ' ', '');
}

function replaceAll(str, find, replace) {
  return str.replace(new RegExp(find, 'g'), replace);
}

function endsWith(str, suffix) {
    return str.indexOf(suffix, str.length - suffix.length) !== -1;
};
