package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	"steno/httptools"
	"steno/redis-store"
)

type QuoteStore interface {
	GetAll(user_id string) ([]string, error)
	GetRandom(user_id string) (string, error)

	Search(user_id string, pattern string) ([]string, error)
	Push(user_id string, quote string) error
	Rm(user_id string, quote string) error
}

func addQuotes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	quote := r.FormValue("quote")
	if quote == "" {
		http.Error(w, "steno: invalid request, No quote provided", http.StatusBadRequest)
		return
	}

	err := steno_store.Push(id, quote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `"{success: true}"`)
}

func removeQuotes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	quote := r.FormValue("quote")
	/* on http DELETE this will ignore values in the body
	 * so this must be provided in the query params I don't like this
	 */

	if quote == "" {
		http.Error(w, "steno: invalid request, No quote provided", http.StatusBadRequest)
		return
	}

	err := steno_store.Rm(id, quote)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `"{success: true}"`)
}

func getQuotesForUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	log.Println("Getting user quotes for id:", id)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(quotes) <= 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `"{success: false}"`)
		return
	}

	quotes_json, err := json.Marshal(quotes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprint(w, string(quotes_json))
}

func authenticate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) bool {
	authorization := r.Header["Authorization"]
	if len(authorization) == 0 {
		http.Error(w, "steno: invalid request, No Authorization", http.StatusBadRequest)
		return false
	}
	auth := authorization[0]
	token_type := strings.Split(auth, " ")[0] // Bearer ...
	// token := strings.Split(auth, " ")[1]   // ... {token}

	if token_type != "Bot" {
		http.Error(w, "steno: invalid request, Bad token", http.StatusBadRequest)
		return false
	}

	// Error conditions for http.NewRequest are
	// - Invalid method (taken care of using http.MethodGet)
	// - Nil context (using background context)
	// - Invalid Url (assume url is valid)
	// - Invalid Body (body is nil here)
	// [see here](https://go.googlesource.com/go/+/go1.16.2/src/net/http/request.go#853)
	discord_req, _ := http.NewRequest(http.MethodGet, "https://discord.com/api/v8/oauth2/applications/@me", nil)
	discord_req.Header.Add("Authorization", auth)

	resp, err := http_client.Do(discord_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var json_resp DiscordResponse
	json.Unmarshal(body, &json_resp)
	log.Println(json_resp)

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, string(body), resp.StatusCode)
		log.Printf("Access Denied from %s: %s\n", r.RemoteAddr, string(body))
		return false
	}

	log.Printf("discord reponse status %s", resp.Status)
	return true
}

var steno_store QuoteStore
var http_client *http.Client

func main() {
	log.SetOutput(os.Stdout)
	steno_store = redis_store.Connect(os.Getenv("STENO_REDIS_ADDR"), "", 0)

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}
	http_client = &http.Client{Transport: tr}

	base_route := httptools.RouteNew().Log().Gate(authenticate)

	router := httprouter.New()
	router.GET("/quotes/:id", base_route.Clone().Finish(getQuotesForUser))
	router.POST("/quotes/:id", base_route.Clone().Finish(addQuotes))
	router.DELETE("/quotes/:id", base_route.Clone().Finish(removeQuotes))

	log.Fatal(http.ListenAndServe(":8080", router))
}
