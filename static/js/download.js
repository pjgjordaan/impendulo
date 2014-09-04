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

var SkeletonDownload = {
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
                SkeletonDownload.addSkeletons(ps[0].Id);
                $('#project-id').change(function() {
                    SkeletonDownload.addSkeletons($(this).val());
                });
            });
        });
    },
    addSkeletons: function() {
        var id = $('#project-id').val();
        $.getJSON('skeletons?project-id=' + id, function(data) {
            $('#skeleton-id').empty();
            $('#skeleton-id').hide();
            if (not(data['skeletons'])) {
                return;
            }
            $('#skeleton-id').show();
            var sk = data['skeletons'];
            for (var i = 0; i < sk.length; i++) {
                $('#skeleton-id').append('<option value="' + sk[i].Id + '">' + sk[i].Name + '</option>');
            }
        });
    }
}

var TestDownload = {
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
                TestDownload.addTests(ps[0].Id);
                $('#project-id').change(function() {
                    TestDownload.addTests($(this).val());
                });
            });
        });
    },

    addTests: function() {
        var id = $('#project-id').val();
        $.getJSON('tests?project-id=' + id, function(data) {
            $('#test-id').empty();
            $('#test-id').hide();
            if (not(data['tests'])) {
                return;
            }
            $('#test-id').show();
            var ts = data['tests'];
            for (var i = 0; i < ts.length; i++) {
                $('#test-id').append('<option value="' + ts[i].Id + '">' + ts[i].Name + ' \u2192 ' + new Date(ts[i].Time).toLocaleString() + '</option>');
            }
        });
    }
}
