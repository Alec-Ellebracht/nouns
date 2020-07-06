
$(document).ready(function () { 

    // Makes a websocket connection
    let conn;
    function connect () {

        if (conn) { return; }
        
        if (window['WebSocket']) {
    
            let room = window.location.pathname.split('/').slice(-1).pop();
            console.log(room);

            this.conn = new WebSocket('ws://' + document.location.host + '/ws/' + room);
    
            this.conn.onopen = evt => {

                console.log('Websocket opened..', evt);
                $('#conn-result').html('<i>Connected to host.</i>');
            };
    
            this.conn.onerror = evt => {

                console.error('Websocket error..', evt.data);
                $('#conn-result').html('<i>Uh oh, something went wrong.</i>');
            };
    
            this.conn.onclose = evt => {

                console.log('Websocket closed..', evt);
                $('#conn-result').html('<i>Connection to host closed.</i>');
            };
    
            this.conn.onmessage = evt => {

                console.log('Message from host..', evt.data);
                $('#conn-result').html('<i>'+evt.data+'</i>');
            };
    
        } else {
    
            $('#conn-result').html('<b>Oh poop your browser does not support WebSockets.</b>');
        }
    }

    connect();

    // // Creates a new room
    // $('#create-btn').click(function () {
    //     connect();
    // });

    // // Joins an existing room
    // $('#join-btn').click(function () {
    //     connect();
    // });
});