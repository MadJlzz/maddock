package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
)

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all good"))
}

func handleExecuteRecipe(w http.ResponseWriter, r *http.Request) {
	f, _, err := r.FormFile("recipe")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer f.Close()

	var recipe Recipe
	if err = yaml.NewDecoder(f).Decode(&recipe); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(recipe)
}
