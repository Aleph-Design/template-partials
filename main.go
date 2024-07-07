package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	// "github.com/tsawler/page"
	"github.com/examples/page-use/page"
)

const PORT = ":8080"

type Data struct {
	Data map[string]any
}

func main() {

	render := page.Render{
		TemplateDir: "./templates",
		TemplateMap: make(map[string]*template.Template),
		Functions:   template.FuncMap{},
		Debug:       true,
		UseCache:    true,
	}

	// Call LoadLayoutsAndPartials to automatically load all such files found in TemplateDir.
	err := render.LoadLayoutsAndPartials([]string{".layout"})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]any)
		data["payload"] = "This is MY passed data."
		err := render.Show(w, "home.page.gohtml", &Data{Data: data})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Println(err)
			return
		}
	})

	http.HandleFunc("/string", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]any)
		data["payload"] = "This is passed data."
		out, err := render.String("home.page.gohtml", &Data{Data: data})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Println(err)
			return
		}
		log.Println(out)
		fmt.Fprint(w, "Check the console; you should see html")
	})

	log.Println("Starting on port", PORT)
	err = http.ListenAndServe(PORT, nil)
	if err != nil {
		log.Fatal(err)
	}
}