package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"html/template"
	"regexp"
	"errors"
)

type Page struct {
	Title string
	Body  []byte
}

var (
	templates = template.Must(template.ParseFiles("templ/edit.html", "templ/view.html"))
	validPaths = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
) 

func fileName(title string) string {
	return "data/page-" + title + ".wiki"
}

func templatePath(name string) string {
	return name + ".html"
}

func (p *Page) save() error {
	filename := fileName(p.Title)
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPaths.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil
}

func renderTemplate(w http.ResponseWriter, tempName string, p *Page) {
	err := templates.ExecuteTemplate(w, templatePath(tempName), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadPage(title string) (*Page, error) {
	filename := fileName(title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/" + title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func makeHanlder(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title, err := getTitle(w, r)
		if err != nil { return }
		fn(w, r, title)
	}
}

func main() {
	http.HandleFunc("/view/", makeHanlder(viewHandler))
	http.HandleFunc("/edit/", makeHanlder(editHandler))
	http.HandleFunc("/save/", makeHanlder(saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
