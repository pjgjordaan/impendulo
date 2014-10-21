//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

var OverviewChart = {
    tipe: '',
    init: function(tipe) {
        $(function() {
            ComparisonChart.init();
            OverviewChart.tipe = tipe;
            OverviewChart.addOptions();
            $('.select-chart').change(function() {
                $('#chart').empty();
                OverviewChart.load($('#x').val(), $('#y').val());
            });
            OverviewChart.addOptions();
        });
    },

    addOptions: function() {
        var x = $('#x').val();
        var y = $('#y').val();
        $.getJSON('chart-options', function(data) {
            var o = data['options'];
            if (not(o)) {
                return;
            }
            for (var i = 0; i < o.length; i++) {
                $('.select-chart').append('<option value="' + o[i].Id + '">' + o[i].Name + '</option>');
            }
            if (x === undefined || x === null || !$('#x option[value="' + x + '"]').length) {
                x = o[0].Id;
            }
            $('#x').val(x);
            if (y === undefined || y === null || !$('#y option[value="' + y + '"]').length) {
                y = o[o.length - 1].Id;
            }
            $('#y').val(y);
            OverviewChart.load(x, y);
        });
    },

    load: function(x, y) {
        var params = {
            'type': 'overview',
            'view': OverviewChart.tipe,
            'x': x,
            'y': y
        };
        ComparisonChart.load(params);
    },

};

var Overview = {
    setup: function(tipe, tipeNames) {
        $(function() {
            $.getJSON('typecounts?view=' + tipe, function(data) {
                if (not(data['typecounts'])) {
                    return;
                }
                var tcs = data['typecounts'];
                var cs = data['categories'];
                for (var i = 0; i < tipeNames.length; i++) {
                    $('tr.info').append('<th>' + tipeNames[i] + '</th>');
                }
                for (var i = 0; i < cs.length; i++) {
                    $('tr.info').append('<th>' + cs[i].toTitleCase() + '</th>');
                }
                for (var i = 0; i < tcs.length; i++) {
                    var tr = '<tr>';
                    if (tipe === 'project') {
                        tr += '<td><a href="assignmentsview?project-id=' + tcs[i].id + '">' + tcs[i].name + '</a></td><td class="rowlink-skip"><a href="#" class="a-info"><span class="glyphicon glyphicon-info-sign"></span><p hidden>' + tcs[i].description + '</p></a></td><td>' + new Date(tcs[i].time).toLocaleString() + '</td><td>' + tcs[i].lang + '</td>';
                    } else {
                        tr += '<td><a href="assignmentsview?user-id=' + tcs[i].name + '">' + tcs[i].name + '</a></td>';
                    }
                    for (var j = 0; j < cs.length; j++) {
                        tr += '<td>' + tcs[i][cs[j]] + '</td>';
                    }
                    $('tbody').append(tr);
                }
                $('#table-overview').tablesorter({
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
};
