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

var CreateAssignment = {
    init: function() {
        $(function() {
            CreateAssignment.addPickers();
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                var ps = data['projects'];
                for (var i = 0; i < ps.length; i++) {
                    $('#project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                }
            });
            $('#assignment-form').submit(function(e) {
                var sval = $('#datetimepicker-start').val();
                var eval = $('#datetimepicker-end').val();
                if (!sval || !eval) {
                    e.preventDefault();
                    alert('Invalid time values');
                    return;
                }
                var sdate = new Date(sval).getTime();
                var edate = new Date(eval).getTime();
                if (sdate > edate) {
                    e.preventDefault();
                    alert('Invalid time values');
                    return;
                }
                $('[name="assignment-start"]').val(sdate);
                $('[name="assignment-end"]').val(edate);
            });
        });
    },
    addPickers: function() {
        $('#datetimepicker-start').datetimepicker({
            onShow: function(ct) {
                this.setOptions({
                    maxDate: $('#datetimepicker-end').val() ? $('#datetimepicker-end').val() : false
                });
            }
        });
        $('#datetimepicker-end').datetimepicker({
            onShow: function(ct) {
                this.setOptions({
                    minDate: $('#datetimepicker-start').val() ? $('#datetimepicker-start').val() : false
                });
            }
        });
        $('#span-start').attr('showing', false);
        $('#span-end').attr('showing', false);
        $('#span-start').click(function() {
            var s = $(this).attr('showing') === 'true';
            if (!s) {
                $('#datetimepicker-start').datetimepicker('show');
            } else {
                $('#datetimepicker-start').datetimepicker('hide');
            }
            $(this).attr('showing', !s);
        });
        $('#span-end').click(function() {
            var s = $(this).attr('showing') === 'true';
            if (!s) {
                $('#datetimepicker-end').datetimepicker('show');
            } else {
                $('#datetimepicker-end').datetimepicker('hide');
            }
            $(this).attr('showing', !s);
        });
    }
};

var AssignmentsView = {
    tipe: '',
    init: function(tipe, id) {
        AssignmentsView.tipe = tipe;
        $(function() {
            AssignmentsView.addPickers();
            $('#button-filter').on('click', AssignmentsView.load);
            $.getJSON(AssignmentsView.tipe + 's', function(data) {
                if (not(data[AssignmentsView.tipe + 's'])) {
                    return;
                }
                var ts = data[AssignmentsView.tipe + 's'];
                for (var i = 0; i < ts.length; i++) {
                    var tid = AssignmentsView.tipe === 'user' ? ts[i].Name : ts[i].Id;
                    $('#type-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="#" ' + AssignmentsView.tipe + 'id="' + tid + '">' + ts[i].Name + '</a></li>');
                    if (id === tid) {
                        $('#type-dropdown-label').attr(AssignmentsView.tipe + 'id', tid);
                        $('#type-dropdown-label').append('<h4><small>' + AssignmentsView.tipe + '</small> ' + ts[i].Name + ' <span class="caret"></span></h4>');
                    }
                }
                $('#type-dropdown ul.dropdown-menu a').on('click', function() {
                    $('#table-assignments > tbody').empty();
                    var tid = $(this).attr(AssignmentsView.tipe + 'id');
                    var params = {};
                    params[AssignmentsView.tipe + '-id'] = tid;
                    setContext(params);
                    $('#type-dropdown-label').attr(AssignmentsView.tipe + 'id', tid);
                    $('#type-dropdown-label h4').html('<small>' + AssignmentsView.tipe + '</small> ' + $(this).html() + ' <span class="caret"></span>');
                    AssignmentsView.load();
                });
                AssignmentsView.load();
            });
        });
    },

    addPickers: function() {
        $('#datetimepicker-start').datetimepicker({
            onShow: function(ct) {
                this.setOptions({
                    maxDate: $('#datetimepicker-end').val() ? $('#datetimepicker-end').val() : false
                });
            }
        });
        $('#datetimepicker-end').datetimepicker({
            onShow: function(ct) {
                this.setOptions({
                    minDate: $('#datetimepicker-start').val() ? $('#datetimepicker-start').val() : false
                });
            }
        });
        AssignmentsView.pickerButton('start');
        AssignmentsView.pickerButton('end');
    },

    pickerButton: function(n) {
        $('#span-' + n).attr('showing', false);
        $('#span-' + n).click(function() {
            var s = $(this).attr('showing') === 'true';
            if (!s) {
                $('#datetimepicker-' + n).datetimepicker('show');
            } else {
                $('#datetimepicker-' + n).datetimepicker('hide');
            }
            $(this).attr('showing', !s);
        });
    },

    time: function(s) {
        var val = $(s).val();
        if (!val) {
            return -1;
        }
        var d = new Date(val);
        if (d === null || d === undefined) {
            return -1;
        }
        return d.getTime();
    },

    load: function() {
        $('#table-assignments > tbody').empty();
        var tid = $('#type-dropdown-label').attr(AssignmentsView.tipe + 'id');
        var params = {
            'counts': true,
            'min-start': AssignmentsView.time('#datetimepicker-start'),
            'max-end': AssignmentsView.time('#datetimepicker-end')
        };
        params[AssignmentsView.tipe + '-id'] = tid;
        $.getJSON('assignments', params, function(data) {
            if (not(data['assignments']) || not(data['counts'])) {
                return;
            }
            var a = data['assignments'];
            var c = data['counts'];
            for (var i = 0; i < a.length; i++) {
                var s = new Date(a[i].Start);
                var e = new Date(a[i].End);
                $('#table-assignments > tbody').append('<tr assignmentid="' + a[i].Id + '"><td><a href="submissionsview?assignment-id=' + a[i].Id + '">' + a[i].Name + '</a></td><td>' + s.toLocaleDateString() + '</td><td>' + s.toLocaleTimeString() + '</td><td>' + e.toLocaleDateString() + '</td><td>' + e.toLocaleTimeString() + '</td><td>' + c[a[i].Id]['submissions'] + '</td><td>' + c[a[i].Id]['source'] + '</td><td>' + c[a[i].Id]['launch'] + '</td><td>' + c[a[i].Id]['test'] + '</td><td>' + c[a[i].Id]['testcases'] + '</td><td>' + c[a[i].Id]['passed'] + ' %</td></tr>');
            }
            $("#table-submissions").tablesorter({
                theme: 'bootstrap',
                dateFormat: 'ddmmyyyy'
            });
        });
    }
};

