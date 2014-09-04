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
