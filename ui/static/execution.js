$(function() {  
  var messagesUrl = 'ws://' + window.location.host + $("#messages").attr("data-socket-path");
  var messagesWs = new WebSocket(messagesUrl);
  messagesWs.onmessage = function(e) {
    var data = JSON.parse(e.data);
    $('#messages pre').append($('<span>').addClass(data.messageType).text(data.line));
  };
  
  var executionUrl = 'ws://' + window.location.host + $("#execution").attr("data-socket-path");
  var executionWs = new WebSocket(executionUrl);
  executionWs.onmessage = function(e) {
    var data = JSON.parse(e.data);
    var currentExecutionId = $("#execution").data('execution-id');
    if (data.executionId == currentExecutionId) {
      var label = $('#execution .exec-status-label');
      label.replaceWith("<div class='exec-status-label " + data.executionStatus + "'>" + data.executionStatusLabel + "</div>");
    }
  };
});
