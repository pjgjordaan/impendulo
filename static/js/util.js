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

function movevalue(srcParentId, destParentId, src) {
    var dest = document.createElement('option');
    dest.innerHTML = src.innerHTML;
    dest.setAttribute('value', src.value);
    dest.setAttribute('onclick', "movevalue('"+destParentId+"', '"+srcParentId+"', this)");
    var destParent = document.getElementById(destParentId);
    var srcParent = document.getElementById(srcParentId);
    srcParent.removeChild(src);
    if(destParent.getAttribute('added') === 'true'){
	dest.setAttribute('selected', true);
    } else{
	var nodes = srcParent.childNodes;
	for(var i=0; i<nodes.length; i++) {
	    if (nodes[i].nodeName.toLowerCase() == 'option') {
		nodes[i].setAttribute('selected', true);
	    }
	}
    }
    destParent.appendChild(dest);
}

function unhide(it, box) {
    var check = (box.checked) ? 'block' : 'none';
    document.getElementById(it).style.display = check;
}

function replacevalue(srcParentID, destID, src) {
    var dest = document.getElementById(destID);
    if(dest === null){
	return;
    }
    var srcParent = document.getElementById(srcParentID);
    if(dest.value != ''){
	var newChild = document.createElement('option');
	newChild.innerHTML = dest.value;
	newChild.setAttribute('value', dest.value);
	newChild.setAttribute('onclick', "replacevalue('"+srcParentID+"', '"+destID+"', this)");
    	srcParent.appendChild(newChild);
    }
    dest.setAttribute('value', src.value);
    dest.setAttribute('onclick', "movevalueback('"+srcParentID+"', this)");
    srcParent.removeChild(src);
}

function movevalueback(destParent, src) {
    var dest = document.createElement('option');
    dest.innerHTML = src.value;
    dest.setAttribute('value', src.value);
    dest.setAttribute('onclick', "replacevalue('"+destParent+"', '"+src.getAttribute("id")+"', this)");
    document.getElementById(destParent).appendChild(dest);
    src.setAttribute('value', '');
}

function showdescription(description) {
    document.getElementById('description').innerHTML = description;
}

function movedescriptionvalue(srcParent, destParent, srcId) {
    var src = document.getElementById(srcId);
    var id = src.getAttribute('ruleid');
    var name = src.getAttribute('rulename');
    var description = src.getAttribute('ruledescription');
    var dest = document.createElement('option');
    dest.innerHTML = name;
    dest.setAttribute("ruleid", id);
    dest.setAttribute("rulename", name);
    dest.setAttribute("ruledescription", description);
    dest.setAttribute("onclick", "showdescription('"+description+"')");
    dest.setAttribute("ondblclick", "addalert('"+destParent+"', '"+srcParent+"', this)");
    document.getElementById(destParent).appendChild(dest);
    document.getElementById(srcParent).removeChild(src);
}

function addalert(srcParent, destParent, src) {
    var id = src.getAttribute("ruleid");
    var name = src.getAttribute("rulename");
    var description = src.getAttribute("ruledescription");
    var dest = document.createElement("div");
    dest.setAttribute("class", "alert alert-dismissable alert-list");
    dest.setAttribute("id", id);
    dest.setAttribute("ruleid", id);
    dest.setAttribute("rulename", name);
    dest.setAttribute("ruledescription", description);
    var destButton = document.createElement("button");
    destButton.setAttribute("class", "close");
    destButton.setAttribute("type", "button");
    destButton.setAttribute("data-dismiss", "alert");
    destButton.setAttribute("aria-hidden", "true");
    destButton.setAttribute("onclick", "movedescriptionvalue('"+destParent+"','"+srcParent+"', '"+id+"')");
    destButton.innerHTML = "&times;";
    var destName = document.createElement("strong");
    destName.innerHTML = name+": ";
    var destDescription = document.createElement("small");
    destDescription.setAttribute("class", "text-muted");
    destDescription.innerHTML = description;
    var destAnchor = document.createElement("input");
    destAnchor.setAttribute("type", "hidden");
    destAnchor.setAttribute("name", "ruleid");
    destAnchor.setAttribute("value", id);
    dest.appendChild(destButton);
    dest.appendChild(destName);
    dest.appendChild(destDescription);
    dest.appendChild(destAnchor);
    document.getElementById(destParent).appendChild(dest);
    document.getElementById(srcParent).removeChild(src);
}

