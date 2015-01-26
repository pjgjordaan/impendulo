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
        var tid = $('#type-dropdown-label').attr(AssignmentsView.tipe + 'id');
        var params = {
            'type': 'assignment',
            'assignment-type': AssignmentsView.tipe,
            'id': tid
        };
        ComparisonTable.load(params, 'assignment', 'submissionsview');
    }
};

var AssignmentsChart = {
    init: function(tipe, id) {
        $(function() {
            $('#granularity').append('<option value="assignment">Assignments</option>');
            $('#granularity').append('<option value="submission">Submissions</option>');
            var params = {
                'type': 'assignment'
            };
            params[tipe + '-id'] = id;
            ComparisonChart.init(params);
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
                    var tid = $(this).attr(tipe + 'id');
                    var params = {};
                    params[tipe + '-id'] = tid;
                    setContext(params);
                    $('#type-dropdown-label').attr(tipe + 'id', tid);
                    $('#type-dropdown-label h4').html('<small>' + tipe + '</small> ' + $(this).html() + ' <span class="caret"></span>');
                    ComparisonChart.params[tipe + '-id'] = tid;
                    ComparisonChart.addOptions();
                });
            });
        });
    },
};
