function basicGraph(name, legendName) {
    $.getJSON("static/data/"+name+".json", function(data) {
	if (data.length == 0){
	    return;
	}
	var palette = new Rickshaw.Color.Palette();
	for (var i = 0; i < data.length; i++) 
	{
	    data[i]["color"] =  palette.color();
	}
	var graph = new Rickshaw.Graph( {
	    element: document.getElementById(name),
	    width: 540,
	    height: 340,
	    renderer: 'line',
	    series: data
	} );   
	graph.render();
	
	var legend = new Rickshaw.Graph.Legend( {
	    element: document.getElementById(legendName),
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
    });    
}