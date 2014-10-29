var JUnitResult = {
    test: '',
    init: function(test) {
        JUnitResult.test = test;
        $(function() {
            if ($('#junit-modal').length === 0) {
                $('<div id="junit-modal" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="junit-modal-label" aria-hidden="true"><div class="modal-dialog"><div class="modal-content"><div class="modal-header"><button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button><h4 class="modal-title" id="junit-modal-label"><br><small></small></h4></div><div class="modal-body"></div></div></div></div>').appendTo('body');
            }
            $('a.testcase').unbind('click', JUnitResult.loadData);
            $('a.testcase').bind('click', JUnitResult.loadData);
        });
    },

    loadData: function(e) {
        var name = $(this).attr('dataname');
        $.getJSON('testdata?data-name=' + name + '&test=' + JUnitResult.test, function(data) {
            if (not(data['data'])) {
                console.log(data);
                return false;
            }
            $('#junit-modal-label').html(name);
            $('#junit-modal .modal-body').html(data['data']);
            $('#junit-modal').modal('show');
        });
        return false;
    }
}
