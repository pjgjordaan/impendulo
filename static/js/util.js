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

function addJPFListeners(dest){
    $.getJSON('jpflisteners', function(data){   
	$('#'+dest).empty();
	for(var i in data.listeners) {
	    var v =data.listeners[i].Package+'.'+data.listeners[i].Name; 
	    $('#'+dest).append('<option value="'+v+'">'+v+'</option>');
	}
	$('#'+dest).multiselect({
	    selectedText: "# of # listeners selected",
	    noneSelectedText: "Select listeners",
	    classes: "multiselect-listeners"
	});
	$('#'+dest).multiselected = true;
    });
}


function addJPFSearches(dest){
    $.getJSON('jpfsearches', function(data){   
	$('#'+dest).empty();
	for(var i in data.searches) {
	    var v =data.searches[i].Package+'.'+data.searches[i].Name; 
	    $('#'+dest).append('<option value="'+v+'">'+v+'</option>');
	}
    });
}

function addPMDRules(dest){
    $.getJSON('pmdrules', function(data){   
	$('#'+dest).empty();
	for(var i in data.rules) {
	    $('#'+dest).append('<option description="'+data.rules[i].Description+'" value="'+data.rules[i].Id+'">'+data.rules[i].Name+'</option>');
	}
	$('#'+dest).multiselect({
	    selectedText: "# of # rules selected",
	    noneSelectedText: "Select rules",
	    classes: "multiselect-rules"
	});
	$('.multiselect-rules li').tooltip({
	    title: function(){
		return $('option[value="'+$(this).find('input').val()+'"]').attr('description');
	    },
	    placement: 'left',
	    container: 'body'
	});
	$('#'+dest).multiselected = true;
    });
}

function addSkeletons(src, dest, skeletonMap){
    var id = $('#'+src).val();
    $.getJSON('skeletons?project-id='+id, function(data){   
	$('#'+dest).empty();
	if (data.skeletons === null || data.skeletons.length === 0){
	    return;
	}
	for(var i = 0; i < data.skeletons.length; i++) {
	    $('#'+dest).append('<option value="'+data.skeletons[i].Id+'">'+data.skeletons[i].Name+'</option>');
	}
    });
}


function populate(src, toolDest, userDest){
    ajaxSelect(src, toolDest, 'tools?project-id=', 'tools');
    ajaxSelect(src, userDest, 'usernames?project-id=', 'usernames');
}

function ajaxSelect(src, dest, url, name){
    var val = $('#'+src).val();
    $.getJSON(url+val, function(data){   
	$('#'+dest).multiselect();
	$('#'+dest).multiselect('destroy');
	$('#'+dest).empty();
	var items = data[name];
	for(var i = 0; i < items.length; i++) {
	    $('#'+dest).append('<option value="'+items[i]+'">'+items[i]+'</option>');
	}
	$('#'+dest).multiselect();
	$('#'+dest).multiselected = true;
    });	 
}

function hideEditing(){
    $('#project-panel').removeClass('in');
    $('#user-panel').removeClass('in');
    $('#submission-panel').removeClass('in');
    $('#file-panel').removeClass('in');
}

