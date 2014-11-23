$(function() {
    var path = $("#messages").attr("data-exec-url");
    url = 'ws://' + window.location.host + path;
    console.log('connecting to:' + url);
    var ws = new WebSocket(url);
    ws.onmessage = function(e) {
        var data = JSON.parse(e.data);
        $('#messages pre').append($('<span>').addClass(data.messageType).text(data.line));
    };
});
