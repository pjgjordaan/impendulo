var Archive = {
    init: function() {
        $(function() {
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                var ps = data['projects'];
                for (var i = 0; i < ps.length; i++) {
                    $('#project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                }
                Archive.addAssignments(ps[0].Id);
                $('#project-id').change(function() {
                    Archive.addAssignments($(this).val());
                });
            });
            $.getJSON('users', function(data) {
                if (not(data['users'])) {
                    return;
                }
                var u = data['users'];
                for (var i = 0; i < u.length; i++) {
                    $('#user-id').append('<option value="' + u[i].Name + '">' + u[i].Name + '</option>');
                }
            });

        });
    },
    addAssignments: function(pid) {
        $('#assignment-id').empty();
        $('#assignment-id').hide();
        $.getJSON('assignments?project-id=' + pid, function(data) {
            if (not(data['assignments'])) {
                return;
            }
            var a = data['assignments'];
            for (var i = 0; i < a.length; i++) {
                $('#assignment-id').append('<option value="' + a[i].Id + '">' + a[i].Name + '</option>');
            }
            $('#assignment-id').show();
        });
    }
}
