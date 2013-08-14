function timeLineGraph(graphArgs) {
    if (graphArgs === null || graphArgs['series'].length === 0){
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
    graph.render();
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
    var axes = new Rickshaw.Graph.Axis.Time( {
	element: document.getElementById('resultGraphX'),
	graph: graph
    } );
    axes.render();   
}