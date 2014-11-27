$(function() {
    var messagesWs = new WebSocket('ws://' + window.location.host + $("#messages").attr("data-socket-path"));
    messagesWs.onmessage = function(e) {
        var data = JSON.parse(e.data);
        $('#messages pre').append($('<span>').addClass(data.messageType).text(data.line));
    };
    var executionWs = new WebSocket('ws://' + window.location.host + $("#execution").attr("data-socket-path"));
    executionWs.onmessage = function(e) {
        var data = JSON.parse(e.data);
        if (data.executionId == $("#execution").data('execution-id')) {
            $('#execution .execution-status div').replaceWith("<div class='" + data.executionStatus + "'>" + data.executionStatusLabel + "</div>");
        }
    };
});
