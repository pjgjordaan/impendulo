var Overview = {
    setup: function(tipe) {
        $(function() {
            $.getJSON('typecounts?view=' + tipe, function(data) {
                console.log(data);
                if (not(data['typecounts'])) {
                    return;
                }
                var tcs = data['typecounts'];
                var cs = data['categories'];
                for (var i = 0; i < cs.length; i++) {
                    $('tr.info').append('<th>' + cs[i].toTitleCase() + '</th>');
                }
                for (var i = 0; i < tcs.length; i++) {
                    var tr = '<tr>';
                    if (tipe === 'project') {
                        tr += '<td><a href="getassignments?project-id=' + tcs[i].id + '">' + tcs[i].name + '</a></td><td class="rowlink-skip"><a href="#" class="a-info"><span class="glyphicon glyphicon-info-sign"></span><p hidden>' + tcs[i].description + '</p></a></td><td>' + new Date(tcs[i].time).toLocaleString() + '</td><td>' + tcs[i].lang + '</td>';
                    } else {
                        tr += '<td><a href="getsubmissions?user-id=' + tcs[i].name + '">' + tcs[i].name + '</a></td>';
                    }
                    for (var j = 0; j < cs.length; j++) {
                        tr += '<td>' + tcs[i][cs[j]] + '</td>';
                    }
                    $('tbody').append(tr);
                }
                $('#table-' + tipe).tablesorter({
                    theme: 'bootstrap'
                });
                if (tipe === 'project') {
                    $('.a-info').popover({
                        content: function() {
                            var d = $(this).find('p').html();
                            return d === '' ? 'No description' : d;
                        }
                    })
                }
            });

        });
    }
}
