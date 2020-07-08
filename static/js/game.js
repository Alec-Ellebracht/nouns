
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
                
                $('.start-btn').hide();

                let envelope = JSON.parse(evt.data);
                let data = envelope.Msg;

                console.log('this',data);

                if (envelope.Type === 'hint') {

                    $('#noun-type').html(data.Noun.Type);
                    $('#noun-hint').html('It\'s ' + data.Hint);
                    $('#latest-hint').prop('hidden', false);
                }
                else if (envelope.Type === 'guess') {

                    console.log(data);

                    let guess = '<span class="uk-badge uk-padding-small">'+data.Guess+'</span>';
                    $('#guess-list').append(guess);
                }

                // UIkit.modal.alert(evt.data).then(function () {

                //     console.log('Message from host..', evt.data);
                // });
            };
    
        } else {
    
            $('#conn-result').html('<b>Oh poop your browser does not support WebSockets.</b>');
        }
    }

    connect();

    // button handler
    UIkit.util.on('#start-btn', 'click', function (event) {

        event.preventDefault();
        event.target.blur();

        $('.start-btn').hide();
        start();
    });

    function start() {

        let envelope = JSON.stringify(
        {
            type: "start",
            msg: {},
        });

        this.conn.send(envelope);
    }

});