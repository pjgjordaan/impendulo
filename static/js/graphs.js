function timeLineGraph(name, jsonData) {
    if (jsonData === null || jsonData.length === 0){
	return;
    }
    var palette = new Rickshaw.Color.Palette();
    for (var i = 0; i < jsonData.length; i++) 
    {
	jsonData[i]["color"] =  palette.color();
    }
    var graph = new Rickshaw.Graph( {
	element: document.getElementById(name),
	width: 700,
	height: 400,
	renderer: 'line',
	series: jsonData
    } );   
    var y_ticks = new Rickshaw.Graph.Axis.Y( {
	graph: graph,
	orientation: 'left',
	tickFormat: Rickshaw.Fixtures.Number.formatKMBT,
	element: document.getElementById(name+'Y'),
    } );
    graph.render();
    var legend = new Rickshaw.Graph.Legend( {
	element: document.getElementById(name+'Legend'),
	graph: graph
    } );
    var hoverDetail = new Rickshaw.Graph.HoverDetail( {
	graph: graph
    } );
    var shelving = new Rickshaw.Graph.Behavior.Series.Toggle( {
	graph: graph,
	legend: legend
    } );
    
    var axes = new Rickshaw.Graph.Axis.Time( {
	graph: graph
    } );
    axes.render();   
}

function javacGraph(jsonData) {   
    if (jsonData === null || jsonData.length === 0){
	return;
    }
    var palette = new Rickshaw.Color.Palette();
    for (var i = 0; i < jsonData.length; i++) 
    {
	jsonData[i]["color"] =  palette.color();
    }
    var graph = new Rickshaw.Graph( {
	element: document.getElementById("resultGraph"),
	width: 700,
	height: 400,
	renderer: 'scatterplot',
	series: jsonData
    } );   
    var yformat = function(n) {
	var map = {
	    0: 'No Compile',
	    1: 'Compiled'
	};
	return map[n];
    };
    var yhover = function(y) {
	return y === 0 ? "No Compile" : "Compiled Successfully";
    };  
    var y_ticks = new Rickshaw.Graph.Axis.Y( {
	graph: graph,
	orientation: 'left',
	tickFormat: yformat,
	element: document.getElementById('resultGraphY'),
    } );
    graph.render();
    var legend = new Rickshaw.Graph.Legend( {
	element: document.getElementById('resultGraphLegend'),
	graph: graph
    } );
    var hoverDetail = new Rickshaw.Graph.HoverDetail( {
	graph: graph,
	yFormatter: yhover
    } );
    var shelving = new Rickshaw.Graph.Behavior.Series.Toggle( {
	graph: graph,
	legend: legend
    } );
    
    var axes = new Rickshaw.Graph.Axis.Time( {
	graph: graph
    } );
    axes.render();
}

