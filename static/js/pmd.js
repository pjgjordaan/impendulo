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
