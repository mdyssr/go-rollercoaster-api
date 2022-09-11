package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Coaster struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	ID           string `json:"id"`
	InPark       string `json:"inPark"`
	Height       int    `json:"height"`
}

type coasterHandlers struct {
	store map[string]Coaster
}

type adminPortal struct {
	password string
}

func main() {
	coasterHandlers := newCoastserHandlers()
	admin := newAdminPotal()
	http.HandleFunc("/coasters", coasterHandlers.coasters)
	http.HandleFunc("/coasters/", coasterHandlers.getCoaster)
	http.HandleFunc("/admin", admin.handler)

	log.Println("server started at http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func (h *coasterHandlers) coasters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
		return
	case http.MethodPost:
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
		return
	}
}

// get all coasters
func (h *coasterHandlers) get(w http.ResponseWriter, r *http.Request) {
	// slice with len(h.store) length
	coasters := make([]Coaster, len(h.store))

	i := 0
	for _, coaster := range h.store {
		coasters[i] = coaster
		i++
	}

	jsonBytes, err := json.Marshal(coasters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

// post a coaster
func (h *coasterHandlers) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(http.StatusText(http.StatusUnsupportedMediaType)))
		return
	}

	var coaster Coaster
	err = json.Unmarshal(bodyBytes, &coaster)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())

	h.store[coaster.ID] = coaster

}

func (h *coasterHandlers) getRandomCoaster(w http.ResponseWriter, r *http.Request) {

	ids := make([]string, len(h.store))

	i := 0
	for id := range h.store {
		ids[i] = id
		i++
	}

	var target string
	if len(ids) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if len(ids) == 1 {
		target = ids[0]
	} else {
		rand.Seed(time.Now().UnixNano())
		target = ids[rand.Intn(len(ids))]
	}

	// redirect
	w.Header().Add("location", fmt.Sprintf("/coasters/%s", target))
	w.WriteHeader(http.StatusFound)
}

func (h *coasterHandlers) getCoaster(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if parts[2] == "random" {
		h.getRandomCoaster(w, r)
		return
	}

	coaster, ok := h.store[parts[2]]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(coaster)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

}

func newCoastserHandlers() *coasterHandlers {
	return &coasterHandlers{
		store: map[string]Coaster{
			"id1": Coaster{
				Name:         "Fury 325",
				Height:       99,
				ID:           "id1",
				InPark:       "Carwinds",
				Manufacturer: "B+M",
			},
		},
	}
}

func newAdminPotal() *adminPortal {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		panic("provide a password in an env variable")
	}

	adminPortal := &adminPortal{password}
	return adminPortal
}

func (a adminPortal) handler(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
		return
	}

	w.Write([]byte("<html><h1>Super Secret Admin Panel YYYAAAAAAYYYY</h1></html>"))
}