function loadproject(id, idDest, nameDest, userDest, langDest, subDest){
    hideEditing();
    $('#project-panel').addClass('in');
    $.getJSON('projects?id='+id, function(data){   
	var p = data['projects'][0];
	$('#'+idDest).val(p.Id) ;
	$('#'+nameDest).val(p.Name) ;
	$.getJSON('usernames', function(udata){   
	    $('#'+userDest).empty();
	    var users = udata['usernames'];
	    if(not(users)){
		return;
	    }
	    for(var i = 0; i < users.length; i++) {
		if(users[i] === p.User){
		    $('#'+userDest).append('<option value="'+users[i]+'" selected>'+users[i]+'</option>');
		}else{
		    $('#'+userDest).append('<option value="'+users[i]+'">'+users[i]+'</option>');
		}
	    }
	});	 
	$.getJSON('langs', function(ldata){   
	    $('#'+langDest).empty();
	    var langs = ldata['langs'];
	    if(not(langs)){
		return;
	    }
	    for(var i = 0; i < langs.length; i++) {
		if(langs[i] === p.Lang){
		    $('#'+langDest).append('<option value="'+langs[i]+'" selected>'+langs[i]+'</option>');
		}else{
		    $('#'+langDest).append('<option value="'+langs[i]+'">'+langs[i]+'</option>');
		}
	    }
	});	 
	$.getJSON('submissions?project-id='+p.Id, function(sdata){   
	    $('#'+subDest+' > ul').empty();
	    var subs = sdata['submissions'];
	    if(not(subs)){
		$('#files > ul').empty();
		$('#file-name').empty();
		$('#file-package').empty();
		$('#submission-project').empty();
		$('#submission-user').empty();
		return;
	    }
	    for(var i = 0; i < subs.length; i++) {	  
		$('#'+subDest+' > ul').append('<li><a href="#" subid="'+subs[i].Id+'">'+subs[i].User+' '+new Date(subs[i].Time).toLocaleString()+'</a></li>');
	    }
	    $('#'+subDest+' a').click(function(){
		loadsubmission($(this).attr('subid'), 'submission-id', 'submission-project', 'submission-user', 'files');
		hideEditing();
		$('#submission-panel').addClass('in');
		return true;
	    });
	    loadsubmission(subs[0].Id, 'submission-id', 'submission-project', 'submission-user', 'files');
	});	 
    });
}

function not(v){
    return v === null || v === undefined || v.length === 0; 
}


function loadsubmission(id, idDest, projectDest, userDest, fileDest){
    $.getJSON('submissions?id='+id, function(data){   
	var s = data['submissions'][0];
	$('#'+idDest).val(s.Id) ;
	$.getJSON('projects', function(pdata){   
	    $('#'+projectDest).empty();
	    var projects = pdata['projects'];
	    if(not(projects)){
		return;
	    }
	    for(var i = 0; i < projects.length; i++) {
		if(projects[i].Id === s.ProjectId){
		    $('#'+projectDest).append('<option value="'+projects[i].Id+'" selected>'+projects[i].Name+'</option>');
		}else{
		    $('#'+projectDest).append('<option value="'+projects[i].Id+'">'+projects[i].Name+'</option>');
		}
	    }
	});	 
	$.getJSON('usernames', function(udata){   
	    $('#'+userDest).empty();
	    var users = udata['usernames'];
	    if(not(users)){
		return;
	    }
	    for(var i = 0; i < users.length; i++) {
		if(users[i] === s.User){
		    $('#'+userDest).append('<option value="'+users[i]+'" selected>'+users[i]+'</option>');
		}else{
		    $('#'+userDest).append('<option value="'+users[i]+'">'+users[i]+'</option>');
		}
	    }
	});	 
	$.getJSON('files?submission-id='+s.Id +'&format=nested', function(fdata){   
	    $('#'+fileDest+' > ul').empty();
	    var files = fdata['files'];
	    if(not(files)){
		$('#file-name').empty();
		$('#file-package').empty();
		return;
	    }
	    var fid = '';
	    var c = 0;
	    for(t in files) {
		var tid = 'type-subdropdown-'+(c++).toString();
		$('#'+fileDest+' > ul').append('<li class="dropdown-submenu"><a tabindex="-1" href="#">'+t+'</a><ul id="'+tid+'" class="dropdown-menu" role="menu"></ul></li>');
		for(n in files[t]){
		    var nid = 'name-subdropdown-'+(c++).toString();
		    $('#'+fileDest+' #'+tid).append('<li class="dropdown-submenu"><a tabindex="-1" href="#">'+n+'</a><ul id="'+nid+'" class="dropdown-menu" role="menu"></ul></li>'); 
		    for(i in files[t][n]){
			if(i == 0){
			    fid = files[t][n][i].Id;
			}
			$('#'+fileDest+' #'+nid).append('<li><a class="a-file" href="#" fileid="'+files[t][n][i].Id+'">'+new Date(files[t][n][i].Time).toLocaleString()+'</a></li>');
	 	    }
		}
	    }
	    $('#'+fileDest+' .a-file').click(function(){
		loadfiles($(this).attr('fileid'), 'file-id', 'file-name', 'file-package', 'file-code');
		hideEditing();
		$('#file-panel').addClass('in');
		return true;
	    });
	    loadfiles(fid, 'file-id', 'file-name', 'file-package', 'file-code');
	});	 	 	 
    });
}

