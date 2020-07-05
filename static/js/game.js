
$(document).ready(function () { 

    // Makes a websocket connection
    let conn;
    function connect () {
        
        if (window['WebSocket']) {
    
            this.conn = new WebSocket('ws://' + document.location.host + '/ws');
    
            this.conn.onopen = evt => {

                console.log('Websocket opened..', evt);
                $('#conn-result').html('<b>Connected to host.</b>');
            };
    
            this.conn.onerror = evt => {

                console.error('Websocket error..', evt.data);
                $('#conn-result').html('<b>Uh oh, something went wrong.</b>');
            };
    
            this.conn.onclose = evt => {

                console.log('Websocket closed..', evt);
                $('#conn-result').html('<b>Connection to host closed.</b>');
            };
    
            this.conn.onmessage = evt => {

                console.log('Message from host..', evt.data);
                $('#conn-result').html('<b>'+evt.data+'</b>');
            };
    
        } else {
    
            $('#conn-result').html('<b>Oh poop your browser does not support WebSockets.</b>');
        }
    }

    // Creates a new room
    $('#create-btn').click(function () {
        connect();
    });

    // Joins an existing room
    $('#join-btn').click(function () {
        connect();
    });
});