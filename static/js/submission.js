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
    init: function(aid, pid, uid, tipe) {
        $('#div-granularity').hide();
        var params = {
            'type': 'submission',
            'assignment-id': aid
        };
        if (tipe === 'project') {
            params['project-id'] = pid;
        } else if (tipe === 'user') {
            params['user-id'] = uid;
        }
        ComparisonChart.init(params);
        $(function() {
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
            var currentId = $(this).attr('currentid');
            var currentName = $(this).html();
            var params = {};
            params['assignment-id'] = currentId;
            setContext(params);
            $('#assignment-dropdown-label').attr('currentid', currentId);
            $('#assignment-dropdown-label h4').html('<small>assignment</small> ' + currentName + ' <span class="caret"></span>');
            ComparisonChart.params['assignment-id'] = currentId;
            ComparisonChart.addOptions();
        });
    },

};

var SubmissionsView = {
    tipe: '',
    init: function(aid, pid, uid, tipe) {
        SubmissionsView.tipe = tipe;
        $(function() {
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
        var aid = $('#assignment-dropdown-label').attr('assignmentid');
        var params = {
            'type': 'submission',
            'assignment-id': aid
        }
        params[SubmissionsView.tipe + '-id'] = $('#' + SubmissionsView.tipe + '-dropdown-label').attr(SubmissionsView.tipe + 'id');
        console.log(params);
        ComparisonTable.load(params, 'submission', 'filesview');
    }
}