function loaduser(id, idDest, nameDest, permDest){
    $.getJSON('users?name='+id, function(data){   
	var u = data['users'][0];
	$('#'+nameDest).val(u.Name);
	$('#'+idDest).val(u.Name);
	$.getJSON('permissions', function(pdata){   
	    $('#'+permDest).empty();
	    var perms = pdata['permissions'];
	    if(not(perms)){
		return;
	    }
	    for(var i = 0; i < perms.length; i++) {
		if(perms[i].Access === u.Access){
		    $('#'+permDest).append('<option value="'+perms[i].Access.toString()+'" selected>'+perms[i].Name+'</option>');
		}else{
		    $('#'+permDest).append('<option value="'+perms[i].Access.toString()+'">'+perms[i].Name+'</option>');
		}
	    }
	});
    });
}

function loadfiles(id, idDest, nameDest, pkgDest, codeDest){
    $.getJSON('files?id='+id, function(data){   
	var f = data['files'][0];
	$('#'+idDest).val(f.Id);
	$('#'+nameDest).val(f.Name);
	if(f.Type !== 'src' && f.Type !== 'test'){
	    $('#'+pkgDest).val('');
	    $('#'+codeDest).html('');	 	    
	    $('#'+pkgDest).hide();
	    $('#'+codeDest).hide();
	    return;
	}
	$('#'+pkgDest).val(f.Package);
	$.getJSON('code?id='+id, function(cdata){   
	    $('#'+codeDest).html(cdata['code']);	 	 
	});
    });
}

function addPopover(dest, src){
    $('body').on('click', function (e) {
        if ($('#'+dest).next('div.popover:visible').length > 0
	    && $(e.target).data('toggle') !== 'popover'
            && e.target.id !== dest
	    && e.target.id !== 'codepopover'
	    && $('#codepopover').find($(e.target)).length === 0){
  		$('#'+dest).click();
        }    
    });
    window.onload = function() {
	$('#'+dest).popover({
	    template: '<div id="codepopover" class="popover code-popover"><div class="arrow"></div><div class="popover-inner"><h3 class="popover-title"></h3><div class="popover-content code-popover-content"><p></p></div></div></div>',
	    placement : 'bottom', 
	    html: 'true',
	    content :  $('#'+src).html(),
	});
    };
}

function showToolCode(name, pid, title){
    var id = 'toolcode-modal';
    var s = '#'+id;
    if($(s).length > 0){
	$(s).modal('show');
	$(s).on('shown.bs.modal', function(e){
	    line.scrollIntoView(); 
	});
	return;
    }
    $.getJSON('code?tool-name='+name+'&project-id='+pid, function(data){
	jQuery('<div id="'+id+'" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="'+id+'label" aria-hidden="true"><div class="modal-dialog"><div class="modal-content"><div class="modal-header"><button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button><h4 class="modal-title" id="'+id+'label">'+title+'</h4></div><div class="modal-body"><script class="brush: java;" type="syntaxhighlighter"><![CDATA['+data.code+']]></script></div></div></div></div>').appendTo('body');
	SyntaxHighlighter.defaults['toolbar'] = false;
	SyntaxHighlighter.defaults['class-name'] = 'error';
	SyntaxHighlighter.highlight(); 
	$(s).modal('show');
	$(s).on('shown.bs.modal', function(e){
	    $(s).animate({
		scrollTop: $(s).offset()
	    });
	});
	return false;
    });
}

