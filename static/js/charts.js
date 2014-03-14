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
var CURRENT_TIME = -1, NEXT_TIME = -1;
var ACTIVE_SHADE = -0.3;
var ACTIVE_WIDTH = 3;
function showChart(fileName, resultName, chartData, compare, currentTime, nextTime){
    if (resultName === 'Summary'){
	summaryTimeChart(fileName, resultName, chartData, compare, currentTime, nextTime);
    } else{
	timeChart(fileName, resultName, chartData, compare, currentTime, nextTime);
    }
}    

function timeChart(fileName, resultName, chartData, compare, currentTime, nextTime) {
    if (chartData === null){
	return;
    }
    CURRENT_TIME = currentTime;
    NEXT_TIME = nextTime;
    var inactiveTools = chartData.filter(function(d){
	return !active(d) && !isLaunch(d);
    });
    var activeTools = chartData.filter(function(d){
	return active(d) && !isLaunch(d);
    });
    var allTools = chartData.filter(function(d){
	return !isLaunch(d);
    });
    var launches = chartData.filter(isLaunch);
    var lineData = d3.nest()
	.key(getKey)
	.entries(allTools);  
    var activeLineData = d3.nest()
	.key(getKey)
	.entries(activeTools);  
    var m = [10, 150, 50, 100];
    var w = 1100 - m[1] - m[3];
    var h = 300 - m[0] - m[2];
    var mid = (d3.max(allTools, getY)-d3.min(allTools, getY))/2;
    var unique = getUnique(chartData);
    var names = unique
	.map(function(d){
	    return d.key;
	});
    var getColour = d3.scale.ordinal()
	.range(COLOURS) 
        .domain(names);  
    var chartColour = function(d) { 
	return getColour(getKey(d)); 
    };    
    var activeColour = function(d){
	return shadeColor(chartColour(d), ACTIVE_SHADE);
    }
    var y = d3.scale.linear()
	.domain(d3.extent(chartData, getY))
	.range([h, 0]);  

    var x = d3.scale.linear()
	.domain(d3.extent(chartData, getX))
	.range([0, w]);  

    var loadLink = function(d) {
	var href = 'displayresult?time='+d.time;
	href = compare ? href + compare : href;
	return href;
    };

    var loadX = function(d,i) { 
	return x(getX(d)); 
    };

    var loadY = function(d) {
	return y(d.y);
    }
    
    var hideTooltip = function(d) {
	var selected = d3.select(this);
	var attr = 'r';
	var val = DOT_RADIUS;
	if(isLaunch(d) || active(d)){
	    var xPos = parseFloat(d3.select(this).attr('cx'));
	    var yPos = parseFloat(d3.select(this).attr('cy'));
	    attr = 'points';
	    if(active(d)){
		val = function(d){
		    return star(xPos, yPos, 2);
		}
	    }else{
		val = function(d){
		    return star(xPos, yPos);
		};
	    }
	}
	selected
	    .transition()
            .duration(500)
            .ease('linear')
	    .attr('fill', chartColour)
	    .attr(attr, val);
	d3.select('#chart-tooltip')
	    .transition()
            .duration(500)
            .ease('linear')
	    .style('opacity', 0);
    };

    var line = d3.svg.line()
	.interpolate('linear')
    	.x(loadX)
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
	    chart.selectAll('.line-active')
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
		.attr('cx', loadX)
		.attr('cy', loadY)
		.attr('r', DOT_RADIUS);
	    chartBody.selectAll('.dot-active')
		.transition()
		.duration(duration)
		.ease(ease)
		.attr('cx', loadX)
		.attr('cy', loadY)
		.attr('points', function(d){
		    return star(loadX(d), loadY(d), 2);
		});
	    chartBody.selectAll('.launch')
		.transition()
		.duration(duration)
		.ease(ease)
		.attr('points', function(d){
		    return star(loadX(d), y(mid));
		})
    		.attr('cx', loadX)
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
        .text('Running Time (s)');
    
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

    chartBody.selectAll('.line-active')
	.data(activeLineData, getKey)
	.enter()
	.append('path')
	.attr('class', 'line-active')
	.attr('key', chartKey)
	.style('stroke', activeColour)
	.style("stroke-width", ACTIVE_WIDTH)
	.attr('d', function(d) { 
	    return line(d.values); 
	})
	.style('opacity', function(d){
	    return d.values[0].show ? 1.0 : 0.0;
	});

    chartBody.selectAll('.link')
	.data(inactiveTools)
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
    	.attr('cx', loadX)
	.attr('cy', loadY)
	.attr('r', DOT_RADIUS)
	.on('mouseover', showTooltip)
	.on('mouseout', hideTooltip);
    
    chartBody.selectAll('.dot-active')
	.data(activeTools)
	.enter()
	.append('svg:polygon')
	.attr('class', 'dot-active')
	.style('opacity', function(d){
	    return d.show ? 1.0 : 0.0;
	})
	.attr('fill', activeColour)
    	.attr('cx', loadX)
	.attr('cy', loadY)
	.attr('points', function(d){
	    return star(loadX(d), loadY(d), 2);
	})
    	.on('mouseover', showTooltip)
	.on('mouseout', hideTooltip);

    chartBody.selectAll('.launch')
	.data(launches)
	.enter()
	.append('svg:polygon')
	.attr('class', 'launch')
	.attr('fill', chartColour)
    	.attr('points', function(d){
	    return star(loadX(d), y(mid));
	})
    	.attr('cx', loadX)
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
	.entries(unique);
    
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

function summaryTimeChart(fileName, resultName, chartData, compare, currentTime, nextTime) {
    if (chartData === null){
	return;
    }
    CURRENT_TIME = currentTime;
    NEXT_TIME = nextTime;
    var inactiveTools = chartData.filter(function(d){
	return !active(d) && !isLaunch(d);
    });
    var activeTools = chartData.filter(function(d){
	return active(d) && !isLaunch(d);
    });
    var allTools = chartData.filter(function(d){
	return !isLaunch(d);
    });

    var launches = chartData.filter(isLaunch);
    var lineData = d3.nest()
	.key(getKey)
	.entries(allTools);  
    var activeLineData = d3.nest()
	.key(getKey)
	.entries(activeTools);  
    var m = [10, 150, 10, 100];
    var w = 1100 - m[1] - m[3];
    var h = 200 - m[0] - m[2];
    var mid = (d3.max(allTools, getY)-d3.min(allTools, getY))/2;
    var unique = getUnique(chartData);
    var names = unique
	.map(function(d){
	    return d.key;
	});
    var getColour = d3.scale.ordinal()
	.range(COLOURS) 
        .domain(names);  
    var chartColour = function(d) { 
	return getColour(getKey(d)); 
    };
    var activeColour = function(d){
	return shadeColor(chartColour(d), ACTIVE_SHADE);
    }
    var scales = d3.map();
    for(var i = 0; i < names.length; i ++){
	var vals = chartData.filter(function(d){
	    return d.key === names[i];
	});
	scales[names[i]] = d3.scale.linear()
	    .domain(d3.extent(vals, getY))
	    .range([h, 0]);
    }
    var x = d3.linear.scale()
	.domain(d3.extent(chartData, getX))
	.range([0, w]);  

    var loadLink = function(d) {
	var href = 'displayresult?time='+d.time;
	href = compare ? href + compare : href;
	return href;
    };

    var loadX = function(d,i) { 
	return x(getX(d)); 
    };

    var loadY = function(d) {
	return scales[d.key](d.y);
    }
    
    var hideTooltip = function(d) {
	var selected = d3.select(this);
	var attr = 'r';
	var val = DOT_RADIUS;
	if(isLaunch(d) || active(d)){
	    var xPos = parseFloat(d3.select(this).attr('cx'));
	    var yPos = parseFloat(d3.select(this).attr('cy'));
	    attr = 'points';
	    if(active(d)){
		val = function(d){
		    return star(xPos, yPos, 2);
		}
	    }else{
		val = function(d){
		    return star(xPos, yPos);
		};
	    }
	}
	selected
	    .transition()
            .duration(500)
            .ease('linear')
	    .attr('fill', chartColour)
	    .attr(attr, val);
	d3.select('#chart-tooltip')
	    .transition()
            .duration(500)
            .ease('linear')
	    .style('opacity', 0);
    };

    var line = d3.svg.line()
	.interpolate('linear')
    	.x(loadX)
	.y(loadY);
    
    var xAxis = d3.svg.axis()
	.scale(x)
	.ticks(7)
	.tickSize(-h)
	.orient('bottom')
	.tickSubdivide(true);
       
    var zoom = d3.behavior.zoom()
	.x(x)
	.on('zoom', function(){
	    var duration = 1000;
	    var ease = 'linear';
	    chart.select('.x.axis')
		.transition()
		.duration(duration)
		.ease(ease)
		.call(xAxis);
	    chart.selectAll('.line')
		.transition()
		.duration(duration)
	    	.ease(ease)
		.attr('d', function(d) { 
		    return line(d.values); 
		});
	    chart.selectAll('.line-active')
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
		.attr('cx', loadX)
		.attr('cy', loadY)
		.attr('r', DOT_RADIUS);
	    chartBody.selectAll('.dot-active')
		.transition()
		.duration(duration)
		.ease(ease)
		.attr('cx', loadX)
		.attr('cy', loadY)
		.attr('points', function(d){
		    return star(loadX(d), loadY(d), 2);
		});
	    chartBody.selectAll('.launch')
		.transition()
		.duration(duration)
		.ease(ease)
		.attr('points', function(d){
		    return star(loadX(d), scales[d.key](mid));
		})
    		.attr('cx', loadX)
		.attr('cy', function(d){
		    return scales[d.key](mid);
		});
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
        .text('Running Time (s)');
       
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

    chartBody.selectAll('.line-active')
	.data(activeLineData, getKey)
	.enter()
	.append('path')
	.attr('class', 'line-active')
	.attr('key', chartKey)
	.style('stroke', activeColour)
	.style("stroke-width", ACTIVE_WIDTH)
	.attr('d', function(d) { 
	    return line(d.values); 
	})
	.style('opacity', function(d){
	    return d.values[0].show ? 1.0 : 0.0;
	});

    chartBody.selectAll('.link')
	.data(inactiveTools)
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
    	.attr('cx', loadX)
	.attr('cy', loadY)
	.attr('r', DOT_RADIUS)
	.on('mouseover', showTooltip)
	.on('mouseout', hideTooltip);

    chartBody.selectAll('.dot-active')
	.data(activeTools)
	.enter()
	.append('svg:polygon')
	.attr('class', 'dot-active')
	.style('opacity', function(d){
	    return d.show ? 1.0 : 0.0;
	})
	.attr('fill', function(d){
	    return shadeColor(chartColour(d), ACTIVE_SHADE);
	})
    	.attr('cx', loadX)
	.attr('cy', loadY)
	.attr('points', function(d){
	    return star(loadX(d), loadY(d), 2);
	})
    	.on('mouseover', showTooltip)
	.on('mouseout', hideTooltip);

    chartBody.selectAll('.launch')
	.data(launches)
	.enter()
	.append('svg:polygon')
	.attr('class', 'launch')
	.attr('fill', chartColour)
    	.attr('points', function(d){
	    return star(loadX(d), scales[d.key](mid));
	})
    	.attr('cx', loadX)
	.attr('cy', function(d){
	    return scales[d.key](mid);
	})
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
	.entries(unique);
    
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
    var xVal = getX(d);
    var xPos = parseFloat(d3.select(this).attr('cx'));
    var yPos = parseFloat(d3.select(this).attr('cy'));
    var text = d.name+': '+d.y;
    var selected = d3.select(this)
    var attr = 'r';
    var val = 8;
    if(active(d)){
	attr = 'points';
	val = function(d){
	    return star(xPos, yPos, 3);
	}
    }
    if(isLaunch(d)){
	attr = 'points';
	text = d.name;
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
    
    var tooltip = d3.select('#chart-tooltip')
	.style('left', xPos + 'px')
	.style('top', (yPos+30) + 'px');	
    tooltip.selectAll('.chart-tooltip-line')
	.remove();
    tooltip.selectAll('pre')
	.remove();
    tooltip
	.append('h5')
	.attr('class', 'chart-tooltip-line')
	.text(d.user);
    tooltip
	.append('p')
	.attr('class', 'chart-tooltip-line')
	.text(text + ' Time: ' + xVal+ 's');
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

function getX(d){
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
    
function invert(hex) {
    if (hex.length != 7 || hex.indexOf('#') != 0) {
	return null;
    }
    return "#" + pad((255 - parseInt(hex.substring(1, 3), 16)).toString(16)) + pad((255 - parseInt(hex.substring(3, 5), 16)).toString(16)) + pad((255 - parseInt(hex.substring(5, 7), 16)).toString(16));
}
    
function pad(num) {
    if (num.length < 2) {
	return "0" + num;
    } else {
	return num;
    }
}

function shape(d){
    if(active(d)){
	return 'svg:polygon';
    }else{
	return 'svg:circle';
    }
}

function active(d){
    if (CURRENT_TIME === -1 || NEXT_TIME === -1){
	return getX(d) === CURRENT_TIME || getX(d) === NEXT_TIME;
    } else{
	return getX(d) >= CURRENT_TIME && getX(d) <= NEXT_TIME;
    }
}

function shadeColor(color, percent) {   
    var f=parseInt(color.slice(1),16),t=percent<0?0:255,p=percent<0?percent*-1:percent,R=f>>16,G=f>>8&0x00FF,B=f&0x0000FF;
    return "#"+(0x1000000+(Math.round((t-R)*p)+R)*0x10000+(Math.round((t-G)*p)+G)*0x100+(Math.round((t-B)*p)+B)).toString(16).slice(1);
}
