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
        ComparisonChart.init();
        $(function() {
            $('.select-chart').change(function() {
                $('#chart').empty();
                SubmissionsChart.load($('#x').val(), $('#y').val());
            });
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    console.log(data);
                    return;
                }
                View.buidDropdown('project', pid, 'assignmentsview', data['projects']);
                $.getJSON('users', function(data) {
                    if (not(data['users'])) {
                        console.log(data);
                        return;
                    }
                    View.buidDropdown('user', uid, 'assignmentsview', data['users']);
                    var id = tipe === 'user' ? uid : pid;
                    $.getJSON('assignments?' + tipe + '-id=' + id, function(data) {
                        if (not(data['assignments'])) {
                            console.log(data);
                            return;
                        }
                        SubmissionsChart.assDropdown(aid, data['assignments']);
                        SubmissionsChart.addOptions();
                    });
                });
            });
        });
    },

    assDropdown: function(id, vals) {
        for (var i = 0; i < vals.length; i++) {
            var currentId = vals[i].Id;
            $('#assignment-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="#" currentid="' + currentId + '">' + vals[i].Name + '</a></li>');
            if (id === currentId) {
                $('#assignment-dropdown-label').attr('currentid', id);
                $('#assignment-dropdown-label').append('<h4><small>assignment</small> ' + vals[i].Name + ' <span class="caret"></span></h4>');
            }
        }
        $('#assignment-dropdown ul.dropdown-menu a').on('click', function() {
            $('#table-submissions > tbody').empty();
            var currentId = $(this).attr('currentid');
            var currentName = $(this).html();
            var params = {};
            params['assignment-id'] = currentId;
            setContext(params);
            $('#assignment-dropdown-label').attr('currentid', currentId);
            $('#assignment-dropdown-label h4').html('<small>assignment</small> ' + currentName + ' <span class="caret"></span>');
            SubmissionsChart.addOptions();
        });
    },

    addOptions: function() {
        var x = $('#x').val();
        var y = $('#y').val();
        $('.select-chart').empty();
        var url = 'chart-options';
        var count = 0;
        var id = $('#assignment-dropdown-label').attr('currentid');
        $.getJSON('chart-options?type=submission&assignment-id=' + id, function(data) {
            var o = data['options'];
            if (not(o)) {
                console.log(data);
                return;
            }
            for (var i = 0; i < o.length; i++) {
                $('.select-chart').append('<option value="' + o[i].id + '">' + o[i].name + '</option>');
            }
            if (not(x) || $('#x option[value="' + x + '"]').length) {
                x = o[0].id;
            }
            $('#x').val(x);
            if (not(y) || $('#y option[value="' + y + '"]').length) {
                y = o[o.length - 1].id;
            }
            $('#y').val(y);
            SubmissionsChart.load(x, y);
        });
    },

    load: function(x, y) {
        var params = {
            'type': 'submission',
            'x': x,
            'y': y,
            'assignment-id': $('#assignment-dropdown-label').attr('currentid')
        };
        params[SubmissionsChart.tipe + '-id'] = $('#' + SubmissionsChart.tipe + '-dropdown-label').attr(SubmissionsChart.tipe + 'id');
        ComparisonChart.load(params);
    }
};

