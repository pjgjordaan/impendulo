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

function uniqueProperties(arr, key) {
    var u = {},
        a = [];
    for (var i = 0, l = arr.length; i < l; ++i) {
        if (u.hasOwnProperty(arr[i][key])) {
            continue;
        }
        a.push(arr[i]);
        u[arr[i][key]] = 1;
    }
    return a;
}

function star(x, y, scale, inner, outer, arms) {
    scale = scale || 1;
    arms = arms || 8;
    var innerRadius = (inner || 2) * scale;
    var outerRadius = (outer || 10) * scale;
    var results = '';
    var angle = Math.PI / arms;
    for (var i = 0; i < 2 * arms; i++) {
        var r = (i & 1) === 0 ? outerRadius : innerRadius;
        var currX = x + Math.cos(i * angle) * r;
        var currY = y + Math.sin(i * angle) * r;
        if (i === 0) {
            results = currX + ',' + currY;
        } else {
            results += ', ' + currX + ',' + currY;
        }
    }
    return results;
}

function shadeColor(color, percent) {
    var f = parseInt(color.slice(1), 16),
        t = percent < 0 ? 0 : 255,
        p = percent < 0 ? percent * -1 : percent,
        R = f >> 16,
        G = f >> 8 & 0x00FF,
        B = f & 0x0000FF;
    return '#' + (0x1000000 + (Math.round((t - R) * p) + R) * 0x10000 + (Math.round((t - G) * p) + G) * 0x100 + (Math.round((t - B) * p) + B)).toString(16).slice(1);
}


function extent(data, f) {
    var e = d3.extent(data, f);
    var s = 0.05 * (e[1] - e[0]);
    if (e[0] == e[1]) {
        s = 10;
    }
    e[0] -= s;
    e[1] += s;
    return e;
}

function intervals(extent, n) {
    var vals = new Array(n);
    vals[0] = extent[0];
    vals[n - 1] = extent[1];
    var step = (extent[1] - extent[0]) / n;
    for (i = 1; i < n - 1; i++) {
        vals[i] = i * step + extent[0];
    }
    return vals;
}

function closest(vals, a) {
    var prev = undefined;
    for (i = 0; i < vals.length; i++) {
        var diff = vals[i] > a ? vals[i] - a : a - vals[i];
        if (prev !== undefined && diff > prev) {
            return vals[i - 1];
        }
        prev = diff;
    }
    return vals[vals.length - 1];
}
