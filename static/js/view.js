var View = {
    buidDropdown: function(tipe, id, url, vals) {
        for (var i = 0; i < vals.length; i++) {
            var currentId = tipe === 'user' ? vals[i].Name : vals[i].Id;
            var link = url + '?' + tipe + '-id=' + currentId;
            $('#' + tipe + '-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="' + link + '" ' + tipe + 'id="' + id + '">' + vals[i].Name + '</a></li>');
            if (id === currentId) {
                $('#' + tipe + '-dropdown-label').attr(tipe + 'id', id);
                $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> ' + vals[i].Name + ' <span class="caret"></span></h4>');
            }
        }
        if (not($('#' + tipe + '-dropdown-label').attr(tipe + 'id'))) {
            $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> None Selected <span class="caret"></span></h4>');
        }
    }

}
