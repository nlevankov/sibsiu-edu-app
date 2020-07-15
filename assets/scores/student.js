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
                    DisciplinesOptions += `<option value="`+ data.Result.Disciplines[i].Value +`" selected>`+ data.Result.Disciplines[i].Text +`</option>`;
                }

                var DisciplinesSelect = `<select name="disciplines_ids" class="form-control my-select" id="disciplines_ids" data-actions-box="true" multiple>
                `+ DisciplinesOptions +`
                </select>`;

                $("[for='disciplines_ids'] + div").replaceWith(DisciplinesSelect);
                $(".my-select").selectpicker(options);

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

    }
);
