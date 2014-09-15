var RenameView = {
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
                RenameView.addClassnames(ps[0].Id);
            });
            $('#project-id').change(function() {
                RenameView.addClassnames($(this).val());
            });
        });
    },

    addClassnames: function(pid) {
        $('#old').empty();
        $.getJSON('basicfileinfos?project-id=' + pid, function(data) {
            if (not(data['fileinfos'])) {
                return;
            }
            var fi = data['fileinfos'];
            for (var i = 0; i < fi.length; i++) {
                $('#old').append('<option fn="' + fi[i].Name + '" pkg="' + fi[i].Package + '">Name: ' + fi[i].Name + ' Package: ' + fi[i].Package + ' Type: ' + fi[i].Type + '</option>');
                $('#file-name-new').val($(this).attr('fn'));
                $('#package-name-new').val($(this).attr('pkg'));
            }
            $('#old').change(function() {
                var s = $('option:selected', this);
                $('#file-name-new').val(s.attr('fn'));
                $('#package-name-new').val(s.attr('pkg'));
                $('#file-name-old').val(s.attr('fn'));
                $('#package-name-old').val(s.attr('pkg'));
            });
            $('#file-name-old').val($('#old option:selected').attr('fn'));
            $('#package-name-old').val($('#old option:selected').attr('pkg'));
            $('#file-name-new').val($('#old option:selected').attr('fn'));
            $('#package-name-new').val($('#old option:selected').attr('pkg'));
        });
    }
}
