function movevalue(srcParent, destParent, src) {
    var dest = document.createElement("option");
    dest.innerHTML = src.value;
    dest.setAttribute("value", src.value);
    dest.setAttribute("onclick", "movevalue('"+destParent+"', '"+srcParent+"', this)");
    if(destParent == "addedL"){
	dest.setAttribute("selected", true);
    }
    document.getElementById(destParent).appendChild(dest);
    document.getElementById(srcParent).removeChild(src);
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