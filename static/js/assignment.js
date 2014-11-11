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
                    console.log(data);
                    return;
                }
                var ps = data['projects'];
                for (var i = 0; i < ps.length; i++) {
                    $('#project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                }
                $('#project-id').change(function() {
                    CreateAssignment.loadSkeletons($(this).val());
                });
                CreateAssignment.loadSkeletons(ps[0].Id);
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

    loadSkeletons: function(pid) {
        $('#skeleton-id').empty();
        $.getJSON('skeletons?project-id=' + pid, function(data) {
            var sk = data['skeletons'];
            if (not(sk)) {
                console.log(data);
                return;
            }
            for (var i = 0; i < sk.length; i++) {
                $('#skeleton-id').append('<option value="' + sk[i].Id + '">' + sk[i].Name + '</option>');
            }
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
            $('#btn-all-submissions').attr('href', 'submissionsview?' + tipe + '-id=' + id);
            $('#button-filter').on('click', AssignmentsView.load);
            $.getJSON(AssignmentsView.tipe + 's', function(data) {
                if (not(data[AssignmentsView.tipe + 's'])) {
                    console.log(data);
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

    load: function() {
        $('#fields').empty();
        $('#table-assignments > tbody').empty();
        $('#table-assignments > thead > tr').empty();
        var tid = $('#type-dropdown-label').attr(AssignmentsView.tipe + 'id');
        var params = {
            'type': 'assignment',
            'assignment-type': AssignmentsView.tipe,
            'id': tid
        };
        $.getJSON('table-info', params, function(data) {
            if (not(data['table-info']) || not(data['table-fields'])) {
                console.log(data);
                return;
            }
            var a = data['table-info'];
            var fs = data['table-fields'];
            for (var j = 0; j < fs.length; j++) {
                if (fs[j].id === 'id') {
                    continue;
                }
                var n = toTitleCase(fs[j].name);
                $('#table-assignments > thead > tr').append('<th key="' + fs[j].id + '">' + n + '</th>');
                $('#fields').append('<option value="' + fs[j].id + '">' + n + '</option>');
                if (AssignmentsView.isMetric(fs[j].id)) {
                    $('#table-assignments > thead > tr > th').last().hide();
                } else {
                    $('#fields > option').last().prop('selected', true);
                }
            }
            if ($('#fields').multiselected) {
                $('#fields').multiselect('destroy');
            }
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
            $('#fields').multiselected = true;
            for (var i = 0; i < a.length; i++) {
                var s = new Date(a[i].start);
                var e = new Date(a[i].end);
                $('#table-assignments > tbody').append('<tr assignmentid="' + a[i].id + '"><td key="name"><a href="submissionsview?assignment-id=' + a[i].id + '">' + a[i].name + '</a></td><td key="start date">' + s.toLocaleDateString() + '</td><td key="start time">' + s.toLocaleTimeString() + '</td><td key="end date">' + e.toLocaleDateString() + '</td><td key="end time">' + e.toLocaleTimeString() + '</td></tr>');
                var s = '#table-assignments > tbody > tr[assignmentid="' + a[i].id + '"]';
                for (var j = 0; j < fs.length; j++) {
                    if (!AssignmentsView.isMetric(fs[j].id)) {
                        continue;
                    }
                    $(s).append('<td key="' + fs[j].id + '">' + a[i][fs[j].id].value + ' ' + a[i][fs[j].id].unit + '</td>');
                    $(s + ' td').last().hide();
                }
            }
            $("#table-assignments").tablesorter({
                theme: 'bootstrap',
                dateFormat: 'ddmmyyyy'
            });
        });
    },
    isMetric: function(k) {
        return notEqual(k, ['id', 'name', 'start date', 'end date', 'start time', 'end time']);
    }
};

var AssignmentsChart = {
    init: function(tipe, id) {
        $(function() {
            ComparisonChart.init();
            AssignmentsChart.addOptions(tipe, id);
            $('.select-chart').change(function() {
                $('#chart').empty();
                AssignmentsChart.load(tipe, $('#type-dropdown-label').attr(tipe + 'id'), $('#x').val(), $('#y').val());
            });
            $.getJSON(tipe + 's', function(data) {
                if (not(data[tipe + 's'])) {
                    console.log(data);
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
                    $('#chart').empty();
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
        $.getJSON('chart-options?type=assignment&' + tipe + '-id=' + id, function(data) {
            var o = data['options'];
            if (not(o)) {
                console.log(data);
                return;
            }
            for (var i = 0; i < o.length; i++) {
                $('.select-chart').append('<option value="' + o[i].id + '">' + o[i].name + '</option>');
            }
            if (not(x) || !$('#x option[value="' + x + '"]').length) {
                x = o[0].id;
            }
            $('#x').val(x);
            if (not(y) || !$('#y option[value="' + y + '"]').length) {
                y = o[o.length - 1].id;
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
        ComparisonChart.load(params);
    },

};
