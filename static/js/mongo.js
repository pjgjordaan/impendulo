var MongoView = {
    init: function() {
        $(function() {
            $.getJSON('databases', function(data) {
                var dbs = data['databases'];
                if (not(dbs)) {
                    return;
                }
                for (var i = 0; i < dbs.length; i++) {
                    $('#db').append('<option value="' + dbs[i] + '">' + dbs[i] + '</option>');
                }
                $('#db').change(function() {
                    MongoView.loadCollections($(this).val());
                });
                MongoView.loadCollections(dbs[0]);
            });
        });

    },
    loadCollections: function(db) {
        clearMulti('#collections');
        $.getJSON('collections?db=' + db, function(data) {
            var c = data['collections'];
            if (not(c)) {
                return;
            }
            for (var i = 0; i < c.length; i++) {
                $('#collections').append('<option value="' + c[i] + '">' + c[i] + '</option>');
            }
            $('#collections').multiselect({
                position: {
                    my: 'left center',
                    at: 'right center'
                },
                noneSelectedText: 'Choose collections to export',
                selectedText: '# collections selected to export'
            });
            $('#collections').multiselected = true;
        });
    }
}
