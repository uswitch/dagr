$(function() {
    $.each($("tr.program-status"), function (index, row) {
        var path = $(row).attr("data-socket-path")        
        url = 'ws://' + window.location.host + path;
        console.log('connecting to:' + url);
        var ws = new WebSocket(url);
        ws.onmessage = function(e) {
            var data = JSON.parse(e.data);
            $(row).find('td.execution-time a').replaceWith("<a href='/executions/" + data.executionId + "'>" + data.executionTime + "</a>");
            
            $(row).find('td.execution-status').empty();
            $("<div class='exec-status-label " + data.executionStatus + "'>" + data.executionStatusLabel + "</div>").appendTo($(row).find('td.execution-status'));
            var btn = $(row).find('button.program-run');
            if (data.executionStatus == "running") {
                btn.addClass("pure-button-disabled");
            } else {
                btn.removeClass("pure-button-disabled");
            }
            console.log(data)

            $('#succeeded h2').text($('td.execution-status div.succeeded').length)
            $('#retryable h2').text($('td.execution-status div.retryable').length)
            $('#failed h2').text($('td.execution-status div.failed').length)
        };
    });
});
