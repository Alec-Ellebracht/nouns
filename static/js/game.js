
$(document).ready(function () { 

    // Makes a websocket connection
    let conn;
    let reconnect = 0;
    function connect () {

        if (conn) { return; }
        
        if (window['WebSocket']) {
    
            let room = window.location.pathname.split('/').slice(-1).pop();
            this.conn = new WebSocket('ws://' + document.location.host + '/ws/' + room);
    
            this.conn.onopen = evt => {

                console.info('Websocket opened..', evt);
                $('#conn-result').html('<i>Connected to host.</i>');
            };
    
            this.conn.onerror = evt => {

                console.error('Websocket error..', evt.data);
                $('#conn-result').html('<i>Uh oh, something went wrong.</i>');
            };
    
            this.conn.onclose = evt => {

                console.info('Websocket closed..', evt);
                $('#conn-result').html('<i>Connection to host closed.</i>');

                if (evt.code == 1006 && reconnect <= 5) {
                    reconnect++;
                    setTimeout(function(){ connect(); }, 2000);
                    
                }
            };
    
            this.conn.onmessage = evt => {

                try {

                    let envelope = JSON.parse(evt.data);
                    handleMessage(envelope);
                }
                catch (err) {

                    console.error(err);
                }
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

        }, 2000);

    });

    // button handler for the player input text box
    UIkit.util.on('#player-send-btn', 'click', function (event) {

        event.preventDefault();
        event.target.blur();

        let element = $('#player-input-box');

        let envelope = JSON.stringify(
            {
                type: "message",
                body: { message: element.val() },
            });
            console.log(envelope);
    
        sendEnvelope(envelope);
        element.val('');
    });

    // handles all the incoming socket messages
    function handleMessage(envelope) {

        let data = envelope.body;
        console.log('~~~ envelope',envelope);
        console.log('~~~ data',data);

        switch (envelope.type) {

            case 'hint':

                $('#loading-spinner').hide();

                $('#noun-type').html(data.noun.type);
                $('#noun-hint').html(data.text);
                $('#latest-hint').prop('hidden', false);
                break;

            case 'guess':

                let side = false ? 'uk-float-right' : 'uk-float-left';
                let guess = '<div class="uk-badge uk-padding-small '+side+'">'
                    +data.text+'</div><br><br>';

                $('#guess-list').prepend(guess);

                if (data.isCorrect) {
                    
                    UIkit.notification({
                        message: 'The correct answer is "'+data.noun+'"',
                        status: 'primary',
                        pos: 'bottom-right',
                        timeout: 5000
                    });
                }

                break;

            case 'noun':
                
                UIkit.modal.alert('Your noun is "'+ data.text +'"');
                $('#current-noun').html(data.text);
                break;

            case 'action':

                let allBadges;
                $('#player-icons').empty();

                for (i in data.players) {

                    let player = data.players[i];

                    console.log(player);
    
                    let alias = 
                        player.name.slice(0, 2).toUpperCase();
    
                    // upate player badge
                    let playerBadge = 
                        '<div class="uk-icon-button uk-margin-small-left uk-margin-small-bottom" >'+alias+'</div>'; 

                    // allBadges += playerBadge;
                    $('#player-icons').append(playerBadge);
                }
                

                break;

            case 'start':

                $('.start-btn').hide();
                $('#loading-spinner').show();
                setRoundTimer(data.roundEnd);
                bindGameKeys();

                break;

            default: 
            
                console.log('~~~ hmm something unexpected happened..');
          }

    }

    // starts the game
    function sendEnvelope(envelope) {

        this.conn.send(envelope);
    }

    // sets the countdown timer for the round
    function setRoundTimer(roundEnd) {

        console.log(roundEnd);

        // Set the date we're counting down to
        let countDownDate = new Date(roundEnd).getTime() ;

        // Update the count down every 1 second
        let counter = setInterval(function() {

            let now = new Date().getTime();
            let distance = countDownDate - now;

            let minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
            let seconds = Math.floor((distance % (1000 * 60)) / 1000);
            let zero = seconds < 10 ? '0' : '';
                
            document.getElementById("current-time").innerHTML = minutes + ":" + zero + seconds;
                
            // stop once the timer is over
            if (distance < 0) {
                clearInterval(counter);
                document.getElementById("current-time").innerHTML = "Round Over";
            }
        }, 1000);
    }

    // bind game play keys
    function bindGameKeys() {

        // binds the enter key to the player input text box
        $(document).keypress(function(e){
            if (e.which == 13){
                $("#player-send-btn").click();
            }
        });
    }
});