$( document ).ready( function() {

    $( "#year,#student_id" ).on( "change", function( event ) {

        var queryArr = $( "#group_id" ).serializeArray();

        if (event.target.id == "student_id") {
            queryArr = queryArr.concat($( event.target ).serializeArray());
        }

        if (event.target.id == "year") {
            queryArr = queryArr.concat($( event.target ).serializeArray(), $( "#student_id" ).serializeArray());
        }

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


                if (event.target.id == "student_id") {
                    var YearsOptions = "";

                    if (Result.Years == null) {
                        YearsOptions += `<option>Нет данных</option>`;
                    } else {
                        for(var i = 0; i < Result.Years.length; i++) {
                            YearsOptions += `<option value="`+ Result.Years[i].Value +`">`+ Result.Years[i].Text +`</option>`;
                        }

                    }
                    $("#year").html(YearsOptions);
                }

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
