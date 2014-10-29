var AnalysisView = {
    init: function(fn, sid, aid, pid, uid, tipe) {
        AnalysisView.loaded = 0;
        $(function() {
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                AnalysisView.buidDropdown('project', pid, 'assignmentsview', data['projects']);
            });
            $.getJSON('users', function(data) {
                if (not(data['users'])) {
                    return;
                }
                AnalysisView.buidDropdown('user', uid, 'assignmentsview', data['users']);
            });
            $.getJSON('assignments?' + tipe + '-id=' + (tipe === 'user' ? uid : pid), function(data) {
                if (not(data['assignments'])) {
                    return;
                }
                AnalysisView.buidDropdown('assignment', aid, 'submissionsview', data['assignments']);
            });
            $.getJSON('submissions?assignment-id=' + aid, function(data) {
                if (not(data['submissions'])) {
                    return;
                }
                AnalysisView.buidDropdown('submission', sid, 'resultsview', data['submissions']);
            });
            $.getJSON('filenames?submission-id=' + sid, function(data) {
                if (not(data['filenames'])) {
                    return;
                }
                AnalysisView.buidFileDropdown(fn, data['filenames']);
            });

        });
    },
    buidFileDropdown: function(fn, vals) {
        for (var i = 0; i < vals.length; i++) {
            var link = 'resultsview?file=' + vals[i];
            $('#file-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="' + link + '">' + vals[i] + '</a></li>');
            if (fn === vals[i]) {
                $('#file-dropdown-label').attr('filename', fn);
                $('#file-dropdown-label').append('<h4><small>file</small> ' + fn + ' <span class="caret"></span></h4>');
            }
        }
        if ($('#file-dropdown-label').attr('filename') === undefined) {
            $('#file-dropdown-label').append('<h4><small>file</small> None Selected <span class="caret"></span></h4>');
        }
    },


    buidDropdown: function(tipe, id, url, vals) {
        for (var i = 0; i < vals.length; i++) {
            var currentId = tipe === 'user' ? vals[i].Name : vals[i].Id;
            var n = '';
            if (tipe === 'submission') {
                n = vals[i].User + ' ' + new Date(vals[i].Time).toLocaleString();
            } else {
                n = vals[i].Name;
            }
            var link = url + '?' + tipe + '-id=' + currentId;
            $('#' + tipe + '-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="' + link + '">' + n + '</a></li>');
            if (id === currentId) {
                $('#' + tipe + '-dropdown-label').attr(tipe + 'id', id);
                $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> ' + n + ' <span class="caret"></span></h4>');
            }
        }
        if (not($('#' + tipe + '-dropdown-label').attr(tipe + 'id'))) {
            $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> None Selected <span class="caret"></span></h4>');
        }
    },

    showToolCode: function(name, pid, title) {
        var id = 'toolcode-modal';
        var s = '#' + id;
        if ($(s).length > 0) {
            AnalysisView.scrollTo(s);
            return;
        }
        $.getJSON('code?tool-name=' + name + '&project-id=' + pid, function(data) {
            jQuery('<div id="' + id + '" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="' + id + 'label" aria-hidden="true"><div class="modal-dialog modal-lg"><div class="modal-content"><div class="modal-header"><button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button><h4 class="modal-title" id="' + id + 'label">' + title + '</h4></div><div class="modal-body"><script class="brush: java;" type="syntaxhighlighter"><![CDATA[' + data.code + ']]></script></div></div></div></div>').appendTo('body');
            SyntaxHighlighter.defaults['toolbar'] = false;
            SyntaxHighlighter.defaults['class-name'] = 'error';
            SyntaxHighlighter.highlight();
            AnalysisView.scrollTo(s);
            return false;
        });
    },

    addCodeModal: function(dest, resultId, bug, start, end) {
        var id = 'bug-modal';
        var s = '#' + id;
        if ($(s).length === 0) {
            $('<div id="' + id + '" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="' + id + 'label" aria-hidden="true"><div class="modal-dialog modal-lg"><div class="modal-content"><div class="modal-header"><button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button><h4 class="modal-title" id="' + id + 'label"><br><small></small></h4></div><div class="modal-body"></div></div></div></div>').appendTo('body');
        }
        $('#' + dest).click(function() {
            $.getJSON('code?result-id=' + resultId, function(data) {
                var h = 'highlight: [';
                for (var i = start; i < end; i++) {
                    h += i + ',';
                }
                h = h + end + '];'
                var preClass = 'brush: java; ' + h;
                $(s + 'label').html(bug.title + '<br><small>' + bug.content + '</small>');
                $(s + ' .modal-body').html('<script class="' + preClass + '" type="syntaxhighlighter"><![CDATA[' + data.code + ']]></script>');
                SyntaxHighlighter.defaults['toolbar'] = false;
                SyntaxHighlighter.defaults['class-name'] = 'error';
                SyntaxHighlighter.highlight();
                $(s).find('.highlighted').attr('style', 'background-color: #ff7777 !important;');
                AnalysisView.scrollTo(s);
            });
            return false;
        });
    },
    scrollTo: function(s) {
        $(s).modal('show');
        $(s).on('shown.bs.modal', function(e) {
            var p = $(s).find('.highlighted').position();
            if (p !== undefined) {
                $(s).scrollTop(p.top);
            }
        });
    }
}
