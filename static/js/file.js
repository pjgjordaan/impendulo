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
        $('#table-files > tbody').empty();
        var params = {
            'type': 'file',
            'submission-id': sid
        }
        $.getJSON('table', params, function(data) {
            if (not(data['table-data']) || not(data['table-fields']) || not(data['table-metrics'])) {
                console.log(data);
                return;
            }
            var td = data['table-data'];
            var tf = data['table-fields'];
            var tm = data['table-metrics'];
            for (var j = 1; j < tf.length; j++) {
                var n = toTitleCase(tf[j].name);
                $('#table-files > thead > tr').append('<th key="' + tf[j].id + '">' + n + '</th>');
                $('#fields').append('<option value="' + tf[j].id + '">' + n + '</option>');
                $('#fields > option').last().prop('selected', true);
            }
            for (var j = 0; j < tm.length; j++) {
                var n = toTitleCase(tm[j].name);
                $('#table-files > thead > tr').append('<th key="' + tm[j].id + '">' + n + '</th>');
                $('#fields').append('<option value="' + tm[j].id + '">' + n + '</option>');
                $('#table-files > thead > tr > th').last().hide();
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
            for (var i = 0; i < td.length; i++) {
                $('#table-files > tbody').append('<tr file="' + td[i].id + '"></tr>')
                var s = '#table-files > tbody > tr[file="' + td[i].id + '"]';
                for (var j = 1; j < tf.length; j++) {
                    if (j === 1) {
                        $(s).append('<td key="' + tf[j].id + '"><a href="resultsview?file=' + td[i].id + '">' + td[i][tf[j].id] + '</a></td>');
                    } else {
                        $(s).append('<td key="' + tf[j].id + '">' + td[i][tf[j].id] + '</td>');
                    }
                }
                for (var j = 0; j < tm.length; j++) {
                    var o = td[i][tm[j].id];
                    var unit = '';
                    var value = 'N/A';
                    if (!not(o) && o.value !== -1) {
                        value = o.value;
                        unit = o.unit;
                    }
                    $(s).append('<td key="' + tm[j].id + '">' + value + ' ' + unit + '</td>');
                    $(s + ' td').last().hide();
                }
            }
            $('#table-files').tablesorter({
                theme: 'bootstrap',
                dateFormat: 'ddmmyyyy'
            });
        });
    }
}
