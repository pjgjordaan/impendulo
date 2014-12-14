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
    init: function(tipe) {
        $(function() {
            $('#granularity').append('<option value="overview">' + toTitleCase(tipe) + '</option>');
            $('#granularity').append('<option value="assignment">Assignments</option>');
            $('#granularity').append('<option value="submission">Submissions</option>');
            var params = {
                'type': 'overview',
                'view': tipe
            };
            ComparisonChart.init(params);
        });
    }
};

var Overview = {
    setup: function(tipe) {
        $(function() {
            var params = {
                'type': 'overview',
                'view': tipe
            };
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
                    $('#table-overview > thead > tr').append('<th key="' + tf[j].id + '">' + n + '</th>');
                    $('#fields').append('<option value="' + tf[j].id + '">' + n + '</option>');
                    $('#fields > option').last().prop('selected', true);
                }
                for (var j = 0; j < tm.length; j++) {
                    var n = toTitleCase(tm[j].name);
                    $('#table-overview > thead > tr').append('<th key="' + tm[j].id + '">' + n + '</th>');
                    $('#fields').append('<option value="' + tm[j].id + '">' + n + '</option>');
                    $('#table-overview > thead > tr > th').last().hide();
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
                    $('#table-overview > tbody').append('<tr ' + tipe + 'id="' + td[i].id + '"></tr>')
                    var s = '#table-overview > tbody > tr[' + tipe + 'id="' + td[i].id + '"]';
                    for (var j = 1; j < tf.length; j++) {
                        if (j === 1) {
                            $(s).append('<td key="' + tf[j].id + '"><a href="assignmentsview?' + tipe + '-id=' + td[i].id + '">' + td[i][tf[j].id] + '</a></td></tr>');
                        } else if (tf[j].id === 'description') {
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
                $('#table-overview').tablesorter({
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
        });
    }
};
