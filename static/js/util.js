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
