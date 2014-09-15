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
    tipe: '',
    init: function(aid, pid, uid, tipe) {
        SubmissionsChart.tipe = tipe;
        $(function() {
            $('.select-chart').change(function() {
                $('#submissions-chart').empty();
                SubmissionsChart.load($('#' + SubmissionsChart.tipe + '-dropdown-label').attr(SubmissionsChart.tipe + 'id'), $('#x').val(), $('#y').val());
            });
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                SubmissionsChart.buidDropdown('project', pid, data['projects']);
                $.getJSON('users', function(data) {
                    if (not(data['users'])) {
                        return;
                    }
                    SubmissionsChart.buidDropdown('user', uid, data['users']);
                    var id = SubmissionsChart.tipe === 'user' ? uid : pid;
                    $.getJSON('assignments?' + SubmissionsChart.tipe + '-id=' + id, function(data) {
                        if (not(data['assignments'])) {
                            return;
                        }
                        SubmissionsChart.buidDropdown('assignment', aid, data['assignments']);
                    });
                });
            });
        });
    },


    buidDropdown: function(tipe, id, vals) {
        for (var i = 0; i < vals.length; i++) {
            var currentId = tipe === 'user' ? vals[i].Name : vals[i].Id;
            $('#' + tipe + '-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="#" ' + tipe + 'id="' + currentId + '">' + vals[i].Name + '</a></li>');
            if (id === currentId) {
                $('#' + tipe + '-dropdown-label').attr(tipe + 'id', id);
                $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> ' + vals[i].Name + ' <span class="caret"></span></h4>');
                if (tipe === SubmissionsChart.tipe) {
                    SubmissionsChart.addOptions(id);
                }
            }
        }
        if ($('#' + tipe + '-dropdown-label').attr(tipe + 'id') === undefined) {
            $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> None Selected <span class="caret"></span></h4>');
        }
        $('#' + tipe + '-dropdown ul.dropdown-menu a').on('click', function() {
            $('#table-submissions > tbody').empty();
            var currentId = $(this).attr(tipe + 'id');
            var currentName = $(this).html();
            var params = {};
            params[tipe + '-id'] = currentId;
            if (tipe !== 'assignment') {
                SubmissionsChart.tipe = tipe;
                params['assignment-id'] = 'all';
            }
            setContext(params);
            $('#' + tipe + '-dropdown-label').attr(tipe + 'id', currentId);
            $('#' + tipe + '-dropdown-label h4').html('<small>' + tipe + '</small> ' + currentName + ' <span class="caret"></span>');
            SubmissionsChart.addOptions(currentId);
        });
    },


    addOptions: function(id) {
        var x = $('#x').val();
        var y = $('#y').val();
        $('.select-chart').empty();
        $.getJSON('chart-options?' + SubmissionsChart.tipe + '-id=' + id, function(data) {
            var o = data['options'];
            if (not(o)) {
                console.log(data);
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
            SubmissionsChart.load(id, x, y);
        });
    },

    load: function(id, x, y) {
        var params = {
            'type': 'submission',
            'id': id,
            'x': x,
            'y': y,
            'submission-type': SubmissionsChart.tipe
        };
        $.getJSON('chart-data', params, function(data) {
            SubmissionsChart.create(data['chart-data'], data['chart-info']);
            $('#checkbox-outliers').click(function() {
                SubmissionsChart.create(data['chart-data'], data['chart-info']);
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

    create: function(chartData, chartInfo) {
        $('#submissions-chart').empty();
        if (not(chartData) || not(chartInfo)) {
            return;
        }
        var m = [10, 150, 100, 100];
        var w = 1100 - m[1] - m[3];
        var h = 480 - m[0] - m[2];
        var radius = 10;
        var y = d3.scale.linear()
            .domain(SubmissionsChart.extent(chartData, SubmissionsChart.getY))
            .range([h, 0]);

        var x = d3.scale.linear()
            .domain(SubmissionsChart.extent(chartData, getX))
            .range([0, w]);

        var loadX = function(d) {
            return x(getX(d));
        };

        var loadY = function(d) {
            return y(SubmissionsChart.getY(d));
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
                        return 'filesview?submission-id=' + d.key;
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

        var sub = chartBody.selectAll('.link')
            .data(chartData)
            .enter()
            .append('svg:a')
            .attr('xlink:href', function(d) {
                return 'filesview?submission-id=' + d.key;
            })
            .attr('class', 'link')
            .attr('fill', SubmissionsChart.colour)
            .attr('transform', function(d) {
                return 'translate(' + loadX(d) + ',' + loadY(d) + ')';
            });

        sub.append('svg:circle')
            .attr('fill', SubmissionsChart.colour)
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

        sub.append('text')
            .attr('class', 'title')
            .attr('dy', '-1.0em')
            .attr('style', function(d) {
                return 'text-anchor: middle; fill: ' + SubmissionsChart.colour(d) + ';';
            })
            .attr('font-size', '10px')
            .text(function(d) {
                return d.project;
            });
    },
    colour: function(d) {
        return d.outlier ? 'red' : 'black';
    },
    getY: function(d) {
        return d.outlier && $('#checkbox-outliers').prop('checked') ? d.outlier : d.y;
    }
};

var SubmissionsView = {
    init: function(aid, pid, uid, tipe) {
        $(function() {
            $("#table-submissions").tablesorter({
                theme: 'bootstrap',
                dateFormat: 'ddmmyyyy'
            });
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                SubmissionsView.buidDropdown('project', pid, data['projects']);
                $.getJSON('users', function(data) {
                    if (not(data['users'])) {
                        return;
                    }
                    SubmissionsView.buidDropdown('user', uid, data['users']);
                    var id = tipe === 'user' ? uid : pid;
                    $.getJSON('assignments?' + tipe + '-id=' + id, function(data) {
                        if (not(data['assignments'])) {
                            return;
                        }
                        SubmissionsView.buidDropdown('assignment', aid, data['assignments']);
                        SubmissionsView.load();
                    });
                });
            });
        });
    },

    buidDropdown: function(tipe, id, vals) {
        for (var i = 0; i < vals.length; i++) {
            var currentId = tipe === 'user' ? vals[i].Name : vals[i].Id;
            $('#' + tipe + '-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="#" ' + tipe + 'id="' + currentId + '">' + vals[i].Name + '</a></li>');
            if (id === currentId) {
                $('#' + tipe + '-dropdown-label').attr(tipe + 'id', id);
                $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> ' + vals[i].Name + ' <span class="caret"></span></h4>');
            }
        }
        if ($('#' + tipe + '-dropdown-label').attr(tipe + 'id') === undefined) {
            $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> None Selected <span class="caret"></span></h4>');
        }
        $('#' + tipe + '-dropdown ul.dropdown-menu a').on('click', function() {
            $('#table-submissions > tbody').empty();
            var currentId = $(this).attr(tipe + 'id');
            var currentName = $(this).html();
            var params = {};
            params[tipe + '-id'] = currentId;
            if (tipe !== 'assignment') {
                params['assignment-id'] = 'all';
            }
            setContext(params);
            $('#' + tipe + '-dropdown-label').attr(tipe + 'id', currentId);
            $('#' + tipe + '-dropdown-label h4').html('<small>' + tipe + '</small> ' + currentName + ' <span class="caret"></span>');
            SubmissionsView.load();
        });
    },

    load: function() {
        $('#table-submissions > tbody').empty();
        var uid = $('#user-dropdown-label').attr('userid');
        var pid = $('#project-dropdown-label').attr('projectid');
        var aid = $('#assignment-dropdown-label').attr('assignmentid');
        var params = {
            'counts': true,
            'assignment-id': aid,
            'project-id': pid,
            'user-id': uid
        }
        $.getJSON('submissions', params, function(data) {
            if (not(data['submissions']) || not(data['counts'])) {
                return;
            }
            var s = data['submissions'];
            var c = data['counts'];
            for (var i = 0; i < s.length; i++) {
                var d = new Date(s[i].Time);
                var uname = s[i].User;
                var aname = '';
                var pname = '';
                $('#assignment-dropdown > ul > li > a[assignmentid]').each(function() {
                    var a = $(this).attr('assignmentid');
                    if (a === s[i].AssignmentId) {
                        aname = $(this).html();
                        return false;
                    }
                });
                $('#project-dropdown > ul > li > a[projectid]').each(function() {
                    var p = $(this).attr('projectid');
                    if (p === s[i].ProjectId) {
                        pname = $(this).html();
                        return false;
                    }
                });
                $('#table-submissions > tbody').append('<tr submissionid="' + s[i].Id + '"><td><a href="filesview?submission-id=' + s[i].Id + '">' + pname + '</a></td><td>' + uname + '</td><td>' + aname + '</td><td>' + d.toLocaleDateString() + '</td><td>' + d.toLocaleTimeString() + '</td><td>' + c[s[i].Id]['source'] + '</td><td>' + c[s[i].Id]['launch'] + '</td><td>' + c[s[i].Id]['test'] + '</td><td>' + c[s[i].Id]['testcases'] + '</td><td>' + c[s[i].Id]['passed'] + ' %</td></tr>');
            }
            $('#table-submissions').trigger('update');
        });
    }
}
