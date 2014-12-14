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
var ComparisonChart = {
    params: {},
    loadParams: {},
    colours: ['A30004', '#A40206', '#A50509', '#A7080C', '#A80A0E', '#AA0D11', '#AB1014', '#AD1317', '#AE1519', '#B0181C', '#B11B1F', '#B31E21', '#B42024', '#B52327', '#B7262A', '#B8292C', '#BA2B2F', '#BB2E32', '#BD3134', '#BE3437', '#C0363A', '#C1393D', '#C33C3F', '#C43F42', '#C64145', '#C74447', '#C8474A', '#CA4A4D', '#CB4C50', '#CD4F52', '#CE5255', '#D05558', '#D1575A', '#D35A5D', '#D45D60', '#D66063', '#D76265', '#D96568', '#DA686B', '#DB6B6D', '#DD6D70', '#DE7073', '#E07376', '#E17678', '#E3787B', '#E47B7E', '#E67E80', '#E78183', '#E98386', '#EA8689', '#EC898B', '#ED8C8E', '#EE8E91', '#F09193', '#F19496', '#F39799', '#F4999C', '#F69C9E', '#F79FA1', '#F9A2A4', '#FAA4A6', '#FCA7A9', '#FDAAAC', '#FFADAF'],
    init: function(params) {
        ComparisonChart.params = params;
        ComparisonChart.colours.reverse();
        $('#checkbox-labels').change(function() {
            $('.link text').toggle();
        });
        $('#checkbox-mean').change(function() {
            $('#mean').toggle();
        });
        $('#checkbox-stddev').change(function() {
            $('.stddev').toggle();
        });
        $('#checkbox-outliers-hide').change(function() {
            $('.link[outlier="true"]').toggle();
        });
        $('.select-chart').change(function() {
            ComparisonChart.load($('#x').val(), $('#y').val(), $('#granularity').val());
        });
        ComparisonChart.addOptions();
    },

    addOptions: function() {
        var x = $('#x').val();
        var y = $('#y').val();
        $.getJSON('chart-options', ComparisonChart.params, function(data) {
            var o = data['options'];
            if (not(o)) {
                console.log(data);
                return;
            }
            for (var i = 0; i < o.length; i++) {
                $('.select-axis').append('<option value="' + o[i].id + '">' + o[i].name + '</option>');
            }
            if (not(x) || !$('#x option[value="' + x + '"]').length) {
                x = o[0].id;
            }
            $('#x').val(x);
            if (not(y) || !$('#y option[value="' + y + '"]').length) {
                y = o[o.length - 1].id;
            }
            $('#y').val(y);
            ComparisonChart.load(x, y, $('#granularity').val());
        });
    },

    load: function(x, y, g) {
        $('#chart').empty();
        gs = not(g) ? '' : '&granularity=' + g;
        $.getJSON('chart-data?x=' + x + '&y=' + y + gs, ComparisonChart.params, function(data) {
            ComparisonChart.create(data['chart-data'], data['chart-info']);
            $('#checkbox-outliers-adjust').change(function() {
                ComparisonChart.create(data['chart-data'], data['chart-info']);
            });
        });
    },

    create: function(data, info) {
        $('#chart').empty();
        if (not(data) || not(info)) {
            $('#chart').append('<h3 class="centered">No Data</h3>');
            return;
        }
        var m = [10, 150, 100, 10];
        var w = 1100 - m[1] - m[3];
        var h = 480 - m[0] - m[2];
        var radius = 5;
        var maxX = d3.max(data, ComparisonChart.getActualX);
        var minX = d3.min(data, ComparisonChart.getActualX);
        var minXO = d3.min(data, ComparisonChart.getX);
        var maxY = d3.max(data, ComparisonChart.getActualY);
        var maxYO = d3.max(data, ComparisonChart.getY);
        var yExtent = extent(data, ComparisonChart.getY);
        var xExtent = extent(data, ComparisonChart.getX);
        var y = d3.scale.linear()
            .domain(yExtent)
            .range([h, 0]);
        var x = d3.scale.linear()
            .domain(xExtent)
            .range([0, w]);
        var xScale = w / maxX;
        var loadX = function(d) {
            return x(ComparisonChart.getX(d));
        };
        var loadY = function(d) {
            return y(ComparisonChart.getY(d));
        }
        var loadRX = function(d) {
            return x(!$('#checkbox-outliers-adjust').prop('checked') ? minX + d.rx : minXO + d['rx_o']);
        }
        var loadRY = function(d) {
            return y(!$('#checkbox-outliers-adjust').prop('checked') ? maxY - d.ry : maxYO - d['ry_o']);
        }
        var actualCombined = function(d) {
            return ComparisonChart.getActualX(d) / maxX + ComparisonChart.getActualY(d) / maxY;
        };
        var tooltip = function() {
            $('ui.tooltip').siblings('.tooltip').remove();
            var d = this.__data__;
            var yVal = info['y'] + ': ' + ComparisonChart.getActualY(d) + ' ' + (info['y-unit'] === '' ? '' : info['y-unit']);
            var xVal = info['x'] + ': ' + ComparisonChart.getActualX(d) + ' ' + (info['x-unit'] === '' ? '' : info['x-unit']);
            return '<ul class="list-unstyled list-left"><li><strong>' + d.title + '</strong></li><li>' + yVal + '</li><li>' + xVal + '</li></ul><div style="clear: both;"></div>';
        };
        var cvals = intervals(extent(data, actualCombined), ComparisonChart.colours.length);
        var getColour = d3.scale.ordinal()
            .domain(cvals)
            .range(ComparisonChart.colours);
        var chartColour = function(d) {
            return getColour(closest(cvals, actualCombined(d)));
        };
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
        var chart = d3.select('#chart')
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
            .attr('class', 'x axis')
            .attr('transform', 'translate(0,' + h + ')')
            .call(xAxis);
        var yTitle = info['y-unit'] === '' ? info['y'] : info['y'] + ' (' + info['y-unit'] + ')';
        var xTitle = info['x-unit'] === '' ? info['x'] : info['x'] + ' (' + info['x-unit'] + ')';
        chart.append('text')
            .attr('x', w / 2)
            .attr('y', h + 40)
            .attr('font-size', '20px')
            .style('text-anchor', 'middle')
            .text(xTitle);
        chart.append('text')
            .attr('font-size', '20px')
            .attr('transform', 'translate(' + (w + 120) + ',' + (h * 0.6) + ')rotate(90)')
            .style('text-anchor', 'middle')
            .text(yTitle);
        chart.append('svg:g')
            .attr('class', 'y axis')
            .attr('font-size', '10px')
            .attr('transform', 'translate(' + (w + 25) + ',0)')
            .call(yAxis);
        chart.append('defs')
            .append('svg:clipPath')
            .attr('id', 'clip')
            .append('svg:rect')
            .attr('x', 0)
            .attr('y', 0)
            .attr('width', w)
            .attr('height', h);

        var chartBody = chart.append('g')
            .attr('clip-path', 'url(#clip)');

        var sd = info.stddev;
        sd.c = '#9ecae1';
        var stddev2 = {
            "x": sd.x,
            "y": sd.y,
            "x_o": sd.x_o,
            "y_o": sd.y_o,
            "rx": 2 * sd.rx,
            "ry": 2 * sd.ry,
            "rx_o": 2 * sd.rx_o,
            'ry_o': 2 * sd.ry_o,
            'c': '#c6dbef'
        };
        chartBody
            .selectAll('.stddev')
            .data([stddev2, sd])
            .enter()
            .append('svg:ellipse')
            .attr('class', 'stddev')
            .attr('cx', loadX)
            .attr('cy', loadY)
            .attr('fill', function(d) {
                return d.c;
            })
            .attr('opacity', '0.5')
            .attr('rx', loadRX)
            .attr('ry', loadRY);
        if (!$('#checkbox-stddev').prop('checked')) {
            $('.stddev').hide();
        }

        chartBody
            .selectAll('#mean')
            .data([info.mean])
            .enter()
            .append('svg:circle')
            .attr('id', 'mean')
            .attr('transform', function(d) {
                return 'translate(' + loadX(d) + ',' + loadY(d) + ')';
            })
            .attr('fill', '#31a354')
            .attr('r', radius);

        $('#mean').tooltip({
            html: true,
            title: tooltip,
            container: 'body'
        });
        if (!$('#checkbox-mean').prop('checked')) {
            $('#mean').hide();
        }

        var link = chartBody.selectAll('.link')
            .data(data)
            .enter()
            .append('svg:a')
            .attr('xlink:href', function(d) {
                return d.url;
            })
            .attr('class', 'link')
            .attr('fill', chartColour)
            .attr('outlier', function(d) {
                return d['y_o'] !== undefined || d['x_o'] !== undefined;
            })
            .attr('transform', function(d) {
                return 'translate(' + loadX(d) + ',' + loadY(d) + ')';
            });
        link.append('svg:circle')
            .attr('fill', chartColour)
            .attr('r', radius);
        $('.link').tooltip({
            html: true,
            title: tooltip,
            container: 'body'
        });
        link.append('text')
            .attr('class', 'title')
            .attr('dy', '-1.0em')
            .attr('style', function(d) {
                return 'text-anchor: middle; fill: black;';
            })
            .attr('font-size', '10px')
            .text(function(d) {
                return d.title;
            });
        if (!$('#checkbox-labels').prop('checked')) {
            $('.link text.title').hide();
        }
        if ($('#checkbox-outliers-hide').prop('checked')) {
            $('.link[outlier="true"]').hide();
        };

        var zoom = d3.behavior.zoom()
            .x(x)
            .y(y)
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
                    .transition()
                    .duration(duration)
                    .ease(ease)
                    .attr('transform', function(d) {
                        return 'translate(' + loadX(d) + ',' + loadY(d) + ')';
                    });
                chartBody.selectAll('#mean')
                    .transition()
                    .duration(duration)
                    .ease(ease)
                    .attr('transform', function(d) {
                        return 'translate(' + loadX(d) + ',' + loadY(d) + ')';
                    });
                chartBody.selectAll('.stddev')
                    .transition()
                    .duration(duration)
                    .ease(ease)
                    .attr('transform', 'translate(' + d3.event.translate[0] + ',' + d3.event.translate[1] + ')scale(' + d3.event.scale + ')');
            });
        chart.call(zoom);
    },

    getX: function(d) {
        return d['x_o'] && $('#checkbox-outliers-adjust').prop('checked') ? d['x_o'] : d.x;
    },

    getActualX: function(d) {
        return d.x;
    },

    getY: function(d) {
        return d['y_o'] && $('#checkbox-outliers-adjust').prop('checked') ? d['y_o'] : d.y;
    },

    getActualY: function(d) {
        return d.y;
    },

}
