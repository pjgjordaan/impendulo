var Status = {
    load: function(dest) {
        $.getJSON('status', function(data) {
            if (not(data['status'])) {
                return;
            }
            var s = data['status'];
            var accordionItem = '<div class="panel panel-default"><div class="panel-heading"><h4 class="panel-title"><a data-toggle="collapse" data-parent="#' + dest + '" href="#{0}">{1} <span class="badge">{2}</span></a></h4></div><div id="{0}" class="panel-collapse collapse"><div class="panel-body">{3}</div></div></div>';
            $('#' + dest).append(accordionItem.format('panel-submissions', 'Submissions', Object.keys(s.Submissions).length, '<div class="panel-group" id="submissions-accordion"></div>'));
            $('#' + dest).append(accordionItem.format('panel-files', 'Files', s.FileCount, ''));
            var subItem = '<div class="panel panel-default"><div class="panel-heading"><h4 class="panel-title"><a data-toggle="collapse" data-parent="#submissions-accordion" href="#{0}">{1}</a></h4></div><div id="{0}" class="panel-collapse collapse"><div class="panel-body"><dl class="dl-horizontal"><dt>Files</dt><dd>{2}</dd><dt>User</dt><dd>{3}</dd><dt>Time</dt><dd>{4}</dd><dt>Project</dt><dd>{5}</dd></dl></div></div></div></div>';
            var pmap = {};
            for (var sid in s.Submissions) {
                Status.addSubmissionInfo(sid, s.Submissions[sid], subItem, 'submissions-accordion');
            }
        });
    },

    addSubmissionInfo: function(sid, info, template, dest) {
        $.getJSON('submissions?id=' + sid, function(sdata) {
            if (not(sdata['submissions'])) {
                return;
            }
            var sub = sdata['submissions'][0];
            var fs = Status.countString('', osize(info.Src), 'Source');
            fs = Status.countString(fs, osize(info.Test), 'Test');
            fs = Status.countString(fs, osize(info.Archive), 'Archive');
            $.getJSON('projects?id=' + sub.ProjectId, function(pdata) {
                if (not(pdata['projects'])) {
                    return;
                }
                var p = pdata['projects'][0].Name;
                $('#' + dest).append(template.format('panel-sub-' + sub.Id, p + ' by ' + sub.User, fs, sub.User, new Date(sub.Time).toLocaleString(), p));
            });
        });
    },
    countString: function(val, size, name) {
        if (size > 0) {
            if (val !== '') {
                val += '\n';
            }
            val += size + ' ' + name;
        }
        return val;
    }

}
