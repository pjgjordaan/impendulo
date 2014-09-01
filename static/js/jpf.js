var JPFView = {
    init: function() {
        $(function() {
            JPFView.addProjects();
            JPFView.addListeners();
            JPFView.addSearches();
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

    addListeners: function() {
        clearMulti('#listeners');
        $.getJSON('jpflisteners', function(data) {
            for (var i in data.listeners) {
                var v = data.listeners[i].Package + '.' + data.listeners[i].Name;
                $('#listeners').append('<option value="' + v + '">' + v + '</option>');
            }
            $('#listeners').multiselect({
                selectedText: "# of # listeners selected",
                noneSelectedText: "Select listeners",
                classes: "multiselect-listeners"
            });
            $('#listeners').multiselected = true;
        });
    },

    addSearches: function() {
        $.getJSON('jpfsearches', function(data) {
            $('#search').empty();
            for (var i in data.searches) {
                var v = data.searches[i].Package + '.' + data.searches[i].Name;
                $('#search').append('<option value="' + v + '">' + v + '</option>');
            }
        });
    }
}
