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
                    title + '</h5><p class="annotation-content"></p>');
            }
            $(s + ' .annotation-' + type + ' [annotation-title="' + title + '"] p.annotation-content').append(content + '<br>');
        }
    },

    createPopover: function(fileID, num) {
        if (!$('.popover-content[fileid="' + fileID + '"][linenum="' + num + '"]').length) {
            $('.fileid' + fileID + ' .code .number' + num).attr('data-toggle', 'popover');
            $('.fileid' + fileID + ' .code .number' + num).attr('tabindex', '0');
            $('.fileid' + fileID + ' .code .number' + num).popover({
                html: true,
                placement: 'auto bottom',
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
                console.log('could not load comments for ' + fileID, data);
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
                console.log('could not load results for ' + fileID, data);
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
                        $('.fileid' + fileID + ' .code .number' + line + ' .underlineable-' + type).css('cursor', 'pointer');
                    }
                });
            } else {
                $('.underlineable-' + type).removeClass('underline-' + type);
                $('.underlineable-' + type).css('cursor', 'default');
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
