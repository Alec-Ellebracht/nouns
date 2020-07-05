
$(document).ready(function () { 

    var conn;

    if (window["WebSocket"]) {

        conn = new WebSocket("ws://" + document.location.host + "/ws");

        conn.onopen = function (evt) {
            console.log('on open..', evt);
            $('#conn-result').html("<b>Connected to host.</b>");
        };

        conn.onerror = function (evt) {
            console.error('on error..', evt.data);
            $('#conn-result').html("<b>"+evt.data+"</b>");
        };

        conn.onclose = function (evt) {
            console.log('on close..', evt);
            $('#conn-result').html("<b>Connection to host closed.</b>");
        };

        conn.onmessage = function (evt) {
            console.log('on message..', evt.data);
            $('#conn-result').html("<b>"+evt.data+"</b>");
        };

    } else {

        $('#conn-result').html("<b>Oh poop your browser does not support WebSockets.</b>");
    }

});