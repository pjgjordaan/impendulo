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
    var dest = document.createElement("option");
    dest.innerHTML = src.innerHTML;
    dest.setAttribute("value", src.value);
    dest.setAttribute("onclick", "movevalue('"+destParentId+"', '"+srcParentId+"', this)");
    var destParent = document.getElementById(destParentId);
    var srcParent = document.getElementById(srcParentId);
    srcParent.removeChild(src);
    if(destParent.getAttribute("added") === "true"){
	dest.setAttribute("selected", true);
    } else{
	var nodes = srcParent.childNodes;
	for(var i=0; i<nodes.length; i++) {
	    if (nodes[i].nodeName.toLowerCase() == 'option') {
		nodes[i].setAttribute("selected", true);
	    }
	}
    }
    destParent.appendChild(dest);
}

function unhide(it, box) {
    var check = (box.checked) ? "block" : "none";
    document.getElementById(it).style.display = check;
}

function replacevalue(srcParentID, destID, src) {
    var dest = document.getElementById(destID);
    if(dest === null){
	return;
    }
    var srcParent = document.getElementById(srcParentID);
    if(dest.value != ""){
	var newChild = document.createElement("option");
	newChild.innerHTML = dest.value;
	newChild.setAttribute("value", dest.value);
	newChild.setAttribute("onclick", "replacevalue('"+srcParentID+"', '"+destID+"', this)");
    	srcParent.appendChild(newChild);
    }
    dest.setAttribute("value", src.value);
    dest.setAttribute("onclick", "movevalueback('"+srcParentID+"', this)");
    srcParent.removeChild(src);
}

function movevalueback(destParent, src) {
    var dest = document.createElement("option");
    dest.innerHTML = src.value;
    dest.setAttribute("value", src.value);
    dest.setAttribute("onclick", "replacevalue('"+destParent+"', '"+src.getAttribute("id")+"', this)");
    document.getElementById(destParent).appendChild(dest);
    src.setAttribute("value", "");
}

function showdescription(description) {
    document.getElementById("description").innerHTML = description;
}

function movedescriptionvalue(srcParent, destParent, srcId) {
    var src = document.getElementById(srcId);
    console.log(src);
    var id = src.getAttribute("ruleid");
    var name = src.getAttribute("rulename");
    var description = src.getAttribute("ruledescription");
    var dest = document.createElement("option");
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