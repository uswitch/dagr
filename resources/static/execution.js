$(function() {
    var path = $("#messages").attr("data-socket-path");
    url = 'ws://' + window.location.host + path;
    var ws = new WebSocket(url);
    ws.onmessage = function(e) {
        var data = JSON.parse(e.data);
        $('#messages pre').append($('<span>').addClass(data.messageType).text(data.line));
    };
});