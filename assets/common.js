$( document ).ready( function() {
        $('.my-select').selectpicker(options);
    }
);

const options = {
    selectedTextFormat: 'count > 3',
    countSelectedText: (CountOfSelectedOptions, CountOfTotalOptions) => (CountOfSelectedOptions == CountOfTotalOptions) ? 'Все' : 'Выбрано: ' + CountOfSelectedOptions,
    noneSelectedText: 'Ничего не выбрано',
    selectAllText: 'Выбрать всё',
    deselectAllText: 'Снять выбор'};

const AlertLvlError   = "danger";
const AlertLvlWarning = "warning";
const AlertLvlInfo    = "info";
const AlertLvlSuccess = "success";
const AlertMsgGeneric = "Something went wrong. Please try again, and contact us if the problem persists.";


function showAlert(selector, msg, level) {

    $( "div.alert" ).remove();

    var alertHTML = `<div class="alert alert-`+ level +` alert-dismissible" role="alert" style="display: none;">
    <button type="button" class="close" data-dismiss="alert"
            aria-label="Close">
      <span aria-hidden="true">&times;</span>
    </button>
    `+ msg + `
  </div>`;

    $( selector ).before(alertHTML);
    $( "div.alert" ).slideDown("fast", "linear");

}

