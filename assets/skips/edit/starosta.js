$( document ).ready( function() {

    $( "#date" ).on( "change", function( event ) {

        var queryArr = $( "#group_id" ).serializeArray();
        queryArr = queryArr.concat($( event.target ).serializeArray());

        var jqXHR = $.getJSON('/api/filter/skips/edit', queryArr);

        jqXHR.done(function (data) {

            if (data.Error == "") {
                var DisciplinesOptions = "";

                if (data.Result == null) {
                    DisciplinesOptions += `<option>Нет данных</option>`;
                } else {
                    for(var i = 0; i < data.Result.Disciplines.length; i++) {
                        DisciplinesOptions += `<option value="`+ data.Result.Disciplines[i].Value +`">`+ data.Result.Disciplines[i].Text +`</option>`;
                    }
                }

                $("#discipline_info").html(DisciplinesOptions);
            } else {
                showAlert("div.panel-info", data.Error, AlertLvlError);
            }

        });
    });


    $( "#group_id" ).on( "change", function( event ) {
        window.location.replace("/skips?" + $( "#group_id" ).serialize());
    });


    $( "#result_btn" ).on( "click", function( event ) {
        event.preventDefault();

        var jqXHR = $.post('/api/skips/update', $("#result_form").serialize(), null, "json");

        jqXHR.done(function (data) {

            if (data.Error == "") {
                showAlert("div.panel-info", data.Result, AlertLvlSuccess);
            } else {
                showAlert("div.panel-info", data.Error, AlertLvlError);
            }

        });

    });

    }
);
