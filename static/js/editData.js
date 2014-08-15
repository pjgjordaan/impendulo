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

            $('#accordion form').submit(function(e) {
                e.preventDefault();
                var url = $(this).attr("action");
                $.post(url, $(this).serializeArray(), function(data) {
                    if (data["error"] !== "none" && !not(data["error"])) {
                        console.log(data["error"]);
                    } else {
                        EditView.reload(url.substring(4, url.length));
                    }
                });
            });
        });
    },

    reload: function(tipe) {
        switch (tipe) {
            case "project":
                EditView.loadproject($('#project-id').val());
                break;
            case "user":
                EditView.loaduser($('#user-id').val());
                break;
            case "test":
                EditView.loadtest($('#test-id').val());
                break;
            case "submission":
                EditView.loadsubmission($('#submission-id').val());
                break;
            case "assignment":
                EditView.loadassignment($('#assignment-id').val());
                break;
            case "file":
                EditView.loadfile($('#file-id').val());
                break;
            default:
                console.log("error unknown type: ", tipe);
        }
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
            $.getJSON('assignments?project-id=' + p.Id, function(adata) {
                $('#assignments > ul').empty();
                var a = adata['assignments'];
                if (not(a)) {
                    $('#files > ul').empty();
                    $('.file-input').empty();
                    $('.submission-input').empty();
                    $('.assignment-input').empty();
                    $('.file-input').val('');
                    $('.submission-input').val('');
                    $('.assignment-input').val('');
                    return;
                }
                for (var i = 0; i < a.length; i++) {
                    $('#assignments > ul').append('<li><a href="#" aid="' + a[i].Id + '">' + a[i].User + ' ' + new Date(a[i].End).toLocaleString() + '</a></li>');
                }
                $('#assignments a').click(function() {
                    EditView.loadassignment($(this).attr('aid'));
                    EditView.hideEditing();
                    $('#assignment-panel').addClass('in');
                    return true;
                });
                EditView.loadassignment(a[0].Id);
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
        $('#assignment-panel').removeClass('in');
        $('#submission-panel').removeClass('in');
        $('#file-panel').removeClass('in');
    },

    loadassignment: function(id) {
        $('.assignment-input').empty();
        $('.assignment-input').val('');
        $.getJSON('assignments?id=' + id, function(data) {
            if (not(data['assignments'])) {
                return;
            }
            var a = data['assignments'][0];
            $('#assignment-id').val(a.Id);
            $('#assignment-name').val(a.Name);
            $.getJSON('projects', function(pdata) {
                var projects = pdata['projects'];
                if (not(projects)) {
                    return;
                }
                for (var i = 0; i < projects.length; i++) {
                    if (projects[i].Id === a.ProjectId) {
                        $('#assignment-project').append('<option value="' + projects[i].Id + '" selected>' + projects[i].Name + '</option>');
                    } else {
                        $('#assignment-project').append('<option value="' + projects[i].Id + '">' + projects[i].Name + '</option>');
                    }
                }
            });
            $.getJSON('usernames', function(udata) {
                var users = udata['usernames'];
                if (not(users)) {
                    return;
                }
                for (var i = 0; i < users.length; i++) {
                    if (users[i] === a.User) {
                        $('#assignment-user').append('<option value="' + users[i] + '" selected>' + users[i] + '</option>');
                    } else {
                        $('#assignment-user').append('<option value="' + users[i] + '">' + users[i] + '</option>');
                    }
                }
            });
            $.getJSON('submissions?assignment-id=' + a.Id, function(sdata) {
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
        });
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
            $.getJSON('assignments?project-id=' + s.ProjectId + '&max-start=' + s.Time + '&min-end=' + s.Time, function(adata) {
                var as = adata['assignments'];
                if (not(as)) {
                    return;
                }
                for (var i = 0; i < as.length; i++) {
                    if (as[i].Id === s.AssignmentId) {
                        $('#submission-assignment').append('<option value="' + as[i].Id + '" selected>' + as[i].Name + '</option>');
                    } else {
                        $('#submission-assignment').append('<option value="' + as[i].Id + '">' + as[i].Name + '</option>');
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