function addCodeModal(dest, resultId, bug, start, end){
    $('#'+dest).click(function(){
	var id = 'bug-modal';
	var s = '#'+id;
	if($(s).length > 0){
	    $(s).modal('show');
	    $(s).on('shown.bs.modal', function(e){
		line.scrollIntoView(); 
	    });
	    return false;
	}
	$.getJSON('code?result-id='+resultId, function(data){
	    var h = 'highlight: [';
	    for(var i = start; i < end; i ++){
		h += i + ',';
	    }
	    h = h + end + '];'
	    var preClass = 'brush: java; '+h;
	    jQuery('<div id="'+id+'" class="modal fade" tabindex="-1" role="dialog" aria-labelledby="'+id+'label" aria-hidden="true"><div class="modal-dialog"><div class="modal-content"><div class="modal-header"><button type="button" class="close" data-dismiss="modal" aria-hidden="true">&times;</button><h4 class="modal-title" id="'+id+'label">'+bug.title+'<br><small>'+bug.content+'</small></h4></div><div class="modal-body"><script class="'+preClass+'" type="syntaxhighlighter"><![CDATA['+data.code+']]></script></div></div></div></div>').appendTo('body');
	    SyntaxHighlighter.defaults['toolbar'] = false;
	    SyntaxHighlighter.defaults['class-name'] = 'error';
	    SyntaxHighlighter.highlight(); 
	    $(s).find('.highlighted').attr('style', 'background-color: #ff7777 !important;');
	    $(s).modal('show');
	    $(s).on('shown.bs.modal', function(e){
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

function ajaxChart(vals){
    if(vals.subID === undefined){
	return;
    }
    if(vals.file === undefined){
	return;
    }
    if(vals.result === undefined){
	return;
    }
    var subs = [vals.subID];
    if(vals.src !== undefined){
	subs = subs.concat($('#' + vals.src).val());
    }
    var params = {'type':'file','submissions': subs, 'file': vals.file, 'result': vals.result};
    if(vals.testfileID !== undefined){
	params.testfileid = vals.testfileID;
    }
    if(vals.srcfileID !== undefined){
	params.srcfileid = vals.srcfileID;
    }
    $.getJSON('chart', params, function(data){
	showChart(data['chart'], vals.currentTime, vals.nextTime, vals.user);
    });
    return false;
}


function addComparables(rid, pid, dest, user){
    $('#'+dest).multiselect();
    $('#'+dest).multiselect('destroy');
    $('#'+dest).empty();
    $.getJSON('comparables?id='+rid, function(cdata){
	var comp = cdata['comparables'];
	for(var i = 0; i < comp.length; i++) {
	    $('#'+dest).append('<option value="'+comp[i].Id+'">'+comp[i].Name+'</option>');
	}
	$.getJSON('submissions?project-id='+pid, function(data){
	    var items = data['submissions'];
	    for(var i = 0; i < items.length; i++) {
		if(items[i].User === user){
		    continue;
		}
		$('#'+dest).append('<option value="'+items[i].Id+'">'+items[i].User+ ' - ' + new Date(items[i].Time).toLocaleString()+'</option>');
	    }
	    $('#'+dest).multiselect({
		noneSelectedText: 'Compare results',
		selectedText: '# selected to compare'
	    });
	    $('#'+dest).multiselected = true;
	});
    });  
}


function addResults(id, dest, tipe){
    var d = '#'+dest;
    $(d).empty();
    $.getJSON('results?'+tipe+'-id='+id, function(data){
	var r = data['results'];
	for(var i = 0; i < r.length; i++) {
	    $(d).append('<option value="'+r[i].Id+'">'+r[i].Name+'</option>');
	}
	loadSubmissionsChart(id, tipe,r[0].Id,$('#score').val());
    }); 
}

function loadSubmissionsChart(id, tipe, r,score){
    var params = {'type':'submission','id':id,'result':r, 'submission-type':tipe,'score':score};
    $.getJSON('chart', params, function(data){
	submissionsChart(data['chart'], tipe);
    });
}

$(function () {
    $('.tree li:has(ul)').addClass('parent_li').find(' > span').attr('title', 'Collapse this branch');
    $('.tree li.parent_li > span').on('click', function (e) {
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


function loadCollections(dbList, collectionList){
    var url = 'collections?db='+$('#'+dbList).val();
    $.getJSON(url, function(data){
	$('#'+collectionList).multiselect();
	$('#'+collectionList).multiselect('destroy');
	$('#'+collectionList).empty();
	var items = data['collections'];
	for(var i = 0; i < items.length; i++) {
	    $('#'+collectionList).append('<option value="'+items[i]+'">'+items[i]+'</option>');
	}
	$('#'+collectionList).multiselect({
	    noneSelectedText: 'Choose collections to export',
	    selectedText: '# collections selected to export'
	});
	$('#'+collectionList).multiselected = true;
    });
}