var SubmissionsView = {
    tipe: '',
    init: function(aid, pid, uid, tipe) {
        SubmissionsView.tipe = tipe;
        $(function() {
            $('#button-filter').on('click', SubmissionsView.load);
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    console.log(data);
                    return;
                }
                View.buidDropdown('project', pid, 'assignmentsview', data['projects']);
                $.getJSON('users', function(data) {
                    if (not(data['users'])) {
                        console.log(data);
                        return;
                    }
                    View.buidDropdown('user', uid, 'assignmentsview', data['users']);
                    var id = tipe === 'user' ? uid : pid;
                    $.getJSON('assignments?' + tipe + '-id=' + id, function(data) {
                        if (not(data['assignments'])) {
                            console.log('could not load assignments', data);
                            return;
                        }
                        SubmissionsView.assDropdown(aid, data['assignments']);
                        SubmissionsView.load();
                    });
                });
            });
        });
    },

    assDropdown: function(id, vals) {
        for (var i = 0; i < vals.length; i++) {
            var currentId = vals[i].Id;
            $('#assignment-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="#" assignmentid="' + currentId + '">' + vals[i].Name + '</a></li>');
            if (id === currentId) {
                $('#assignment-dropdown-label').attr('assignmentid', id);
                $('#assignment-dropdown-label').append('<h4><small>assignment</small> ' + vals[i].Name + ' <span class="caret"></span></h4>');
            }
        }
        if (not($('#assignment-dropdown-label').attr('assignmentid'))) {
            $('#assignment-dropdown-label').append('<h4><small>assignment</small> None Selected <span class="caret"></span></h4>');
        }
        $('#assignment-dropdown ul.dropdown-menu a').on('click', function() {
            $('#table-submissions > tbody').empty();
            var currentId = $(this).attr('assignmentid');
            var currentName = $(this).html();
            var params = {};
            params['assignment-id'] = currentId;
            setContext(params);
            $('#assignment-dropdown-label').attr('assignmentid', currentId);
            $('#assignment-dropdown-label h4').html('<small>assignment</small> ' + currentName + ' <span class="caret"></span>');
            SubmissionsView.load();
        });
    },

    load: function() {
        $('#table-submissions > tbody').empty();
        var aid = $('#assignment-dropdown-label').attr('assignmentid');
        var params = {
            'type': 'submission',
            'assignment-id': aid
        }
        params[SubmissionsView.tipe + '-id'] = $('#' + SubmissionsView.tipe + '-dropdown-label').attr(SubmissionsView.tipe + 'id');
        $.getJSON('table-info', params, function(data) {
            if (not(data['table-info']) || not(data['table-fields'])) {
                console.log(data);
                return;
            }
            var info = data['table-info'];
            var fs = data['table-fields'];
            for (var j = 0; j < fs.length; j++) {
                if (fs[j].id === 'id') {
                    continue;
                }
                var n = toTitleCase(fs[j].name);
                $('#table-submissions > thead > tr').append('<th key="' + fs[j].id + '">' + n + '</th>');
                $('#fields').append('<option value="' + fs[j].id + '">' + n + '</option>');
                if (SubmissionsView.isMetric(fs[j].id)) {
                    $('#table-submissions > thead > tr > th').last().hide();
                } else {
                    $('#fields > option').last().prop('selected', true);
                }
            }
            $('#fields').show();
            $('#fields').multiselect({
                noneSelectedText: 'Add table fields',
                selectedText: '# table fields selected',
                click: function(event, ui) {
                    $('[key="' + ui.value + '"]').toggle();
                    if ($('[key="' + ui.value + '"]').is(":visible")) {
                        $('[key="' + ui.value + '"]').each(function() {
                            $(this).appendTo($(this).parent());
                        });
                    }
                },
                checkAll: function(event, ui) {
                    $('[key]').each(function() {
                        if (!$(this).is(":visible")) {
                            $(this).appendTo($(this).parent());
                        }
                    });
                    $('[key]').show();
                },
                uncheckAll: function(event, ui) {
                    $('[key]').hide();
                }
            });
            for (var i = 0; i < info.length; i++) {
                var s = new Date(info[i].time);
                $('#table-submissions > tbody').append('<tr submissionid="' + info[i].id + '"><td key="name"><a href="filesview?submission-id=' + info[i].id + '">' + info[i].name + '</a></td><td key="start date">' + s.toLocaleDateString() + '</td><td key="start time">' + s.toLocaleTimeString() + '</td></tr>');
                var s = '#table-submissions > tbody > tr[submissionid="' + info[i].id + '"]';
                for (var j = 0; j < fs.length; j++) {
                    if (!SubmissionsView.isMetric(fs[j].id)) {
                        continue;
                    }
                    $(s).append('<td key="' + fs[j].id + '">' + info[i][fs[j].id].value + ' ' + info[i][fs[j].id].unit + '</td>');
                    $(s + ' td').last().hide();
                }
            }
            $("#table-submissions").tablesorter({
                theme: 'bootstrap',
                dateFormat: 'ddmmyyyy'
            });
        });
    },
    isMetric: function(k) {
        return notEqual(k, ['id', 'name', 'start date', 'start time']);
    }
}
