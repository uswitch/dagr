$(function() {
    var path = $("#messages").attr("data-exec-url");
    url = 'ws://' + window.location.host + path;
    console.log('connecting to:' + url);
    var ws = new WebSocket(url);
    ws.onmessage = function(e) {
	      $('#messages pre').text($('#messages pre').text() + e.data);
    };
});
