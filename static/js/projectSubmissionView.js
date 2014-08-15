var ProjectSubmissionsView = {
    init: function() {
        $(function() {
            ProjectSubmissionsView.addPickers();
            $('#button-filter').on('click', ProjectSubmissionsView.load);
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                var ps = data['projects'];
                for (var i = 0; i < ps.length; i++) {
                    $('#project-list').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="getsubmissions?project-id=' + ps[i].Id + '">' + ps[i].Name + '</a></li>');
                }
            });
            ProjectSubmissionsView.load();
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
        ProjectSubmissionsView.pickerButton('start');
        ProjectSubmissionsView.pickerButton('end');
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
        $('#table-submissions > tbody').empty();
        var pid = $('#project-dropdown-label').attr('projectid');
        var params = {
            'counts': true,
            'project-id': pid,
            'time-start': ProjectSubmissionsView.time('#datetimepicker-start'),
            'time-end': ProjectSubmissionsView.time('#datetimepicker-end')
        }
        $.getJSON('submissions', params, function(data) {
            if (not(data['submissions']) || not(data['counts'])) {
                return;
            }
            var s = data['submissions'];
            var c = data['counts'];
            for (var i = 0; i < s.length; i++) {
                var d = new Date(s[i].Time);
                $('#table-submissions > tbody').append('<tr submissionid="' + s[i].Id + '"><td><a href="getfiles?submission-id=' + s[i].Id + '">' + s[i].User + '</a></td><td>' + d.toLocaleDateString() + '</td><td>' + d.toLocaleTimeString() + '</td><td>' + c[s[i].Id]['source'] + '</td><td>' + c[s[i].Id]['launch'] + '</td><td>' + c[s[i].Id]['test'] + '</td><td>' + c[s[i].Id]['testcases'] + '</td><td>' + c[s[i].Id]['passed'] + ' %</td></tr>');
            }
            $("#table-submissions").tablesorter({
                theme: 'bootstrap',
                dateFormat: 'ddmmyyyy'
            });
        });
    }
}
