package views

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gorilla/csrf"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"sibsiu.ru/context"
)

var (
	LayoutDir   string = "views/layouts/"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

type View struct {
	Template *template.Template
	Layout   string
}

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, layoutFiles()...)
	// We are changing how we create our templates, calling
	// New("") to give us a template that we can add a function to
	// before finally passing in files to parse as part of the template.
	t, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			// If this is called without being replace with a proper implementation
			// returning an error as the second argument will cause our template
			// package to return an error when executed.
			return "", errors.New("csrfField is not implemented")
		},
		"pathEscape": func(s string) string {
			return url.PathEscape(s)
		},
		// Once we have our template with a function we are going to pass in files
		// to parse, much like we were previously.
	}).ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	return &View{
		Template: t,
		Layout:   layout,
	}
}

// возможно не самое лучшее имя RenderJSON
// todo статус http ответа в случае ошибки не устанавливается отлиным от 200
func RenderJSON(w http.ResponseWriter, r *http.Request, result interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")

	var msg string
	if err != nil {
		// todo логика взята из func (d *Data) SetAlert(err error), возможно необходим рефакторинг
		if pErr, ok := err.(PublicError); ok {
			msg = pErr.Public()
		} else {
			log.Println(err)
			msg = AlertMsgGeneric
		}
	}

	enc := json.NewEncoder(w)
	d := map[string]interface{}{"Result": result, "Error": msg}
	enc.Encode(d)
}

// todo статус http ответа в случае ошибки не устанавливается отлиным от 200
//  не знаю, что делать в случае если методу поступает ошибка в каче-ве параметра
func RenderXLSX(w http.ResponseWriter, result *excelize.File) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=Отчет.xlsx")
	w.Header().Set("File-Name", "Отчет.xlsx")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")

	if err := result.Write(w); err != nil {
		log.Println(err)
	}
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")

	var vd Data
	switch d := data.(type) {
	case Data:
		// We need to do this so we can access the data in a var
		// with the type Data.
		vd = d
	default:
		// If the data IS NOT of the type Data, we create one
		// and set the data to the Yield field like before.
		vd = Data{
			Yield: data,
		}
	}

	// Lookup the alert and assign it if one is persisted
	if alert := getAlert(r); alert != nil {
		vd.Alert = alert
		clearAlert(w)
	}

	// Lookup and set the user to the User field
	vd.User = context.User(r.Context())

	var buf bytes.Buffer
	// We need to create the csrfField using the current http request.
	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		// We can also change the return type of our function, since we no longer
		// need to worry about errors.
		"csrfField": func() template.HTML {
			// We can then create this closure that returns the csrfField for
			// any templates that need access to it.
			return csrfField
		},
	})
	// Then we continue to execute the template just like before.
	err := tpl.ExecuteTemplate(&buf, v.Layout, vd)
	//log.Println(err.Error())
	if err != nil {
		http.Error(w, "Что-то пошло не так. Если проблема остается, напишите нам на электронную почту support@sibsiu.ru",
			http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)

}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add the new argument - r - here.
	v.Render(w, r, nil)
}

func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}
	return files
}

// addTemplatePath takes in a slice of strings
// representing file paths for templates, and it prepends
// the TemplateDir directory to each string in the slice
//
// Eg the input {"home"} would result in the output
// {"views/home"} if TemplateDir == "views/"
func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

// addTemplateExt takes in a slice of strings
// representing file paths for templates and it appends
// the TemplateExt extension to each string in the slice
//
// Eg the input {"home"} would result in the output
// {"home.gohtml"} if TemplateExt == ".gohtml"
func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}
