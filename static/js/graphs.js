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

function drawGraph(graphArgs){
    if (graphArgs["type"] === "time"){
	timeGraph(graphArgs);
    } else if (graphArgs["type"] == "num"){
	lineGraph(graphArgs);
    }
}

function lineGraph(graphArgs) {
    if (graphArgs === null){
	return;
    }
    var palette = new Rickshaw.Color.Palette();
    for (var i = 0; i < graphArgs['series'].length; i++) 
    {
	graphArgs['series'][i]["color"] =  palette.color();
    }
    graphArgs['element'] = document.getElementById('resultGraph');
    var graph = new Rickshaw.Graph(graphArgs);   
    var yTickFormat = Rickshaw.Fixtures.Number.formatKMBT;
    var yHoverFormat =  function(n) {
	return n === null ? n : n.toFixed(2);
    };
    var xHoverFormat =  function(n) {
	return n === null ? n : n.toFixed(2);
    };
    if (graphArgs["yformat"] != null){
	yTickFormat = function(n) {
	    var map = graphArgs["yformat"];
	    return map[n];
	};
	yHoverFormat = yTickFormat; 
    }
    var y_ticks = new Rickshaw.Graph.Axis.Y( {
	graph: graph,
	orientation: 'left',
	tickFormat: yTickFormat,
	element: document.getElementById('resultGraphY'),
    } );
    var x_ticks = new Rickshaw.Graph.Axis.X( {
	graph: graph,
	orientation: 'bottom',
	element: document.getElementById('resultGraphX'),
	tickFormat: Rickshaw.Fixtures.Number.formatKMBT
    } );
    
    var legend = new Rickshaw.Graph.Legend( {
	element: document.getElementById('resultGraphLegend'),
	graph: graph
    } );
    var hoverDetail = new Rickshaw.Graph.HoverDetail( {
	graph: graph,
	yFormatter: yHoverFormat,
	xFormatter: xHoverFormat
    } );
    var shelving = new Rickshaw.Graph.Behavior.Series.Toggle( {
	graph: graph,
	legend: legend
    } );      
    graph.render();
}

function timeGraph(graphArgs) {
    if (graphArgs === null){
	return;
    }
    var palette = new Rickshaw.Color.Palette();
    for (var i = 0; i < graphArgs['series'].length; i++) 
    {
	graphArgs['series'][i]["color"] =  palette.color();
    }
    graphArgs['element'] = document.getElementById('resultGraph');
    var graph = new Rickshaw.Graph(graphArgs);   
    var yTickFormat = Rickshaw.Fixtures.Number.formatKMBT;
    var yHoverFormat =  function(n) {
	return n === null ? n : n.toFixed(2);
    };
    var xHoverFormat =  function(n) {
	return n === null ? n : n.toFixed(2);
    };
    if (graphArgs["yformat"] != null){
	yTickFormat = function(n) {
	    var map = graphArgs["yformat"];
	    return map[n];
	};
	yHoverFormat = yTickFormat; 
    }
    var y_ticks = new Rickshaw.Graph.Axis.Y( {
	graph: graph,
	orientation: 'left',
	tickFormat: yTickFormat,
	element: document.getElementById('resultGraphY'),
    } );
    new Rickshaw.Graph.Axis.Time( { graph: graph } );
    var legend = new Rickshaw.Graph.Legend( {
	element: document.getElementById('resultGraphLegend'),
	graph: graph
    } );
    var hoverDetail = new Rickshaw.Graph.HoverDetail( {
	graph: graph,
	yFormatter: yHoverFormat
    } );
    var shelving = new Rickshaw.Graph.Behavior.Series.Toggle( {
	graph: graph,
	legend: legend
    } );      
    graph.render();
}