function highlight(bug){
    SyntaxHighlighter.defaults['toolbar'] = false;
    SyntaxHighlighter.defaults['class-name'] = 'error';
    SyntaxHighlighter.all();
    if(bug == undefined){
	return;
    }
    window.onload = function(){
	var lines = document.getElementsByClassName('highlighted');
	if(lines[0] != undefined){
	    lines[0].scrollIntoView();
	    var body = bug.Title + '\n';
	    var content = bug.Content;
	    for(var i = 0; i < content.length; i ++){
		body += content[i].trim();
		if(i < content.length - 1){
		    body += '\n';
		}
	    }
	    $(".highlighted").tooltip({
		title: body,
		container : 'body',
		trigger: 'hover',
		placement: 'bottom'
	    });
	    $('.highlighted').attr('style', 'background-color: #ff7777 !important;');
	}
    }
}

function addSkeletons(src, dest, skeletonMap){
    var xmlhttp = new XMLHttpRequest();
    xmlhttp.onreadystatechange=function()
    {
	if (xmlhttp.readyState==4 && xmlhttp.status==200)
	{
	    var destList = document.getElementById(dest);
	    destList.options.length = 0;
	    var skeletons = xmlhttp.responseText.split(",");
	    if (skeletons.length === 1 && skeletons[0].split('_').length === 1){
		return;
	    }
	    for(var i = 0; i < skeletons.length; i++) {
		var option = document.createElement('option');
		var vals = skeletons[i].split('_');
		option.value = vals[0];
		option.text = vals[1];
		destList.add(option);
	    }
	}
    }
    var srcList = document.getElementById(src);
    var id = srcList.options[srcList.selectedIndex].value;
    xmlhttp.open('GET','skeletons?projectid='+id,true);
    xmlhttp.send();
}

function populate(src, toolDest, userDest){
    addTools(src, toolDest);
    addUsers(src, userDest);
}

function ajaxSelect(src, dest, url){
   var xmlhttp = new XMLHttpRequest();
    xmlhttp.onreadystatechange=function()
    {
	if (xmlhttp.readyState==4 && xmlhttp.status==200)
	{
	    $("#"+dest).multiselect();
	    $("#"+dest).multiselect("destroy");
	    var destList = document.getElementById(dest);
	    destList.options.length = 0;
	    var items = xmlhttp.responseText.split(",");
	    for(var i = 0; i < items.length; i++) {
		var option = document.createElement('option');
		option.value = option.text = items[i];
		destList.add(option);
	    }
	    $("#"+dest).multiselect();
	    $("#"+dest).multiselected = true;
	}
    }
    var srcList = document.getElementById(src);
    var val = srcList.options[srcList.selectedIndex].value;
    xmlhttp.open('GET',url+val,true);
    xmlhttp.send();
}

function addTools(src, dest){
    ajaxSelect(src, dest, 'tools?projectid=');
}
 
function addUsers(src, dest){
   ajaxSelect(src, dest, 'users?projectid=');
}


function addPopover(dest, src){
    window.onload = function() {
	$("[rel=codepopover]").popover({
	    template: '<div class="popover code-popover"><div class="arrow"></div><div class="popover-inner"><h3 class="popover-title"></h3><div class="popover-content code-popover-content"><p></p></div></div></div>',
	    placement : 'bottom', 
	    html: 'true',
	    content :  $('#'+src).html()
	});
    };
}
