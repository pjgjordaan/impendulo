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