var AssignmentsChart = {
    init: function(tipe, id) {
        $(function() {
            AssignmentsChart.addOptions(tipe, id);
            $('.select-chart').change(function() {
                $('#assignments-chart').empty();
                AssignmentsChart.load(tipe, $('#type-dropdown-label').attr(tipe + 'id'), $('#x').val(), $('#y').val());
            });
            $.getJSON(tipe + 's', function(data) {
                if (not(data[tipe + 's'])) {
                    return;
                }
                var ts = data[tipe + 's'];
                for (var i = 0; i < ts.length; i++) {
                    var tid = ts[i].Id ? ts[i].Id : ts[i].Name;
                    $('#type-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="#" ' + tipe + 'id="' + tid + '">' + ts[i].Name + '</a></li>');
                    if (id === tid) {
                        $('#type-dropdown-label').attr(tipe + 'id', tid);
                        $('#type-dropdown-label').append('<h4><small>' + tipe + '</small> ' + ts[i].Name + ' <span class="caret"></span></h4>');
                    }
                }
                $('#type-dropdown ul.dropdown-menu a').on('click', function() {
                    $('#assignments-chart').empty();
                    var tid = $(this).attr(tipe + 'id');
                    var params = {};
                    params[tipe + '-id'] = tid;
                    setContext(params);
                    $('#type-dropdown-label').attr(tipe + 'id', tid);
                    $('#type-dropdown-label h4').html('<small>' + tipe + '</small> ' + $(this).html() + ' <span class="caret"></span>');
                    AssignmentsChart.addOptions(tipe, tid);
                });
            });
        });
    },

    addOptions: function(tipe, id) {
        var x = $('#x').val();
        var y = $('#y').val();
        $('.select-chart').empty();
        $.getJSON('chart-options?' + tipe + '-id=' + id, function(data) {
            var o = data['options'];
            if (not(o)) {
                return;
            }
            for (var i = 0; i < o.length; i++) {
                $('.select-chart').append('<option value="' + o[i].Id + '">' + o[i].Name + '</option>');
            }
            if (x === undefined || x === null || $('#x option[value="' + x + '"]').length) {
                x = o[0].Id;
            }
            $('#x').val(x);
            if (y === undefined || y === null || $('#y option[value="' + y + '"]').length) {
                y = o[o.length - 1].Id;
            }
            $('#y').val(y);
            AssignmentsChart.load(tipe, id, x, y);
        });
    },

    load: function(tipe, id, x, y) {
        var params = {
            'type': 'assignment',
            'id': id,
            'x': x,
            'y': y,
            'assignment-type': tipe
        };
        $.getJSON('chart-data', params, function(data) {
            AssignmentsChart.create(data['chart-data'], data['chart-info'], tipe);
            $('#checkbox-outliers').click(function() {
                AssignmentsChart.create(data['chart-data'], data['chart-info'], tipe);
            });
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

    create: function(chartData, chartInfo, tipe) {
        $('#assignments-chart').empty();
        if (not(chartData) || not(chartInfo)) {
            return;
        }
        var m = [10, 150, 100, 100];
        var w = 1100 - m[1] - m[3];
        var h = 480 - m[0] - m[2];
        var radius = 10;
        var y = d3.scale.linear()
            .domain(AssignmentsChart.extent(chartData, AssignmentsChart.getY))
            .range([h, 0]);

        var x = d3.scale.linear()
            .domain(AssignmentsChart.extent(chartData, getX))
            .range([0, w]);

        var loadX = function(d) {
            return x(getX(d));
        };

        var loadY = function(d) {
            return y(AssignmentsChart.getY(d));
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

        var chart = d3.select('#assignments-chart')
            .append('svg:svg')
            .attr('width', w + m[1] + m[3])
            .attr('height', h + m[0] + m[2])
            .append('svg:g')
            .attr('transform', 'translate(' + m[3] + ',' + m[0] + ')');

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
                    .attr('xlink:href', function(d) {
                        return 'submissionsview?assignment-id=' + d.key;
                    })
                    .attr('class', 'link')
                    .transition()
                    .duration(duration)
                    .ease(ease)
                    .attr('transform', function(d) {
                        return 'translate(' + loadX(d) + ',' + loadY(d) + ')';
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

        var yTitle = chartInfo['y-unit'] === '' ? chartInfo['y'] : chartInfo['y'] + ' (' + chartInfo['y-unit'] + ')';
        var xTitle = chartInfo['x-unit'] === '' ? chartInfo['x'] : chartInfo['x'] + ' (' + chartInfo['x-unit'] + ')';

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

        chart.append('svg:clipPath')
            .attr('id', 'clip')
            .append('svg:rect')
            .attr('x', -10)
            .attr('y', -10)
            .attr('width', w + 20)
            .attr('height', h + 20);

        var chartBody = chart.append('g')
            .attr('clip-path', 'url(#clip)');

        var ass = chartBody.selectAll('.link')
            .data(chartData)
            .enter()
            .append('svg:a')
            .attr('xlink:href', function(d) {
                return 'submissionsview?assignment-id=' + d.key;
            })
            .attr('class', 'link')
            .attr('fill', AssignmentsChart.colour)
            .attr('transform', function(d) {
                return 'translate(' + loadX(d) + ',' + loadY(d) + ')';
            });

        ass.append('svg:circle')
            .attr('fill', AssignmentsChart.colour)
            .attr('r', 5);

        $('.link').tooltip({
            html: true,
            title: function() {
                var d = this.__data__;
                var yVal = d.outlier ? d.outlier : d.y;
                yVal = chartInfo['y-unit'] === '' ? yVal : yVal + ' ' + chartInfo['y-unit'];
                var xVal = chartInfo['x-unit'] === '' ? d.x : d.x + ' ' + chartInfo['x-unit'];
                return '<ul class="list-unstyled list-left"><li><strong>' + d.user + '\'s ' + d.project + '</strong></li><li>' + chartInfo.y + '<span class="span-right">' + yVal + '</span></li><li>' + chartInfo.x + '<span class="span-right">' + xVal + '</span></li></ul><div style="clear: both;"></div>';
            },
            container: 'body'
        });

        ass.append('text')
            .attr('class', 'title')
            .attr('dy', '-1.0em')
            .attr('style', function(d) {
                return 'text-anchor: middle; fill: ' + AssignmentsChart.colour(d) + ';';
            })
            .attr('font-size', '10px')
            .text(function(d) {
                return tipe === 'project' ? d.name : d.project + ' ' + d.name;
            });
    },
    colour: function(d) {
        return d.outlier ? 'red' : 'black';
    },
    getY: function(d) {
        return d.outlier && $('#checkbox-outliers').prop('checked') ? d.outlier : d.y;
    }
};
