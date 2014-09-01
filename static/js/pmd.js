var PMDView = {
    init: function() {
        $(function() {
            PMDView.addProjects();
            PMDView.addRules();
        });
    },
    addProjects: function() {
        $.getJSON('projects', function(data) {
            if (not(data['projects'])) {
                return;
            }
            var ps = data['projects'];
            for (var i = 0; i < ps.length; i++) {
                $('#project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
            }
        });
    },

    addRules: function() {
        clearMulti('#rules');
        $.getJSON('pmdrules', function(data) {
            if (not(data['rules'])) {
                return;
            }
            var rs = data['rules'];
            for (var i in rs) {
                $('#rules').append('<option description="' + rs[i].Description + '" value="' + rs[i].Id + '">' + rs[i].Name + '</option>');
            }
            $('#rules').multiselect({
                selectedText: "# of # rules selected",
                noneSelectedText: "Select rules",
                classes: "multiselect-rules"
            });
            $('.multiselect-rules .ui-multiselect-checkboxes li').tooltip({
                title: function() {
                    return $('option[value="' + $(this).find('input').val() + '"]').attr('description');
                },
                placement: 'left',
                container: 'body'
            });
            $('#rules').multiselected = true;
        });
    }
}
