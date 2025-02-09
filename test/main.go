package main

import (
	"fmt"
	"net/http"

	"github.com/mekramy/gofs"
	"github.com/mekramy/gotemplate"
)

func main() {
	// Initialize tempalate
	fs := gofs.NewDir("./assets")
	tpl := gotemplate.New(
		fs,
		gotemplate.WithRoot("views"),
		gotemplate.WithPartials("views/partials"),
	)

	if err := tpl.Load(); err != nil {
		fmt.Println(err)
		return
	}

	// Handle requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := tpl.Render(w, "pages/home", nil, "layout"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})

	http.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		if err := tpl.Render(w, "pages/contact", nil, "layout"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})

	http.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		if err := tpl.Render(w, "errors", nil); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	})

	fmt.Println("Starting server at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println(err)
	}

}
