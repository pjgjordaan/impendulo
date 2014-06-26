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

var DeleteView = {
    init: function() {
        $(function() {
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                var ps = data['projects'];
                for (var i = 0; i < ps.length; i++) {
                    $('#tp-project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                    $('#ts-project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                    $('#tr-project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                    $('#tsk-project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                }
                $('#tp-project-id').multiselect({
                    selectedText: '# of # projects selected',
                    noneSelectedText: 'Delete projects'
                });
                $('#tp-project-id').multiselected = true;
                DeleteView.loadSubmissions('#ts-submission-id', ps[0].Id, 'Delete submissions');
                $('#ts-project-id').change(function() {
                    DeleteView.loadSubmissions('#ts-submission-id', $(this).val(), 'Delete submissions');
                });
                DeleteView.loadSubmissions('#tr-submission-id', ps[0].Id, 'Delete submission results');
                $('#tr-project-id').change(function() {
                    DeleteView.loadSubmissions('#tr-submission-id', $(this).val(), 'Delete submission results');
                });
                DeleteView.loadSkeletons(ps[0].Id);
                $('#tsk-project-id').change(function() {
                    DeleteView.loadSkeletons($(this).val());
                });
            });
            $.getJSON('usernames', function(data) {
                if (not(data['usernames'])) {
                    return;
                }
                var us = data['usernames'];
                for (var i = 0; i < us.length; i++) {
                    $('#tu-user-id').append('<option value="' + us[i] + '">' + us[i] + '</option>');
                }
                $('#tu-user-id').multiselect({
                    selectedText: '# of # users selected',
                    noneSelectedText: 'Delete users'
                });
                $('#tu-user-id').multiselected = true;
            });
        });
    },

    loadSkeletons: function(pid) {
        var id = '#tsk-skeleton-id';
        clearMulti(id);
        $(id).hide();
        $.getJSON('skeletons?project-id=' + pid, function(data) {
            if (not(data['skeletons'])) {
                return;
            }
            $(id).show();
            var sk = data['skeletons'];
            for (var i = 0; i < sk.length; i++) {
                $(id).append('<option value="' + sk[i].Id + '">' + sk[i].Name + '</option>');
            }
            $(id).multiselect({
                selectedText: '# of # skeletons selected',
                noneSelectedText: 'Delete skeletons'
            });
            $(id).multiselected = true;
        });
    },

    loadSubmissions: function(id, pid, desc) {
        clearMulti(id);
        $(id).hide();
        $.getJSON('submissions?project-id=' + pid, function(data) {
            if (not(data['submissions'])) {
                return;
            }
            $(id).show();
            var ss = data['submissions'];
            for (var i = 0; i < ss.length; i++) {
                DeleteView.loadSubmission(id, ss[i], desc, ss.length);
            }
        });
    },

    loadSubmission: function(id, sub, desc, num) {
        $.getJSON('counts?submission-id=' + sub.Id, function(data) {
            if (not(data['counts'])) {
                return;
            }
            var c = data['counts'];
            var l = c['launch'];
            var s = c['source'];
            var t = c['test'];
            $(id).append('<option date="' + new Date(sub.Time).toLocaleString() + '" source="' + s + '" launch="' + l + '" test="' + t + '" value="' + sub.Id + '">' + sub.User + '</option>');
            if ($(id).children().length === num) {
                $(id).multiselect({
                    selectedText: '# of # submissions selected',
                    noneSelectedText: desc,
                    classes: 'multiselect-submissions'
                });
                $('.multiselect-submissions .ui-multiselect-checkboxes li').tooltip({
                    title: function() {
                        var sl = 'option[value="' + $(this).find('input').val() + '"]';
                        var d = $(sl).attr('date');
                        var sc = $(sl).attr('source');
                        var lc = $(sl).attr('launch');
                        var tc = $(sl).attr('test');
                        return 'Date: ' + d + '\nSource Files: ' + sc + '\nLaunches: ' + lc + '\nTests: ' + tc;
                    },
                    placement: 'left',
                    container: 'body'
                });
                $(id).multiselected = true;
            }
        });
    }

}

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

