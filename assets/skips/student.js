$( document ).ready( function() {

    $( "#year" ).on( "change", function( event ) {

        var queryArr = $( "#group_id" ).serializeArray();
        queryArr = queryArr.concat($( event.target ).serializeArray());

        var jqXHR = $.getJSON('/api/filter/skips', queryArr);

        jqXHR.done(function (data) {

            if (data.Error == "") {
                var SemestersOptions = "";
                var Result = data.Result;

                if (Result.Semesters == null) {
                    SemestersOptions += `<option>Нет данных</option>`;
                } else {
                    for(var i = 0; i < Result.Semesters.length; i++) {
                        SemestersOptions += `<option>`+ Result.Semesters[i].Text +`</option>`;
                    }

                }
                $("#semester").html(SemestersOptions);

            } else {
                showAlert("div.panel-info", data.Error, AlertLvlError);
            }

        });
    });


    $( "#group_id" ).on( "change", function( event ) {
        window.location.replace("/skips?" + $( "#group_id" ).serialize());
    });

    }
);
