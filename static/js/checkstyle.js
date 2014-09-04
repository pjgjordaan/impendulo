var CheckstyleResult = {
    addFile: function(parentId, eid, pkg, clss, severity, message, lines, resultID) {
        var pkgID = parentId + pkg;
        var aID = 'accordion' + pkgID;
        if ($('#' + aID).length == 0) {
            $('#' + parentId).append('<div class="panel panel-default"><div class="panel-heading"><a class="accordion-toggle" data-toggle="collapse" data-parent="' + parentId + '" href="#' + pkgID + '"><h4 class="text-center">' + pkg + '</h4></a></div><div id="' + pkgID + '" class="panel-collapse collapse"><div class="accordion-inner"><div class="panel-group" id="' + aID + '"></div></div></div></div>');
        }
        $('#' + aID).append('<div class="panel panel-default"><div class="panel-heading"><a class="accordion-toggle" data-toggle="collapse" data-parent="' + aID + '" href="#' + eid + '"><h5>' + clss + '</h5></a></div><div id="' + eid + '" class="panel-collapse collapse"><div class="accordion-inner"><dl class="dl-horizontal"><dt>Lines</dt><dd class="lines"></dd><dt>Severity</dt><dd>' + severity + '</dd><dt>Description</dt><dd>' + message + '</dd></dl></div></div></div>');
    },
    addLine: function(resultID, eid, num, title, message) {
        var lid = eid + num;
        $('#' + eid + ' dd.lines').append('<a href="#" id="' + lid + '"> ' + num + '; </a>');
        var info = {
            'title': title,
            'content': message
        };
        AnalysisView.addCodeModal(lid, resultID, info, num, num);
    }
}
