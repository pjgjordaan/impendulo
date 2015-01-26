var ComparisonTable = {
    load: function(params, tipe, url) {
        $('#fields').empty();
        $('#table-comparison > tbody').empty();
        $.getJSON('table', params, function(data) {
            if (not(data['table-data']) || not(data['table-fields']) || not(data['table-metrics'])) {
                console.log('could not load table data', data);
                return;
            }
            var td = data['table-data'];
            var tf = data['table-fields'];
            var tm = data['table-metrics'];
            for (var j = 1; j < tf.length; j++) {
                var n = toTitleCase(tf[j].name);
                $('#table-comparison > thead > tr').append('<th key="' + tf[j].id + '">' + n + '</th>');
                $('#fields').append('<option value="' + tf[j].id + '">' + n + '</option>');
                $('#fields > option').last().prop('selected', true);
            }
            for (var j = 0; j < tm.length; j++) {
                var n = toTitleCase(tm[j].name);
                $('#table-comparison > thead > tr').append('<th key="' + tm[j].id + '">' + n + '</th>');
                $('#fields').append('<option value="' + tm[j].id + '">' + n + '</option>');
                $('#table-comparison > thead > tr > th').last().hide();
            }
            $('#fields').show();
            $('#fields').multiselect({
                noneSelectedText: 'Add table fields',
                selectedText: '# table fields selected',
                click: function(event, ui) {
                    $('[key="' + ui.value + '"]').toggle();
                    if ($('[key="' + ui.value + '"]').is(":visible")) {
                        $('[key="' + ui.value + '"]').each(function() {
                            $(this).appendTo($(this).parent());
                        });
                    }
                },
                checkAll: function(event, ui) {
                    $('[key]').each(function() {
                        if (!$(this).is(":visible")) {
                            $(this).appendTo($(this).parent());
                        }
                    });
                    $('[key]').show();
                },
                uncheckAll: function(event, ui) {
                    $('[key]').hide();
                }
            });
            for (var i = 0; i < td.length; i++) {
                $('#table-comparison > tbody').append('<tr ' + tipe + 'id="' + td[i].id + '"></tr>')
                var s = '#table-comparison > tbody > tr[' + tipe + 'id="' + td[i].id + '"]';
                for (var j = 1; j < tf.length; j++) {
                    if (j === 1) {
                        $(s).append('<td key="' + tf[j].id + '"><a href="' + url + '?' + tipe + '-id=' + td[i].id + '">' + td[i][tf[j].id] + '</a></td></tr>');
                    } else if (tipe === 'project' && tf[j].id === 'description') {
                        $(s).append('<td class="rowlink-skip" key="description"><a href="#" class="a-info"><span class="glyphicon glyphicon-info-sign"></span><p hidden>' + td[i][tf[j].id] + '</p></a></td>');
                    } else {
                        $(s).append('<td key="' + tf[j].id + '">' + td[i][tf[j].id] + '</td>');
                    }
                }
                for (var j = 0; j < tm.length; j++) {
                    var o = td[i][tm[j].id];
                    var unit = '';
                    var value = 'N/A';
                    if (!not(o) && o.value !== -1) {
                        value = o.value;
                        unit = o.unit;
                    }
                    $(s).append('<td key="' + tm[j].id + '">' + value + ' ' + unit + '</td>');
                    $(s + ' td').last().hide();
                }
            }
            $('#table-comparison').tablesorter({
                theme: 'bootstrap',
                dateFormat: 'ddmmyyyy',
                textExtraction: tableSortExtraction
            });
            if (tipe === 'project') {
                $('.a-info').popover({
                    content: function() {
                        var d = $(this).find('p').html();
                        return d === '' ? 'No description' : d;
                    }
                });
            }
        });
    }
};
