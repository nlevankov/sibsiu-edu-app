$( document ).ready( function() {


    $( "#result_btn" ).on( "click", function( event ) {
        event.preventDefault();

        var jqXHR = $.post('/api/groups/statuses/update', $("#result_form").serialize(), null, "json");

        jqXHR.done(function (data) {

            if (data.Error == "") {
                showAlert("div.panel-default", data.Result, AlertLvlSuccess);
            } else {
                showAlert("div.panel-default", data.Error, AlertLvlError);
            }

        });

    });

    }
);
