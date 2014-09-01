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

var OverviewChart = {
    tipe: '',
    init: function(tipe) {
        $(function() {
            OverviewChart.tipe = tipe;
            var params = {
                'type': 'overview',
                'view': tipe,
            };
            $.getJSON('chart-data', params, function(data) {
                OverviewChart.active = data['categories'][0];
                OverviewChart.create(data['chart'], data['categories'], data['categories'][0]);
            });
        });
    },

    create: function(chartData, categories, active) {
        if (chartData === null) {
            return;
        }
        d3.select('#overview-chart').html('');
        var m = [10, 150, 50, 100];
        var w = 1100 - m[1] - m[3];
        var h = 500 - m[0] - m[2];
        var size = w / chartData.length;
        var yMax = h / 3;
        var duration = 1500;
        var ease = 'quad-in-out';
        var currentY = function(d) {
            return d[active];
        };
        var names = chartData.map(function(d) {
            return d.key;
        });
        var chartColour = function(t) {
            return d3.scale.category10()
                .domain(categories)(t);
        };
        var getAssignments = function(d) {
            var url = '';
            if (OverviewChart.tipe === 'project') {
                for (var i = 0; i < chartData.length; i++) {
                    if (chartData[i].key === d) {
                        url = 'assignmentsview?project-id=' + chartData[i].id;
                        break;
                    }
                }
            } else if (OverviewChart.tipe === 'user') {
                url = 'assignmentsview?user-id=' + d;
            }
            if (url !== '') {
                window.location = url;
            }
        };
        var xScale = d3.scale.ordinal()
            .domain(names)
            .rangeBands([0, w]);
        var xPos = function(d) {
            return xScale(d.key);
        };
        var domain = [0, d3.max(chartData, currentY)];
        var yScale = d3.scale.linear()
            .domain(d3.extent(chartData, currentY))
            .range([h, 0]);
        var yPos = function(d) {
            return h - height(d);
        };
        var height = function(d) {
            return d3.scale.linear()
                .domain(domain)
                .range([0, yMax])(d[active]);
        };
        var x = w / chartData.length - 10;
        var xAxis = d3.svg.axis()
            .scale(xScale)
            .tickSize(-h)
            .orient('bottom')
            .tickSubdivide(true);
        var yAxis = d3.svg.axis()
            .scale(yScale)
            .ticks(5)
            .orient('left')
            .tickSubdivide(true);

        var chart = d3.select('#overview-chart')
            .append('svg:svg')
            .attr('width', w + m[1] + m[3])
            .attr('height', h + m[0] + m[2])
            .append('svg:g')
            .attr('transform', 'translate(' + m[3] + ',' + m[0] + ')');
        chart.append('svg:rect')
            .attr('width', w)
            .attr('height', h)
            .attr('class', 'plot');
        chart.append('svg:g')
            .attr('class', 'y axis')
            .attr('transform', 'translate(-25,0)')
            .call(yAxis);
        chart.append('svg:g')
            .attr('font-size', '10px')
            .attr('class', 'x axis')
            .attr('transform', 'translate(0,' + h + ')')
            .call(xAxis);
        chart.select('.x.axis')
            .selectAll('.tick.major')
            .attr('class', 'tick major clickable')
            .on('click', getAssignments)
        chart.append('text')
            .attr('x', w / 2)
            .attr('y', h + 40)
            .attr('font-size', '20px')
            .style('text-anchor', 'middle')
            .text(OverviewChart.tipe === 'project' ? 'Project' : 'User');
        chart.append('text')
            .attr('font-size', '20px')
            .attr('transform', 'translate(' + (-70) + ',' + (h * 0.5) + ')rotate(-90)')
            .style('text-anchor', 'middle')
            .text(active.toTitleCase());

        chart.append('svg:clipPath')
            .attr('id', 'clip')
            .append('svg:rect')
            .attr('x', -10)
            .attr('y', -10)
            .attr('width', w + 20)
            .attr('height', h + 20);
        var chartBody = chart.append('g')
            .attr('clip-path', 'url(#clip)');
        var bars = chartBody.selectAll('.bar')
            .data(chartData)
            .enter()
            .append('g')
            .attr('class', 'bar');
        bars.append('rect')
            .attr("x", xPos)
            .attr('fill', chartColour(active))
            .attr("y", yPos)
            .attr("width", x)
            .attr("height", height)
            .attr('status', 'stacked')
            .append('title')
            .attr('class', 'description')
            .text(function(d) {
                return d.key +
                    '\n' + d[active] + ' ' + active.toTitleCase();
            });
        var legend = chart.append('g')
            .attr('class', 'legend')
            .attr('height', 100)
            .attr('width', 100)
            .attr('transform', 'translate(-100,0)');
        var changeActive = function(d) {
            OverviewChart.create(chartData, categories, d);
        };
        var legendElements = legend.selectAll('g')
            .data(categories)
            .enter()
            .append('g');
        legendElements
            .attr('class', 'clickable')
            .style('opacity', function(d) {
                return d === active ? 1.0 : 0.5;
            })
            .on('click', changeActive);
        legendElements.append('text')
            .attr('x', 120 + w)
            .attr('y', function(d, i) {
                return i * 20 + 60;
            })
            .attr('font-size', '12px')
            .text(function(d) {
                return d.toTitleCase();
            });
        legendElements.append('rect')
            .attr('class', 'legendrect')
            .attr('x', 100 + w)
            .attr('y', function(d, i) {
                return i * 20 + 50;
            })
            .attr('width', 15)
            .attr('height', 15)
            .style('fill', chartColour);
    }
};

var Overview = {
    setup: function(tipe, tipeNames) {
        $(function() {
            $.getJSON('typecounts?view=' + tipe, function(data) {
                console.log(data);
                if (not(data['typecounts'])) {
                    return;
                }
                var tcs = data['typecounts'];
                var cs = data['categories'];
                for (var i = 0; i < tipeNames.length; i++) {
                    $('tr.info').append('<th>' + tipeNames[i] + '</th>');
                }
                for (var i = 0; i < cs.length; i++) {
                    $('tr.info').append('<th>' + cs[i].toTitleCase() + '</th>');
                }
                for (var i = 0; i < tcs.length; i++) {
                    var tr = '<tr>';
                    if (tipe === 'project') {
                        tr += '<td><a href="assignmentsview?project-id=' + tcs[i].id + '">' + tcs[i].name + '</a></td><td class="rowlink-skip"><a href="#" class="a-info"><span class="glyphicon glyphicon-info-sign"></span><p hidden>' + tcs[i].description + '</p></a></td><td>' + new Date(tcs[i].time).toLocaleString() + '</td><td>' + tcs[i].lang + '</td>';
                    } else {
                        tr += '<td><a href="assignmentsview?user-id=' + tcs[i].name + '">' + tcs[i].name + '</a></td>';
                    }
                    for (var j = 0; j < cs.length; j++) {
                        tr += '<td>' + tcs[i][cs[j]] + '</td>';
                    }
                    $('tbody').append(tr);
                }
                $('#table-overview').tablesorter({
                    theme: 'bootstrap'
                });
                if (tipe === 'project') {
                    $('.a-info').popover({
                        content: function() {
                            var d = $(this).find('p').html();
                            return d === '' ? 'No description' : d;
                        }
                    })
                }
            });

        });
    }
};
