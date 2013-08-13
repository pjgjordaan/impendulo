function movevalue (srcParent, destParent, src) {
    var dest = document.createElement("option");
    dest.innerHTML = src.value;
    dest.setAttribute("value", src.value);
    dest.setAttribute("onclick", "movevalue('"+destParent+"', '"+srcParent+"', this)");
    document.getElementById(destParent).appendChild(dest);
    document.getElementById(srcParent).removeChild(src);
}

function unhide (it, box) {
    var check = (box.checked) ? "block" : "none";
    document.getElementById(it).style.display = check;
}
