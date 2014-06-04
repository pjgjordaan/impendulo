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

var OverviewChart = {
    init: function(tipe) {
        $(function() {
            var params = {
                'type': 'overview',
                'view': tipe,
            };
            $.getJSON('chart', params, function(data) {
                OverviewChart.create(data['chart'], tipe);
            });
        });
    },
    create: function(chartData, tipe) {
        if (chartData === null) {
            return;
        }
        var m = [10, 150, 50, 100];
        var w = 1100 - m[1] - m[3];
        var h = 500 - m[0] - m[2];
        var size = w / chartData.length;
        var yMax = h / 3;
        var duration = 1500;
        var ease = 'quad-in-out';
        var names = chartData.map(function(d) {
            return d.key;
        });
        var categories = ['Submissions', 'Snapshots', 'Launches'];
        var chartColour = function(tipe) {
            return d3.scale.category10()
                .domain(categories)(tipe);
        };
        var getSubmissions = function(d) {
            var url = '';
            if (tipe === 'project') {
                for (var i = 0; i < chartData.length; i++) {
                    if (chartData[i].key === d) {
                        url = 'getsubmissions?project-id=' + chartData[i].id;
                        break;
                    }
                }
            } else if (tipe === 'user') {
                url = 'getsubmissions?user-id=' + d;
            }
            if (url !== '') {
                window.location = url;
            }
        }
        var xScale = d3.scale.ordinal()
            .domain(names)
            .rangeBands([0, w]);
        var xPos = function(d) {
            return xScale(d.key);
        };
        var subDomain = [0, d3.max(chartData, OverviewChart.submissions)];
        var snapDomain = [0, d3.max(chartData, OverviewChart.snapshots)];
        var launchDomain = [0, d3.max(chartData, OverviewChart.launches)];
        var subPos = function(d) {
            return h - subHeight(d);
        };
        var snapPos = function(d) {
            return subPos(d) - snapHeight(d);
        };
        var launchPos = function(d) {
            return snapPos(d) - launchHeight(d);
        };
        var subHeight = function(d) {
            var ret = d3.scale.linear()
                .domain(subDomain)
                .range([0, yMax])(d.submissions);
            return ret;
        };
        var snapHeight = function(d) {
            var ret = d3.scale.linear()
                .domain(snapDomain)
                .range([0, yMax])(d.snapshots);
            return ret;
        };
        var launchHeight = function(d) {
            var ret = d3.scale.linear()
                .domain(launchDomain)
                .range([0, yMax])(d.launches);
            return ret;
        };
        var groupBuffer = 5;
        var x = w / chartData.length - 2 * groupBuffer;
        var groupWidth = x / 3;
        var groupSub = function(d) {
            var ret = d3.scale.linear()
                .domain(subDomain)
                .range([0, h])(d.submissions);
            return ret;
        };
        var groupSnap = function(d) {
            var ret = d3.scale.linear()
                .domain(snapDomain)
                .range([0, h])(d.snapshots);
            return ret;
        };
        var groupLaunch = function(d) {
            var ret = d3.scale.linear()
                .domain(launchDomain)
                .range([0, h])(d.launches);
            return ret;
        };
        var transitionStacked = function() {
            bars.selectAll('.submission')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr('status', 'stacked')
                .attr("x", xPos)
                .attr("y", subPos)
                .attr("width", x)
                .attr("height", subHeight);
            bars.selectAll('.snapshot')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", xPos)
                .attr("y", snapPos)
                .attr("width", x)
                .attr("height", snapHeight);
            bars.selectAll('.launch')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", xPos)
                .attr("y", launchPos)
                .attr("width", x)
                .attr("height", launchHeight);
            bars.selectAll('.subinfo')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", function(d) {
                    return xPos(d) + x / 2.5;
                })
                .attr("y", function(d) {
                    return h - 2;
                });
            bars.selectAll('.snapinfo')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", function(d) {
                    return xPos(d) + x / 2.5;
                })
                .attr("y", function(d) {
                    return subPos(d) - 2;
                });
            bars.selectAll('.launchinfo')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", function(d) {
                    return xPos(d) + x / 2.5;
                })
                .attr("y", function(d) {
                    return snapPos(d) - 2;
                });
        };
        var transitionGrouped = function() {
            bars.selectAll('.submission')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr('status', 'grouped')
                .attr("x", function(d) {
                    return xPos(d) + groupBuffer;
                })
                .attr("y", function(d) {
                    return h - groupSub(d);
                })
                .attr("width", groupWidth)
                .attr("height", groupSub);
            bars.selectAll('.snapshot')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", function(d) {
                    return xPos(d) + groupWidth + groupBuffer;
                })
                .attr("y", function(d) {
                    return h - groupSnap(d);
                })
                .attr("width", groupWidth)
                .attr("height", groupSnap);
            bars.selectAll('.launch')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", function(d) {
                    return xPos(d) + groupWidth * 2 + groupBuffer;
                })
                .attr("y", function(d) {
                    return h - groupLaunch(d);
                })
                .attr("width", groupWidth)
                .attr("height", groupLaunch);
            bars.selectAll('.subinfo')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", function(d) {
                    return xPos(d) + groupWidth * 0.2 + groupBuffer;
                })
                .attr("y", function(d) {
                    return h - groupSub(d) - 2;
                });
            bars.selectAll('.snapinfo')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", function(d) {
                    return xPos(d) + groupWidth * 1.2 + groupBuffer;
                })
                .attr("y", function(d) {
                    return h - groupSnap(d) - 2;
                });
            bars.selectAll('.launchinfo')
                .transition()
                .duration(duration)
                .ease(ease)
                .attr("x", function(d) {
                    return xPos(d) + groupWidth * 2.2 + groupBuffer;
                })
                .attr("y", function(d) {
                    return h - groupLaunch(d) - 2;
                });
        };
        var change = function(d) {
            var status = d3.select('.submission')
                .attr('status');
            if (status === 'stacked') {
                transitionGrouped();
            } else {
                transitionStacked();
            }
        };
        var xAxis = d3.svg.axis()
            .scale(xScale)
            .tickSize(-h)
            .orient('bottom')
            .tickSubdivide(true);
        var chart = d3.select('#overview-chart')
            .append('svg:svg')
            .attr('width', w + m[1] + m[3])
            .attr('height', h + m[0] + m[2])
            .append('svg:g')
            .attr('transform', 'translate(' + m[3] + ',' + m[0] + ')');
        chart.append('svg:rect')
            .attr('width', w)
            .attr('height', h)
            .attr('class', 'plot');
        chart.append('svg:g')
            .attr('font-size', '10px')
            .attr('class', 'x axis')
            .attr('transform', 'translate(0,' + h + ')')
            .call(xAxis);
        chart.select('.x.axis')
            .selectAll('.tick.major')
            .attr('class', 'tick major clickable')
            .on('click', getSubmissions)
        chart.append('text')
            .attr('x', w / 2)
            .attr('y', h + 40)
            .attr('font-size', '20px')
            .style('text-anchor', 'middle')
            .text(tipe === 'project' ? 'Project' : 'User');
        chart.append('svg:clipPath')
            .attr('id', 'clip')
            .append('svg:rect')
            .attr('x', -10)
            .attr('y', -10)
            .attr('width', w + 20)
            .attr('height', h + 20);
        var chartBody = chart.append('g')
            .attr('clip-path', 'url(#clip)')
            .attr('class', 'clickable')
            .on('click', change);
        var bars = chartBody.selectAll('.bars')
            .data(chartData)
            .enter()
            .append('g')
            .attr('class', 'bar');
        bars.append('rect')
            .attr('class', 'submission')
            .attr("x", xPos)
            .attr('fill', chartColour('Submissions'))
            .attr("y", subPos)
            .attr("width", x)
            .attr("height", subHeight)
            .attr('status', 'stacked')
            .append('title')
            .attr('class', 'description')
            .text(function(d) {
                return d.key +
                    '\n' + d.submissions + ' submissions';
            });
        bars.append("rect")
            .attr('class', 'snapshot')
            .attr("x", xPos)
            .attr('fill', chartColour('Snapshots'))
            .attr("y", snapPos)
            .attr("width", x)
            .attr("height", snapHeight)
            .append('title')
            .attr('class', 'description')
            .text(function(d) {
                return d.key +
                    '\n' + d.snapshots + ' snapshots';
            });
        bars.append("rect")
            .attr('class', 'launch')
            .attr("x", xPos)
            .attr('fill', chartColour('Launches'))
            .attr("y", launchPos)
            .attr("width", x)
            .attr("height", launchHeight)
            .append('title')
            .attr('class', 'description')
            .text(function(d) {
                return d.key +
                    '\n' + d.launches + ' launches';
            });
        bars.append('text')
            .attr('class', 'subinfo')
            .attr("x", function(d) {
                return xPos(d) + x / 2.5;
            })
            .attr("y", function(d) {
                return h - 2;
            })
            .attr('font-size', '9px')
            .text(OverviewChart.submissions);
        bars.append('text')
            .attr('class', 'snapinfo')
            .attr("x", function(d) {
                return xPos(d) + x / 2.5;
            })
            .attr("y", function(d) {
                return subPos(d) - 2;
            })
            .attr('font-size', '9px')
            .text(OverviewChart.snapshots);
        bars.append('text')
            .attr('class', 'launchinfo')
            .attr("x", function(d) {
                return xPos(d) + x / 2.5;
            })
            .attr("y", function(d) {
                return snapPos(d) - 2;
            })
            .attr('font-size', '9px')
            .text(OverviewChart.launches);
        var legend = chart.append('g')
            .attr('class', 'legend')
            .attr('height', 100)
            .attr('width', 100)
            .attr('transform', 'translate(-100,0)');
        var legendElements = legend.selectAll('g')
            .data(categories)
            .enter()
            .append('g');
        legendElements.append('text')
            .attr('x', 20)
            .attr('y', function(d, i) {
                return i * 20 + 60;
            })
            .attr('font-size', '12px')
            .text(function(d) {
                return d;
            });
        legendElements.append('rect')
            .attr('class', 'legendrect')
            .attr('x', 0)
            .attr('y', function(d, i) {
                return i * 20 + 50;
            })
            .attr('width', 15)
            .attr('height', 15)
            .style('fill', chartColour);
    },

    submissions: function(d) {
        return d.submissions;
    },
    snapshots: function(d) {
        return d.snapshots;
    },
    launches: function(d) {
        return d.launches;
    }

}