var SkeletonDowload = {
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
                SkeletonDowload.addSkeletons(ps[0].Id);
                $('#project-id').change(function() {
                    SkeletonDowload.addSkeletons($(this).val());
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

var TestDowload = {
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
                TestDowload.addTests(ps[0].Id);
                $('#project-id').change(function() {
                    TestDowload.addTests($(this).val());
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
    return v === null || v === undefined || v.length === 0;
}

var Analysis = {
    showToolCode: function(name, pid, title) {
        var id = 'toolcode-modal';
        var s = '#' + id;
        if ($(s).length > 0) {
            $(s).modal('show');
            $(s).on('shown.bs.modal', function(e) {
                line.scrollIntoView();
            });
            return;
        }
        $.getJSON('code?tool-name=' + name + '&project-id=' + pid, function(data) {
            jQuery('<div id="' + id + '" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="' + id + 'label" aria-hidden="true"><div class="modal-dialog"><div class="modal-content"><div class="modal-header"><button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button><h4 class="modal-title" id="' + id + 'label">' + title + '</h4></div><div class="modal-body"><script class="brush: java;" type="syntaxhighlighter"><![CDATA[' + data.code + ']]></script></div></div></div></div>').appendTo('body');
            SyntaxHighlighter.defaults['toolbar'] = false;
            SyntaxHighlighter.defaults['class-name'] = 'error';
            SyntaxHighlighter.highlight();
            $(s).modal('show');
            $(s).on('shown.bs.modal', function(e) {
                $(s).animate({
                    scrollTop: $(s).offset()
                });
            });
            return false;
        });
    },

    addCodeModal: function(dest, resultId, bug, start, end) {
        $('#' + dest).click(function() {
            var id = 'bug-modal';
            var s = '#' + id;
            if ($(s).length > 0) {
                $(s).modal('show');
                $(s).on('shown.bs.modal', function(e) {
                    line.scrollIntoView();
                });
                return false;
            }
            $.getJSON('code?result-id=' + resultId, function(data) {
                var h = 'highlight: [';
                for (var i = start; i < end; i++) {
                    h += i + ',';
                }
                h = h + end + '];'
                var preClass = 'brush: java; ' + h;
                jQuery('<div id="' + id + '" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="' + id + 'label" aria-hidden="true"><div class="modal-dialog"><div class="modal-content"><div class="modal-header"><button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button><h4 class="modal-title" id="' + id + 'label">' + bug.title + '<br><small>' + bug.content + '</small></h4></div><div class="modal-body"><script class="' + preClass + '" type="syntaxhighlighter"><![CDATA[' + data.code + ']]></script></div></div></div></div>').appendTo('body');
                SyntaxHighlighter.defaults['toolbar'] = false;
                SyntaxHighlighter.defaults['class-name'] = 'error';
                SyntaxHighlighter.highlight();
                $(s).find('.highlighted').attr('style', 'background-color: #ff7777 !important;');
                $(s).modal('show');
                $(s).on('shown.bs.modal', function(e) {
                    var offset = $(s).find('.highlighted').offset();
                    var offsetParent = $(s).offset();
                    $(s).animate({
                        scrollTop: offset.top - offsetParent.top
                    });
                });
            });
            return false;
        });
    }
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


function loadCollections(dbList, collectionList) {
    var url = 'collections?db=' + $('#' + dbList).val();
    $.getJSON(url, function(data) {
        clearMulti('#' + collectionList);
        var items = data['collections'];
        for (var i = 0; i < items.length; i++) {
            $('#' + collectionList).append('<option value="' + items[i] + '">' + items[i] + '</option>');
        }
        $('#' + collectionList).multiselect({
            noneSelectedText: 'Choose collections to export',
            selectedText: '# collections selected to export'
        });
        $('#' + collectionList).multiselected = true;
    });
}

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

function addProjects(dest) {
    $.getJSON('projects', function(data) {
        if (not(data['projects'])) {
            return;
        }
        var ps = data['projects'];
        for (var i = 0; i < ps.length; i++) {
            $('#' + dest).append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
        }
        addFilenames('file-name-old', ps[0].Id);
    });
}

function addFilenames(dest, pid) {
    $('#' + dest).empty();
    $.getJSON('filenames?project-id=' + pid, function(data) {
        if (not(data['filenames'])) {
            return;
        }
        var ns = data['filenames'];
        for (var i = 0; i < ns.length; i++) {
            $('#' + dest).append('<option value="' + ns[i] + '">' + ns[i] + '</option>');
        }
    });
}

function loadStatus(dest) {
    $.getJSON('status', function(data) {
        if (not(data['status'])) {
            return;
        }
        var s = data['status'];
        var accordionItem = '<div class="panel panel-default"><div class="panel-heading"><h4 class="panel-title"><a data-toggle="collapse" data-parent="#' + dest + '" href="#{0}">{1} <span class="badge">{2}</span></a></h4></div><div id="{0}" class="panel-collapse collapse"><div class="panel-body">{3}</div></div></div>';
        $('#' + dest).append(accordionItem.format('panel-submissions', 'Submissions', Object.keys(s.Submissions).length, '<div class="panel-group" id="submissions-accordion"></div>'));
        $('#' + dest).append(accordionItem.format('panel-files', 'Files', s.FileCount, ''));
        var subItem = '<div class="panel panel-default"><div class="panel-heading"><h4 class="panel-title"><a data-toggle="collapse" data-parent="#submissions-accordion" href="#{0}">{1}</a></h4></div><div id="{0}" class="panel-collapse collapse"><div class="panel-body"><dl class="dl-horizontal"><dt>Files</dt><dd>{2}</dd><dt>User</dt><dd>{3}</dd><dt>Time</dt><dd>{4}</dd><dt>Project</dt><dd>{5}</dd></dl></div></div></div></div>';
        var pmap = {};
        for (var sid in s.Submissions) {
            var fc = Object.keys(s.Submissions[sid]).length;
            addSubmissionInfo(sid, fc, subItem, 'submissions-accordion');
        }
    });
}

function addSubmissionInfo(sid, fc, template, dest) {
    $.getJSON('submissions?id=' + sid, function(sdata) {
        if (not(sdata['submissions'])) {
            return;
        }
        var sub = sdata['submissions'][0];
        $.getJSON('projects?id=' + sub.ProjectId, function(pdata) {
            if (not(pdata['projects'])) {
                return;
            }
            var p = pdata['projects'][0].Name;
            $('#' + dest).append(template.format('panel-sub-' + sub.Id, p + ' by ' + sub.User, fc, sub.User, new Date(sub.Time).toLocaleString(), p));
        });
    });
}

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

var EditView = {
    init: function() {
        $(function() {
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                var ps = data['projects'];
                for (var i = 0; i < ps.length; i++) {
                    $('#projects ul').append('<li><a href="#" projectid="' + ps[i].Id + '">' + ps[i].Name + '</a></li>');
                }
                EditView.loadproject(ps[0].Id);
                $('#projects a').click(function() {
                    EditView.loadproject($(this).attr('projectid'));
                    return true;
                });
            });

            $.getJSON('usernames', function(data) {
                if (not(data['usernames'])) {
                    return;
                }
                var us = data['usernames'];
                for (var i = 0; i < us.length; i++) {
                    $('#users ul').append('<li><a href="#" user="' + us[i] + '">' + us[i] + '</a></li>');
                }
                $('#users a').click(function() {
                    EditView.hideEditing();
                    $('#user-panel').addClass('in');
                    EditView.loaduser($(this).attr('user'));
                    return true;
                });
            });
        });
    },

    loadproject: function(id) {
        $('.project-input').empty();
        $('.project-input').val('');
        EditView.hideEditing();
        $('#project-panel').addClass('in');
        $.getJSON('projects?id=' + id, function(data) {
            if (not(data['projects'])) {
                return;
            }
            var p = data['projects'][0];
            $('#project-id').val(p.Id);
            $('#project-name').val(p.Name);
            $('#project-description').val(p.Description);
            $.getJSON('usernames', function(udata) {
                var users = udata['usernames'];
                if (not(users)) {
                    return;
                }
                for (var i = 0; i < users.length; i++) {
                    if (users[i] === p.User) {
                        $('#project-user').append('<option value="' + users[i] + '" selected>' + users[i] + '</option>');
                    } else {
                        $('#project-user').append('<option value="' + users[i] + '">' + users[i] + '</option>');
                    }
                }
            });
            $.getJSON('langs', function(ldata) {
                var langs = ldata['langs'];
                if (not(langs)) {
                    return;
                }
                for (var i = 0; i < langs.length; i++) {
                    if (langs[i] === p.Lang) {
                        $('#project-lang').append('<option value="' + langs[i] + '" selected>' + langs[i] + '</option>');
                    } else {
                        $('#project-lang').append('<option value="' + langs[i] + '">' + langs[i] + '</option>');
                    }
                }
            });
            $.getJSON('submissions?project-id=' + p.Id, function(sdata) {
                $('#submissions > ul').empty();
                var subs = sdata['submissions'];
                if (not(subs)) {
                    $('#files > ul').empty();
                    $('.file-input').empty();
                    $('.submission-input').empty();
                    $('.file-input').val('');
                    $('.submission-input').val('');
                    return;
                }
                for (var i = 0; i < subs.length; i++) {
                    $('#submissions > ul').append('<li><a href="#" subid="' + subs[i].Id + '">' + subs[i].User + ' ' + new Date(subs[i].Time).toLocaleString() + '</a></li>');
                }
                $('#submissions a').click(function() {
                    EditView.loadsubmission($(this).attr('subid'));
                    EditView.hideEditing();
                    $('#submission-panel').addClass('in');
                    return true;
                });
                EditView.loadsubmission(subs[0].Id);
            });
            $.getJSON('tests?project-id=' + p.Id, function(tdata) {
                $('#tests > ul').empty();
                var tests = tdata['tests'];
                if (not(tests)) {
                    $('.test-input').empty();
                    $('.test-input').val('');
                    return;
                }
                for (var i = 0; i < tests.length; i++) {
                    $('#tests > ul').append('<li><a href="#" testid="' + tests[i].Id + '">' + tests[i].Name + ' ' + new Date(tests[i].Time).toLocaleString() + '</a></li>');
                }
                $('#tests a').click(function() {
                    EditView.loadtest($(this).attr('testid'));
                    EditView.hideEditing();
                    $('#test-panel').addClass('in');
                    return true;
                });
                EditView.loadtest(tests[0].Id);
            });
        });
    },

    hideEditing: function() {
        $('#project-panel').removeClass('in');
        $('#user-panel').removeClass('in');
        $('#submission-panel').removeClass('in');
        $('#file-panel').removeClass('in');
    },

    loadsubmission: function(id) {
        $('.submission-input').empty();
        $('.submission-input').val('');
        $.getJSON('submissions?id=' + id, function(data) {
            if (not(data['submissions'])) {
                return;
            }
            var s = data['submissions'][0];
            $('#submission-id').val(s.Id);
            $.getJSON('projects', function(pdata) {
                var projects = pdata['projects'];
                if (not(projects)) {
                    return;
                }
                for (var i = 0; i < projects.length; i++) {
                    if (projects[i].Id === s.ProjectId) {
                        $('#submission-project').append('<option value="' + projects[i].Id + '" selected>' + projects[i].Name + '</option>');
                    } else {
                        $('#submission-project').append('<option value="' + projects[i].Id + '">' + projects[i].Name + '</option>');
                    }
                }
            });
            $.getJSON('usernames', function(udata) {
                var users = udata['usernames'];
                if (not(users)) {
                    return;
                }
                for (var i = 0; i < users.length; i++) {
                    if (users[i] === s.User) {
                        $('#submission-user').append('<option value="' + users[i] + '" selected>' + users[i] + '</option>');
                    } else {
                        $('#submission-user').append('<option value="' + users[i] + '">' + users[i] + '</option>');
                    }
                }
            });
            $.getJSON('files?submission-id=' + s.Id + '&format=nested', function(fdata) {
                $('#files > ul').empty();
                var files = fdata['files'];
                if (not(files)) {
                    $('.file-input').empty();
                    $('.file-input').val('');
                    return;
                }
                var fid = '';
                var c = 0;
                for (t in files) {
                    var tid = 'type-subdropdown-' + (c++).toString();
                    $('#files > ul').append('<li class="dropdown-submenu"><a tabindex="-1" href="#">' + t + '</a><ul id="' + tid + '" class="dropdown-menu" role="menu"></ul></li>');
                    for (n in files[t]) {
                        var nid = 'name-subdropdown-' + (c++).toString();
                        $('#files #' + tid).append('<li class="dropdown-submenu"><a tabindex="-1" href="#">' + n + '</a><ul id="' + nid + '" class="dropdown-menu" role="menu"></ul></li>');
                        for (i in files[t][n]) {
                            if (i == 0) {
                                fid = files[t][n][i].Id;
                            }
                            $('#files #' + nid).append('<li><a class="a-file" href="#" fileid="' + files[t][n][i].Id + '">' + new Date(files[t][n][i].Time).toLocaleString() + '</a></li>');
                        }
                    }
                }
                $('#files .a-file').click(function() {
                    EditView.loadfile($(this).attr('fileid'));
                    EditView.hideEditing();
                    $('#file-panel').addClass('in');
                    return true;
                });
                EditView.loadfile(fid);
            });
        });
    },

    loaduser: function(id) {
        $('.user-input').empty();
        $('.user-input').val('');
        $.getJSON('users?name=' + id, function(data) {
            if (not(data['users'])) {
                return;
            }
            var u = data['users'][0];
            $('#user-name').val(u.Name);
            $('#user-id').val(u.Name);
            $.getJSON('permissions', function(pdata) {
                var perms = pdata['permissions'];
                if (not(perms)) {
                    return;
                }
                for (var i = 0; i < perms.length; i++) {
                    if (perms[i].Access === u.Access) {
                        $('#user-perm').append('<option value="' + perms[i].Access.toString() + '" selected>' + perms[i].Name + '</option>');
                    } else {
                        $('#user-perm').append('<option value="' + perms[i].Access.toString() + '">' + perms[i].Name + '</option>');
                    }
                }
            });
        });
    },

    loadfile: function(id) {
        $('.file-input').empty();
        $('.file-input').val('');
        $.getJSON('files?id=' + id, function(data) {
            if (not(data['files'])) {
                return;
            }
            var f = data['files'][0];
            $('#file-id').val(f.Id);
            $('#file-name').val(f.Name);
            if (f.Type !== 'src' && f.Type !== 'test') {
                $('#file-package').val('');
                $('#file-package').hide();
                return;
            }
            $('#file-package').show();
            $('#file-package').val(f.Package);
        });
    },

    loadtest: function(id) {
        $('.test-input').empty();
        $('.test-input').val('');
        $.getJSON('tests?id=' + id, function(data) {
            if (not(data['tests'])) {
                return;
            }
            var t = data['tests'][0];
            $.getJSON('projects', function(pdata) {
                var projects = pdata.projects;
                if (not(projects)) {
                    return;
                }
                for (var i = 0; i < projects.length; i++) {
                    if (projects[i].Id === t.ProjectId) {
                        $('#test-project').append('<option value="' + projects[i].Id + '" selected>' + projects[i].Name + '</option>');
                    } else {
                        $('#test-project').append('<option value="' + projects[i].Id + '">' + projects[i].Name + '</option>');
                    }
                }
            });
            $('#test-id').val(t.Id);
            $('#test-name').val(t.Name);
            $('#test-package').val(t.Package);
            $.getJSON('test-types', function(tdata) {
                var tipes = tdata['types'];
                if (not(tipes)) {
                    return;
                }
                for (var i = 0; i < tipes.length; i++) {
                    if (tipes[i].ID === t.Type) {
                        $('#test-type').append('<option value="' + tipes[i].ID + '" selected>' + tipes[i].Name + '</option>');
                    } else {
                        $('#test-type').append('<option value="' + tipes[i].ID + '">' + tipes[i].Name + '</option>');
                    }
                }
            });
            if (t.Target) {
                $('#test-target-name').val(t.Target.Name + '.' + t.Target.Ext);
                $('#test-target-package').val(t.Target.Package);
            }
        });
    }
}

function round(n, p) {
    if (p === undefined) {
        p = 2
    }
    return +Number(n).toFixed(2);
}

var CodeView = {
    annotations: false,
    init: function(user, currentID, nextID) {
        $(function() {
            $('#checkbox-annotations').bootstrapSwitch();
            $('#checkbox-annotations').on('switchChange.bootstrapSwitch', CodeView.toggleAnnotations);
            $('#btn-cfg-annotations').click(function() {
                $('#modal-annotations').modal('show');
                return false;
            });
            SyntaxHighlighter.defaults['toolbar'] = false;
            SyntaxHighlighter.highlight();
            $('#modal-code #input-user').val(user);
            $('.gutter .line').wrap('<a href=""></a>');
            $('.gutter a').click(CodeView.commentModal);
            $('#submit-comment').click(CodeView.submitComment);
            CodeView.addFile(currentID);
            CodeView.addFile(nextID);
        });
    },
    toggleAnnotations: function() {
        CodeView.annotations = !CodeView.annotations;
        $('#btn-cfg-annotations').toggle();
        if (CodeView.annotations) {
            $('#btn-cfg-annotations').click();
        }
        return false;
    },
    addFile: function(fileID) {
        $('.fileid' + fileID).attr('fileid', fileID);
        CodeView.loadComments(fileID);
        CodeView.loadResults(fileID);
    },
    submitComment: function() {
        $('#modal-code .alert p').empty();
        $('#modal-code .alert').hide();
        var r = CodeView.createComment();
        if (r[1] !== '') {
            $('#modal-code .alert p').append(r[1]);
            $('#modal-code .alert').show();
        } else if (r[0] !== null) {
            CodeView.addInfo(r[0]['file-id'], 'Comments', r[0]['user'], r[0]['data'], r[0]['start'], r[0]['end']);
            $.post('addcomment', r[0], function(data) {
                $('#modal-code').modal('hide');
                if (data["error"] !== "none" && !not(data["error"])) {
                    console.log(data["error"]);
                }
            });
        }
        return false;
    },
    createComment: function() {
        var data = $('#modal-code textarea').val().trim();
        if (data === '') {
            return [null, 'Cannot submit empty comment.']
        }
        var start = Number($('#line-start').val());
        var end = Number($('#line-end').val());
        if (end < start) {
            return [null, 'End line cannot be smaller than start line.'];
        }
        var min = Number($('#line-start').attr('min'));
        if (start < min) {
            return [null, 'Comment start line cannot be smaller than the first line of the file.'];
        }
        var max = Number($('#line-start').attr('max'));
        if (end > max) {
            return [null, 'Comment end line cannot be larger than the last line of the file.'];
        }
        return [{
            'data': data,
            'start': start,
            'end': end,
            'file-id': $('#input-file-id').val(),
            'user': $('#input-user').val()
        }, '']
    },
    commentModal: function() {
        var fileID = $(this).closest('.code').attr('fileid');
        var line = Number($(this).find('.line').html());
        var min = Number($('.fileid' + fileID + ' .gutter .line').first().html());
        var max = Number($('.fileid' + fileID + ' .gutter .line').last().html());
        $('#modal-code #input-file-id').val(fileID);
        $('#modal-code .input-line').val(min);
        $('#modal-code .input-line').attr('min', min);
        $('#modal-code .input-line').attr('max', max);
        $('#modal-code .alert p').empty();
        $('#modal-code textarea').val('');
        $('#modal-code .alert').hide();
        $('#modal-code .input-line').val(line);
        $('#modal-code').modal('show');
        return false;
    },
    addInfo: function(fileID, type, title, content, start, end, added) {
        for (var i = start; i <= end; i++) {
            if (added !== undefined && i > 0 && i <= added.length) {
                if (added[i - 1][type + title + content] === true) {
                    continue;
                } else {
                    added[i - 1][type + title + content] = true;
                }
            }
            CodeView.createPopover(fileID, i);
            $('.fileid' + fileID + ' .code .number' + i).children(':not(.spaces)').addClass('underlineable-' + type);
            var s = '.popover-content[fileid="' + fileID + '"][linenum="' + i + '"]';
            if (!$(s + ' .annotation-' + type).length) {
                $(s).append('<div class="annotation-' + type + '" show-annotation="false"><h4>' + type + '</h4></div>');
            }
            if (!$(s + ' .annotation-' + type + ' [annotation-title="' + title + '"]').length) {
                $(s + ' .annotation-' + type).append('<div annotation-title="' + title + '"><h5>' +
                    title + '</h5><p style="font-size:80%;"></p>');
            }
            $(s + ' .annotation-' + type + ' [annotation-title="' + title + '"] p').append(content + '\n');
        }
    },
    createPopover: function(fileID, num) {
        if (!$('.popover-content[fileid="' + fileID + '"][linenum="' + num + '"]').length) {
            $('.fileid' + fileID + ' .code .number' + num).attr('data-toggle', 'popover');
            $('.fileid' + fileID + ' .code .number' + num).popover({
                html: true,
                trigger: 'hover',
                placement: 'top',
                content: function() {
                    return CodeView.lineContent(fileID, num);
                },
                container: 'body'
            });
            $('body').append('<div class="popover-content" fileid="' + fileID + '" linenum="' + num + '" hidden></div>');
        }

    },
    lineContent: function(fileID, num) {
        if (!CodeView.annotations) {
            return '';
        }
        var c = '';
        $('.popover-content[fileid="' + fileID + '"][linenum="' + num + '"] [show-annotation="true"]').each(function() {
            if ($(this).children('div').not('[style*="display: none"]').length) {
                c += $(this).html();
            }
        });
        return c;
    },
    loadComments: function(fileID) {
        CodeView.addConfiguration('Comments');
        $.getJSON('comments?file-id=' + fileID, function(data) {
            var cs = data['comments'];
            if (not(cs)) {
                return;
            }
            for (var i = 0; i < cs.length; i++) {
                CodeView.addAdvancedConfiguration('Comments', cs[i].User);
                CodeView.addInfo(fileID, 'Comments', cs[i].User, cs[i].Data, cs[i].Start, cs[i].End);
            }
        });
    },
    loadResults: function(fileID) {
        $.getJSON('fileresults?id=' + fileID, function(data) {
            var rs = data['fileresults'];
            if (not(rs)) {
                return;
            }
            var added = Array.apply(null, Array(Number($('.fileid' + fileID + ' .gutter .line').last().html()))).map(function() {
                return {};
            });
            for (var k in rs) {
                CodeView.addConfiguration(k);
                for (var i = 0; i < rs[k].length; i++) {
                    CodeView.addAdvancedConfiguration(k, rs[k][i].Title);
                    CodeView.addInfo(fileID, k, rs[k][i].Title, rs[k][i].Description, rs[k][i].Start, rs[k][i].End, added);
                }
            }
        });
    },
    addAdvancedConfiguration: function(type, title) {
        if ($('#accordion-' + type + '-advanced .panel-body [infotitle="' + title + '"]').length) {
            return;
        }
        var c = $('#accordion-' + type + '-advanced .panel-body').length;
        $('#accordion-' + type + '-advanced .panel-body').append('<div class="form-group"><label class="col-sm-5 control-label" for="checkbox-' + type + '-advanced-' + c + '">' + title + '</label><div class="col-sm-7"><input type="checkbox" id="checkbox-' + type + '-advanced-' + c + '" class="form-control" infotitle="' + title + '" checked></div></div>');
        $('#accordion-' + type + '-advanced .panel-body [infotitle="' + title + '"]').bootstrapSwitch();
        $('#accordion-' + type + '-advanced .panel-body [infotitle="' + title + '"]').on('switchChange.bootstrapSwitch', function() {
            $(' .annotation-' + type + ' [annotation-title="' + title + '"]').toggle();
            if ($('.annotation-' + type).attr('show-annotation') === 'false') {
                return;
            }
            $('.popover-content .annotation-' + type + ':has([annotation-title="' + title + '"])').each(function() {
                var c = $(this).closest('.popover-content');
                var fileID = c.attr('fileid');
                var line = c.attr('linenum');
                if ($(this).children('div').not('[style*="display: none"]').length) {
                    $('.fileid' + fileID + ' .code .number' + line + ' .underlineable-' + type).addClass('underline-' + type);
                } else {
                    $('.fileid' + fileID + ' .code .number' + line + ' .underlineable-' + type).removeClass('underline-' + type);
                }
            });
        });
    },
    addConfiguration: function(type) {
        if ($('#modal-form-' + type).length) {
            return;
        }
        $('#modal-annotations .modal-body').append('<form id="modal-form-' + type + '" class="form-horizontal"><h4>' + type + '</h4></form>');
        $('#modal-form-' + type).append('<div class="form-group"><label class="col-sm-5 control-label" for="checkbox-' + type + '">State</label><div class="col-sm-7"><input type="checkbox" id="checkbox-' + type + '" class="form-control"></div></div>');
        $('#checkbox-' + type).bootstrapSwitch();
        $('#modal-form-' + type).append('<div class="form-group"><label class="col-sm-5 control-label" for="picker-' + type + '">Colour</label><div class="col-sm-7"><input type="text" id="picker-' + type + '" class="pick-a-color form-control" value="000"></div></div>');
        $('#picker-' + type).pickAColor({
            inlineDropdown: true
        });
        $('#modal-form-' + type).append('<input type="hidden" id="color-' + type + '" value="000">');
        $('#modal-form-' + type).append('<input type="hidden" id="prevcolor-' + type + '" value="000">');
        $('#modal-form-' + type).append('<div class="panel-group" id="accordion-' + type + '-advanced"><div class="panel panel-default"><div class="panel-heading"><h5 class="panel-title"><a data-toggle="collapse" data-parent="#accordion-' + type + '-advanced" href="#collapse-' + type + '-advanced">Advanced</a></h5></div><div id="collapse-' + type + '-advanced" class="panel-collapse collapse"><div class="panel-body"></div></div></div></div>');
        $('<style type="text/css">.line .underlineable-' + type + '.underline-' + type + '{border-bottom-width : 1px !important; border-bottom-style : solid !important;} </style>').appendTo('head');
        $('#checkbox-' + type).on('switchChange.bootstrapSwitch', function() {
            var ov = $('.annotation-' + type).attr('show-annotation');
            var nv = $('.annotation-' + type).attr('show-annotation') === 'true' ? 'false' : 'true';
            $('.annotation-' + type).attr('show-annotation', nv);
            if (nv === 'true') {
                $('.popover-content .annotation-' + type).each(function() {
                    if ($(this).children('div').not('[style*="display: none"]').length) {
                        var c = $(this).closest('.popover-content');
                        var fileID = c.attr('fileid');
                        var line = c.attr('linenum');
                        $('.fileid' + fileID + ' .code .number' + line + ' .underlineable-' + type).addClass('underline-' + type);
                    }
                });
            } else {
                $('.underlineable-' + type).removeClass('underline-' + type);
            }
        });
        $('#picker-' + type).on('change', function() {
            var c = $(this).val();
            $('#color-' + type).val(c);
            jss.set('.line .underlineable-' + type + '.underline-' + type, {
                'border-bottom-color': '#' + c + ' !important'
            });
        });
    }
}
