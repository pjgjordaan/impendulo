var FilesView = {
    init: function(sid, aid, pid, uid, tipe) {
        $(function() {
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    console.log(data);
                    return;
                }
                FilesView.buidDropdown('project', pid, 'assignmentsview', data['projects']);
                $.getJSON('users', function(data) {
                    if (not(data['users'])) {
                        console.log(data);
                        return;
                    }
                    FilesView.buidDropdown('user', uid, 'assignmentsview', data['users']);
                    var id = tipe === 'user' ? uid : pid;
                    $.getJSON('assignments?' + tipe + '-id=' + id, function(data) {
                        if (not(data['assignments'])) {
                            console.log(data);
                            return;
                        }
                        FilesView.buidDropdown('assignment', aid, 'submissionsview', data['assignments']);
                        $.getJSON('submissions?assignment-id=' + aid, function(data) {
                            if (not(data['submissions'])) {
                                console.log(data);
                                return;
                            }
                            FilesView.subDropdown(sid, data['submissions']);
                            FilesView.load(sid);
                        });
                    });
                });
            });
        });
    },

    subDropdown: function(id, vals) {
        for (var i = 0; i < vals.length; i++) {
            $('#submission-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="#" submissionid="' + vals[i].Id + '">' + vals[i].User + ' at  ' + new Date(vals[i].Time).toLocaleString() + '</a></li>');
            if (id === vals[i].Id) {
                $('#submission-dropdown-label').attr('submissionid', id);
                $('#submission-dropdown-label').append('<h4><small>submission</small> ' + vals[i].User + ' at  ' + new Date(vals[i].Time).toLocaleString() + ' <span class="caret"></span></h4>');
            }
        }
        $('#submission-dropdown ul.dropdown-menu a').on('click', function() {
            $('#file-list').empty();
            var currentId = $(this).attr('submissionid');
            var currentName = $(this).html();
            var params = {
                'submission-id': currentId
            };
            setContext(params);
            $('#submission-dropdown-label').attr('submissionid', currentId);
            $('#submission-dropdown-label h4').html('<small>submission</small> ' + currentName + ' <span class="caret"></span>');
            FilesView.load(currentId);
        });
    },

    buidDropdown: function(tipe, id, url, vals) {
        for (var i = 0; i < vals.length; i++) {
            var currentId = tipe === 'user' ? vals[i].Name : vals[i].Id;
            var link = url + '?' + tipe + '-id=' + currentId;
            $('#' + tipe + '-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="' + link + '">' + vals[i].Name + '</a></li>');
            if (id === currentId) {
                $('#' + tipe + '-dropdown-label').attr(tipe + 'id', id);
                $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> ' + vals[i].Name + ' <span class="caret"></span></h4>');
            }
        }
        if (not($('#' + tipe + '-dropdown-label').attr(tipe + 'id'))) {
            $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> None Selected <span class="caret"></span></h4>');
        }
    },

    load: function(sid) {
        var params = {
            'type': 'file',
            'submission-id': sid
        }
        ComparisonTable.load(params, 'file', 'resultsview');
    }
}
