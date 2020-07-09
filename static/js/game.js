
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

                let envelope = JSON.parse(evt.data);
                handleMessage(envelope);
            };
    
        } else {
    
            $('#conn-result').html('<b>Oh poop your browser does not support WebSockets.</b>');
        }
    }

    // start the socket connection
    connect();

    // button handler for the start button
    UIkit.util.on('#start-btn', 'click', function (event) {

        event.preventDefault();
        event.target.blur();

        $('.start-btn').hide();
        $('#loading-spinner').show();
        setTimeout(() => {

            let envelope = JSON.stringify(
                {
                    type: "start",
                    msg: {},
                });
        
            sendEnvelope(envelope);
            $('#loading-spinner').hide();

        }, 2000);

    });

    // button handler for the player input text box
    UIkit.util.on('#player-send-btn', 'click', function (event) {

        event.preventDefault();
        event.target.blur();

        let element = $('#player-input-box');

        let envelope = JSON.stringify(
            {
                type: "guess",
                msg: {
                    guess: element.val(),
                },
            });
    
        sendEnvelope(envelope);
        element.val('');
    });

    // binds the enter key to the player input text box
    $(document).keypress(function(e){
        if (e.which == 13){
            $("#player-send-btn").click();
        }
    });

    // handles all the incoming socket messages
    function handleMessage(envelope) {

        let data = envelope.Msg;
        console.log('~~~ envelope',envelope);
        console.log('~~~ data',data);

        switch (envelope.Type) {

            case 'hint':

                $('#noun-type').html(data.Noun.Type);
                $('#noun-hint').html(data.Hint);
                $('#latest-hint').prop('hidden', false);
                break;

            case 'guess':

                let guess = '<span class="uk-badge uk-padding-small">'+data.Guess+'</span><br><br>';
                $('#guess-list').prepend(coolGuess);
                break;

            case 'player':

                UIkit.notification({
                    message: data.msg,
                    status: 'primary',
                    pos: 'top-right',
                    timeout: 5000
                });
                break;

            case 'start':

                $('.start-btn').hide();
                break;

            default: 
            
                console.log('~~~ hmm something unexpected happened..');
          }

        // UIkit.modal.alert(evt.data).then(function () {

        //     console.log('Message from host..', evt.data);
        // });
    }

    // starts the game
    function sendEnvelope(envelope) {

        this.conn.send(envelope);
    }

});