$(function() {
    var url = $("#messages").attr("data-exec-url");
    console.log(url);
    var ws = new WebSocket(url);
    ws.onmessage = function(e) {
	      $('#messages pre').text($('#messages pre').text() + e.data);
    };
});
