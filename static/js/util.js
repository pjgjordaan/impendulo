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

var ToolView = {
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
                $('#project-id').change(function() {
                    var pid = $(this).val();
                    ToolView.loadTools(pid);
                    ToolView.loadUsers(pid);
                });
                ToolView.loadTools(ps[0].Id);
                ToolView.loadUsers(ps[0].Id);
            });
        });
    },
    loadTools: function(pid) {
        clearMulti('#tools');
        $.getJSON('tools?project-id=' + pid, function(data) {
            var t = data['tools'];
            for (var i = 0; i < t.length; i++) {
                var n = t[i].replace(':', '\u2192')
                $('#tools').append('<option value="' + t[i] + '">' + n + '</option>');
            }
            $('#tools').multiselect();
            $('#tools').multiselected = true;
        });
    },
    loadUsers: function(pid) {
        clearMulti('#users');
        $.getJSON('usernames?project-id=' + pid, function(data) {
            var u = data['usernames'];
            for (var i = 0; i < u.length; i++) {
                $('#users').append('<option value="' + u[i] + '">' + u[i] + '</option>');
            }
            $('#users').multiselect();
            $('#users').multiselected = true;
        });
    }
}


function clearMulti(id) {
    $(id).multiselect();
    $(id).multiselect('destroy');
    $(id).empty();
}

function not(v) {
    return v === null || v === undefined || v.length === 0 || jQuery.isEmptyObject(v);
}


$(function() {
    $('.tree li:has(ul)').addClass('parent_li').find(' > span').attr('title', 'Collapse this branch');
    $('.tree li.parent_li > span').on('click', function(e) {
        var children = $(this).parent('li.parent_li').find(' > ul > li');
        if (children.is(':visible')) {
            children.hide('fast');
            $(this).attr('title', 'Expand this branch').find(' > i').addClass('icon-plus-sign').removeClass('icon-minus-sign');
        } else {
            children.show('fast');
            $(this).attr('title', 'Collapse this branch').find(' > i').addClass('icon-minus-sign').removeClass('icon-plus-sign');
        }
        e.stopPropagation();
    });
});


function loadSkeletonInfo(id, dest) {
    $.getJSON('skeletoninfo?id=' + id, function(data) {
        var items = data['info'];
        for (var i = 0; i < items.length; i++) {
            $('#' + dest).append('<option value="' + items[i] + '">' + items[i] + '</option>');
        }
        $('#' + dest).multiselect({
            noneSelectedText: 'Select all files which should be ignored by Impendulo',
            selectedText: '# files selected to ignore'
        });
        $('#' + dest).multiselected = true;
    });
}


function setContext(changes) {
    $.post('setcontext', changes, function(data) {
        if (data["error"] !== "none" && !not(data["error"])) {
            console.log(data["error"]);
        }
    });
}

function getY(d) {
    return +d.y;
}

function getTime(d) {
    return +d.time;
}


function getX(d) {
    return +d.x;
}

function trimSpace(s) {
    return replaceAll(s, ' ', '');
}

function replaceAll(str, find, replace) {
    return str.replace(new RegExp(find, 'g'), replace);
}

function endsWith(str, suffix) {
    return str.indexOf(suffix, str.length - suffix.length) !== -1;
};

if (typeof String.prototype.format != 'function') {
    String.prototype.format = function() {
        var args = arguments;
        return this.replace(/{(\d+)}/g, function(match, number) {
            return typeof args[number] != 'undefined' ? args[number] : match;
        });
    };
}

if (typeof String.prototype.startsWith != 'function') {
    String.prototype.startsWith = function(str) {
        return this.slice(0, str.length) == str;
    };
}

if (typeof String.prototype.endsWith != 'function') {
    String.prototype.endsWith = function(str) {
        return this.slice(-str.length) == str;
    };
}

function round(n, p) {
    if (p === undefined) {
        p = 2
    }
    return +Number(n).toFixed(2);
}

function osize(o) {
    var ks = Object.keys(o);
    return not(o) ? 0 : ks.length;
}

function toTitleCase(str) {
    return str.replace(/\w\S*/g, function(txt) {
        return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();
    });
}
