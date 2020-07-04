
$(document).ready(function () { 

    $('#guess-btn').click(function () {
        doGuess();
    });

    function doGuess() {
        $.ajax({
            url: 'guess',
            type: 'post',
            dataType: 'html',
            data : { guess: 'hello'},
            success : function(data) {
                // $('#guess-result').html('<div class="uk-alert-primary" uk-alert ><a class="uk-alert-close" uk-close></a><p>Your guess is '+data+'</p></div>');
                // UIkit.alert($('#guess-result').first(), {duration: 150, animation: true});
                let resultContainer = document.getElementById('guess-result');
                resultContainer.innerHTML = '<a class="uk-alert-close" uk-close></a><p>Your guess is '+data+'</p></div>';
                UIkit.alert(resultContainer, {duration: 150, animation: true});
            },
        });
    }
});
