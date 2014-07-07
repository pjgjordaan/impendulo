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
var SubmissionsChart = {
    init: function(tipe) {
        $(function() {
            SubmissionsChart.addResults(tipe, $('#' + tipe + '-dropdown-label').attr(tipe + 'id'));
            $('#tool').change(function() {
                SubmissionsChart.load(tipe, $('#' + tipe + '-dropdown-label').attr(tipe + 'id'), $(this).val(), $('[name="score"]').val());
                var params = {
                    'result': $(this).val()
                };
                setContext(params);
            });
            $('[name="score"]').change(function() {
                SubmissionsChart.load(tipe, $('#' + tipe + '-dropdown-label').attr(tipe + 'id'), $('#tool').val(), $(this).val());
            });

            $('#dropdown-' + tipe + ' > ul >  li > a').on('click', function() {
                var tool = $('#tool').val();
                var id = $(this).attr(tipe + 'id');
                SubmissionsChart.addResults(tipe, id, tool);
                var params = {};
                params[tipe + '-id'] = id;
                setContext(params);
                $('#' + tipe + '-dropdown-label').html('<h4><small>' + tipe + '</small> ' + $(this).text() + ' <span class="caret"></span></h4>');
                $('#' + tipe + '-dropdown-label').attr(tipe + 'id', id);
            });
        });
    },


    addResults: function(tipe, id, result) {
        var d = '#tool';
        $(d).empty();
        $.getJSON('results?' + tipe + '-id=' + id, function(data) {
            var r = data['results'];
            if (not(r)) {
                console.log(data);
                return;
            }
            for (var i = 0; i < r.length; i++) {
                $(d).append('<option value="' + r[i].Id + '">' + r[i].Name + '</option>');
            }
            if (result !== undefined && $(d + ' option[value="' + result + '"]').length) {
                $(d).val(result);
            } else {
                result = r[0].Id
            }
            SubmissionsChart.load(tipe, id, result, $('[name="score"]').val());
        });
    },

    load: function(tipe, id, r, score) {
        var params = {
            'type': 'submission',
            'id': id,
            'result': r,
            'submission-type': tipe,
            'score': score
        };
        $.getJSON('chart', params, function(data) {
            SubmissionsChart.create(data['chart'], tipe);
        });
    },


    extent: function(data, f) {
        var e = d3.extent(data, f);
        var s = 0.05 * (e[1] - e[0]);
        if (e[0] == e[1]) {
            s = 10;
        }
        if (e[0] >= 0) {
            e[0] = Math.max(0, e[0] - s);
        } else {
            e[0] -= s;
        }
        if (e[1] <= 100) {
            e[1] = Math.min(100, e[1] + s);
        } else {
            e[1] += s;
        }
        return e;
    },

    create: function(chartData, tipe) {
        if (chartData === null || chartData === undefined || chartData.length === 0) {
            return;
        }
        $('#submissions-chart').empty();
        var m = [10, 150, 100, 100];
        var w = 1100 - m[1] - m[3];
        var h = 480 - m[0] - m[2];
        var tipes = ['Launches', 'Snapshots'];
        var radius = 10;
        var chartColour = function(tipe) {
            return d3.scale.category10()
                .domain(tipes)(tipe);
        };
        var y = d3.scale.linear()
            .domain(SubmissionsChart.extent(chartData, getY))
            .range([h, 0]);

        var x = d3.scale.linear()
            .domain(SubmissionsChart.extent(chartData, getTime))
            .range([0, w]);

        var rs = d3.scale.linear()
            .domain([0, d3.max(chartData, function(d) {
                return d.snapshots;
            })])
            .range([0, radius]);

        var rl = function(offset) {
            var outerRadius = Math.sqrt(offset * offset + radius * radius);
            return d3.scale.linear()
                .domain([0, d3.max(chartData, function(d) {
                    return d.launches;
                })])
                .range([offset, outerRadius])
        };

        chartData = chartData.map(function(d) {
            d.rs = rs(d.snapshots);
            d.rlSmall = rl(0)(d.launches);
            d.rlBig = rl(d.rs)(d.launches);
            return d;
        });

        var loadDate = function(d, i) {
            return x(getTime(d));
        };

        var loadY = function(d) {
            return y(getY(d));
        }
        var xAxis = d3.svg.axis()
            .scale(x)
            .ticks(7)
            .tickSize(-h)
            .orient('bottom')
            .tickSubdivide(true);

        var yAxis = d3.svg.axis()
            .scale(y)
            .ticks(5)
            .tickSubdivide(true)
            .orient('right');

        var chart = d3.select('#submissions-chart')
            .append('svg:svg')
            .attr('width', w + m[1] + m[3])
            .attr('height', h + m[0] + m[2])
            .append('svg:g')
            .attr('transform', 'translate(' + m[3] + ',' + m[0] + ')');

        var zoom = d3.behavior.zoom()
            .x(x)
            .on('zoom', function() {
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
                chartBody.selectAll('.link')
                    .attr('xlink:href', function(d) {
                        return 'getfiles?submission-id=' + d.key;
                    })
                    .attr('class', 'link')
                    .transition()
                    .duration(duration)
                    .ease(ease)
                    .attr('transform', function(d) {
                        return 'translate(' + loadDate(d) + ',' + loadY(d) + ')';
                    });

            });
        chart.call(zoom);

        chart.append('svg:rect')
            .attr('width', w)
            .attr('height', h)
            .attr('class', 'plot');

        chart.append('svg:g')
            .attr('class', 'x axis')
            .attr('transform', 'translate(0,' + h + ')')
            .call(xAxis);

        chart.append('text')
            .attr('x', w / 2)
            .attr('y', h + 40)
            .attr('font-size', '20px')
            .style('text-anchor', 'middle')
            .text('Running Time (s)');

        chart.append('text')
            .attr('font-size', '20px')
            .attr('transform', 'translate(' + (w + 120) + ',' + (h * 0.6) + ')rotate(90)')
            .style('text-anchor', 'middle')
            .text(chartData[0]["description"]);

        chart.append('svg:g')
            .attr('class', 'y axis')
            .attr('font-size', '10px')
            .attr('transform', 'translate(' + (w + 25) + ',0)')
            .call(yAxis);

        chart.append('svg:clipPath')
            .attr('id', 'clip')
            .append('svg:rect')
            .attr('x', -10)
            .attr('y', -10)
            .attr('width', w + 20)
            .attr('height', h + 20);

        var chartBody = chart.append('g')
            .attr('clip-path', 'url(#clip)');

        var sub = chartBody.selectAll('.link')
            .data(chartData)
            .enter()
            .append('svg:a')
            .attr('xlink:href', function(d) {
                return 'getfiles?submission-id=' + d.key;
            })
            .attr('class', 'link')
            .attr('transform', function(d) {
                return 'translate(' + loadDate(d) + ',' + loadY(d) + ')';
            });

        sub.append('svg:circle')
            .attr('class', 'launches')
            .attr('key', 'chartLaunches')
            .attr('fill', chartColour('Launches'))
            .attr('r', SubmissionsChart.rlBig);


        sub.append('svg:circle')
            .attr('class', 'snapshots')
            .attr('key', 'chartSnapshots')
            .attr('fill', chartColour('Snapshots'))
            .attr('r', SubmissionsChart.rSnapshot);

        $('.link').tooltip({
            html: true,
            title: function() {
                var d = this.__data__;
                var yVal = d.outlier ? d.outlier : d.y;
                return '<ul class="list-unstyled list-left"><li><strong>' + d.user + '\'s ' + d.project + '</strong></li><li>' + d.description + '<span class="span-right">' + yVal + '</span></li><li>Time<span class="span-right">' + d.time + 's</span></li><li>Snapshots<span class="span-right">' + d.snapshots + '</span></li><li>Launches<span class="span-right">' + d.launches + '</span></li></ul><div style="clear: both;"></div>';
            },
            container: 'body'
        });

        sub.append('text')
            .attr('class', 'title')
            .attr('dy', '-1.0em')
            .attr('style', function(d) {
                var s = 'text-anchor: middle; fill: ';
                return d.outlier ? s + 'red;' : s + 'black;';
            })
            .attr('font-size', '10px')
            .text(function(d) {
                return tipe === 'project' ? d.user : d.project;
            });

        var legend = chart.append('g')
            .attr('class', 'legend')
            .attr('height', 100)
            .attr('width', 100)
            .attr('transform', 'translate(-100,0)');

        var legendElements = legend.selectAll('g')
            .data(tipes)
            .enter()
            .append('g');

        legendElements.append('text')
            .attr('class', 'clickable')
            .attr('x', 20)
            .attr('y', function(d, i) {
                return i * 20 + 60;
            })
            .attr('key', SubmissionsChart.legendKey)
            .attr('font-size', '12px')
            .text(function(d) {
                return d;
            })
            .on('click', SubmissionsChart.toggleVisibility);

        legendElements.append('rect')
            .attr('class', 'legendrect clickable')
            .attr('x', 0)
            .attr('y', function(d, i) {
                return i * 20 + 50;
            })
            .attr('key', SubmissionsChart.legendKey)
            .attr('width', 15)
            .attr('height', 15)
            .style('fill', chartColour)
            .on('click', SubmissionsChart.toggleVisibility);

    },

    toggleVisibility: function(d) {
        var duration = 500;
        var ease = 'linear';
        var visibleSnapshots = d3.select('[key=legendSnapshots]').style('opacity') === '1';
        var visibleLaunches = d3.select('[key=legendLaunches]').style('opacity') === '1'
        var legendOpacity = 1.0;
        if (d === 'Launches') {
            legendOpacity = visibleLaunches ? 0.3 : 1.0;
        } else {
            legendOpacity = visibleSnapshots ? 0.3 : 1.0;
        }
        if (visibleSnapshots && visibleLaunches) {
            if (d === 'Snapshots') {
                SubmissionsChart.hideSnapshots();
                SubmissionsChart.smallLaunches();
            } else {
                SubmissionsChart.hideLaunches();
            }
        } else if (visibleSnapshots) {
            if (d === 'Snapshots') {
                SubmissionsChart.hideSnapshots();
            } else {
                SubmissionsChart.bigLaunches();
                SubmissionsChart.showSnapshots();
            }
        } else if (visibleLaunches) {
            if (d === 'Snapshots') {
                SubmissionsChart.bigLaunches();
                SubmissionsChart.showSnapshots();
            } else {
                SubmissionsChart.hideLaunches()
            }
        } else {
            if (d === 'Snapshots') {
                SubmissionsChart.showSnapshots();
            } else {
                SubmissionsChart.smallLaunches()
            }
        }
        d3.selectAll('[key=legend' + d + ']')
            .transition()
            .duration(duration)
            .ease(ease)
            .style('opacity', legendOpacity);
    },

    legendKey: function(d) {
        return 'legend' + d;
    },

    rSnapshot: function(d) {
        return d.rs;
    },

    rlSmall: function(d) {
        return d.rlSmall;
    },

    rlBig: function(d) {
        return d.rlBig;
    },

    setRadius: function(key, r) {
        d3.selectAll('[key=chart' + key + ']')
            .transition()
            .duration(500)
            .ease('linear')
            .attr('r', r);
    },

    hideSnapshots: function() {
        SubmissionsChart.setRadius('Snapshots', 0);
    },
    showSnapshots: function() {
        SubmissionsChart.setRadius('Snapshots', SubmissionsChart.rSnapshot);
    },

    hideLaunches: function() {
        SubmissionsChart.setRadius('Launches', 0);
    },


    smallLaunches: function() {
        SubmissionsChart.setRadius('Launches', SubmissionsChart.rlSmall);
    },

    bigLaunches: function() {
        SubmissionsChart.setRadius('Launches', SubmissionsChart.rlBig);
    }
};
