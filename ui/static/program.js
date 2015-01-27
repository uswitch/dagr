$(function() {
    var path = $('#executions').attr("data-socket-path")        
    url = 'ws://' + window.location.host + path;
    var ws = new WebSocket(url);
    ws.onmessage = function(e) {
        var data = JSON.parse(e.data);
        var row = $('tr.execution-status[data-execution-id="' + data.executionId + '"]')[0]
        var timeElement = $("<a href='/executions/" + data.executionId + "'>" + data.executionTime + "</a>")
        var statusElement = $("<div class='exec-status-label " + data.executionStatus + "'>" + data.executionStatusLabel + "</div>");
        var lastOutputElement = $("<code>").text(data.executionLastOutput);
        
        if (row) {
            $(row).find('td.execution-time a').replaceWith(timeElement);
            $(row).find('td.execution-status div').replaceWith(statusElement);
            $(row).find('td.execution-last-output code').replaceWith(lastOutputElement);
        } else {
            row = $('<tr class="execution-status" data-execution-id="' + data.executionId + '">').appendTo('#executions tbody');
            timeElement.appendTo($('<td class="execution-time">').appendTo(row));
            statusElement.appendTo($('<td class="execution-status">').appendTo(row));
            lastOutputElement.appendTo($('<td class="execution-last-output">').appendTo(row));
        }
    };
});
