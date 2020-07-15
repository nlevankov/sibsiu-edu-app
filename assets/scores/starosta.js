$( document ).ready( function() {

    $( "#year,#semester" ).on( "change", function( event ) {

        var queryArr = $( "#group_id" ).serializeArray();

        if (event.target.id == "year") {
            queryArr = queryArr.concat($( event.target ).serializeArray());
        }

        if (event.target.id == "semester") {
            queryArr = queryArr.concat($( event.target ).serializeArray(), $( "#year" ).serializeArray());
        }

        var jqXHR = $.getJSON('/api/filter/scores', queryArr);

        jqXHR.done(function (data) {

            if (data.Error == "") {
                var DisciplinesOptions = "";

                for(var i = 0; i < data.Result.Disciplines.length; i++) {
                    DisciplinesOptions += `<option value="`+ data.Result.Disciplines[i].Value +`">`+ data.Result.Disciplines[i].Text +`</option>`;
                }

                $("#discipline_id").html(DisciplinesOptions);

                if (event.target.id == "year") {
                    var SemestersOptions = "";

                    for(var i = 0; i < data.Result.Semesters.length; i++) {
                        SemestersOptions += `<option>`+ data.Result.Semesters[i].Text +`</option>`;
                    }

                    $("#semester").html(SemestersOptions);
                }

            } else {
                showAlert("div.panel-info", data.Error, AlertLvlError);
            }

        });

    });


    $( "#group_id" ).on( "change", function( event ) {
        window.location.replace("/scores?" + $( "#group_id" ).serialize());
    });


    $( "#result_btn" ).on( "click", function( event ) {
        event.preventDefault();

        var jqXHR = $.post('/api/scores/update', $("#result_form").serialize(), null, "json");

        jqXHR.done(function (data) {

            if (data.Error == "") {
                showAlert("div.panel-info", data.Result, AlertLvlSuccess);
            } else {
                showAlert("div.panel-info", data.Error, AlertLvlError);
            }

        });

    });

    $( "#report_link" ).on( "click", function( event ) {
        event.preventDefault();
        window.location.replace("/scores/report?" + $( "#filter" ).serialize());
    });

    }
);
