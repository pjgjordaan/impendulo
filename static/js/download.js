var SkeletonDowload = {
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
                SkeletonDowload.addSkeletons(ps[0].Id);
                $('#project-id').change(function() {
                    SkeletonDowload.addSkeletons($(this).val());
                });
            });
        });
    },
    addSkeletons: function() {
        var id = $('#project-id').val();
        $.getJSON('skeletons?project-id=' + id, function(data) {
            $('#skeleton-id').empty();
            $('#skeleton-id').hide();
            if (not(data['skeletons'])) {
                return;
            }
            $('#skeleton-id').show();
            var sk = data['skeletons'];
            for (var i = 0; i < sk.length; i++) {
                $('#skeleton-id').append('<option value="' + sk[i].Id + '">' + sk[i].Name + '</option>');
            }
        });
    }
}

var TestDownload = {
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
                TestDownload.addTests(ps[0].Id);
                $('#project-id').change(function() {
                    TestDownload.addTests($(this).val());
                });
            });
        });
    },

    addTests: function() {
        var id = $('#project-id').val();
        $.getJSON('tests?project-id=' + id, function(data) {
            $('#test-id').empty();
            $('#test-id').hide();
            if (not(data['tests'])) {
                return;
            }
            $('#test-id').show();
            var ts = data['tests'];
            for (var i = 0; i < ts.length; i++) {
                $('#test-id').append('<option value="' + ts[i].Id + '">' + ts[i].Name + ' \u2192 ' + new Date(ts[i].Time).toLocaleString() + '</option>');
            }
        });
    }
}
