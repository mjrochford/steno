package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"steno/redis-store"
)

type QuoteStore interface {
	GetAll(user_id string) ([]string, error)
	GetRandom(user_id string) (string, error)

	Search(user_id string, pattern string) ([]string, error)
	Push(user_id string, quote string) error
	Rm(user_id string, quote string) error
}

func writeJson(w http.ResponseWriter, v interface{}) {
	marshaled_json, err := json.Marshal(v)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `{"success": true, "data": %s}`, string(marshaled_json))
}

func writeError(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error": "%s", "success": false}`, err)
}

func addQuotes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	quote := r.FormValue("quote")
	if quote == "" {
		writeError(w, errors.New("Invalid request"), http.StatusBadRequest)
		return
	}

	err := steno_store.Push(id, quote)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `"{success: true}"`)
}

func removeQuotes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	quote := r.FormValue("quote")
	if quote == "" {
		writeError(w, errors.New("Invalid request"), http.StatusBadRequest)
		return
	}

	err := steno_store.Rm(id, quote)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `"{success: true}"`)
}

func getQuotesForUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	random := strings.Compare(strings.ToLower(r.FormValue("random")), "true") == 0
	limit, string_err := strconv.Atoi(r.FormValue("limit"))
	if string_err != nil {
		limit = math.MaxInt64
	}

	search_string := r.FormValue("search")

	var quotes []string
	var err error
	if len(search_string) > 0 {
		quotes, err = steno_store.Search(id, search_string)
	} else if random {
		var quote string
		quote, err = steno_store.GetRandom(id)

		quotes = []string{quote}
	} else {
		quotes, err = steno_store.GetAll(id)
	}

	limit = int(math.Min(float64(len(quotes)), float64(limit)))
	quotes = quotes[0:limit]

	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if len(quotes) <= 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `"{success: false}"`)
		return
	}

	writeJson(w, quotes)
}

var steno_store QuoteStore

func main() {
	steno_store = redis_store.Connect("redis:6379", "", 0)
	router := httprouter.New()
	router.GET("/quotes/:id", getQuotesForUser)
	router.POST("/quotes/:id", addQuotes)
	router.DELETE("/quotes/:id", removeQuotes)

	log.Fatal(http.ListenAndServe(":8080", router))
}
