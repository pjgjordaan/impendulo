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
    setup: function(tipe) {
        $(function() {
            var params = {
                'type': 'overview',
                'view': tipe
            };
            $.getJSON('table-info', params, function(data) {
                if (not(data['table-info']) || not(data['table-fields'])) {
                    return;
                }
                var info = data['table-info'];
                var fs = data['table-fields'];
                for (var j = 0; j < fs.length; j++) {
                    if (fs[j].id === 'id') {
                        continue;
                    }
                    var n = toTitleCase(fs[j].name);
                    $('#table-overview > thead > tr').append('<th key="' + fs[j].id + '">' + n + '</th>');
                    $('#fields').append('<option value="' + fs[j].id + '">' + n + '</option>');
                    if (Overview.isMetric(fs[j].id, info[0])) {
                        $('#table-overview > thead > tr > th').last().hide();
                    } else {
                        $('#fields > option').last().prop('selected', true);
                    }
                }
                $('#fields').show();
                $('#fields').multiselect({
                    noneSelectedText: 'Add table fields',
                    selectedText: '# table fields selected',
                    click: function(event, ui) {
                        $('[key="' + ui.value + '"]').toggle();
                    },
                    checkAll: function(event, ui) {
                        $('[key]').show();
                    },
                    uncheckAll: function(event, ui) {
                        $('[key]').hide();
                    }
                });
                for (var i = 0; i < info.length; i++) {
                    $('#table-overview > tbody').append('<tr ' + tipe + 'id="' + info[i].id + '"><td key="name"><a href="assignmentsview?' + tipe + '-id=' + info[i].id + '">' + info[i].name + '</a></td></tr>');
                    var s = '#table-overview > tbody > tr[' + tipe + 'id="' + info[i].id + '"]';
                    if (tipe === 'project') {
                        $(s).append('<td class="rowlink-skip" key="description"><a href="#" class="a-info"><span class="glyphicon glyphicon-info-sign"></span><p hidden>' + info[i].description + '</p></a></td>');
                    }
                    for (var j = 0; j < fs.length; j++) {
                        if (!Overview.isMetric(fs[j].id, info[i])) {
                            continue;
                        }
                        $(s).append('<td key="' + fs[j].id + '">' + info[i][fs[j].id].value + ' ' + info[i][fs[j].id].unit + '</td>');
                        $(s + ' td').last().hide();
                    }
                }
                $('#table-overview').tablesorter({
                    theme: 'bootstrap',
                    dateFormat: 'ddmmyyyy'
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
        });
    },
    isMetric: function(k, m) {
        return k !== 'id' && k !== 'name' && k !== 'description' && !not(m[k])
    }
};
