package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"io/ioutil"
	"math"
	"time"
	"net/http"
	"os"
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

type RouteHandle func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) bool
type Route struct {
	handlers []RouteHandle
}

func New() Route {
	return Route{
		handlers: make([]RouteHandle, 0, 4),
	}
}

func (rt Route) Handle() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		for _, h := range rt.handlers {
			if !h(w, r, ps) {
				break
			}
		}
	}
}

func (rt Route) Apply(handler httprouter.Handle) Route {
	rt.handlers = append(rt.handlers,
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) bool {
			handler(w, r, ps)
			return true
		})
	return rt
}

func (rt Route) Gate(handler RouteHandle) Route {
	rt.handlers = append(rt.handlers, handler)
	return rt
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
	fmt.Fprintf(w, `{"error": "%s", "success": false}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
}

func addQuotes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	quote := r.FormValue("quote")
	if quote == "" {
		writeError(w, errors.New("steno: invalid request, No quote provided"), http.StatusBadRequest)
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
	/* on http DELETE this will ignore values in the body
	 * so this must be provided in the query params I don't like this
	 */

	if quote == "" {
		writeError(w, errors.New("steno: invalid request, No quote provided"), http.StatusBadRequest)
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

func httplog(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Printf("%s %s --- %s %s", r.UserAgent(), r.RemoteAddr, r.Method, r.URL)
}

func authenticate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) bool {
	auth := r.Header["Authorization"][0]
	token_type := strings.Split(auth, " ")[0] // Bearer ...
	// token := strings.Split(auth, " ")[1]      // ... {token}

	if token_type != "Bot" {
		writeError(w, errors.New("steno: invalid request, Bad token"), http.StatusBadRequest)
		return false
	}

	discord_req, err := http.NewRequest("GET", "https://discord.com/api/v8/oauth2/applications/@me", nil)
	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return false
	}
	discord_req.Header.Add("Authorization", auth)

	// log.Println(discord_req.Header["Authorization"])

	resp, err := http_client.Do(discord_req)

	if err != nil {
		writeError(w, err, http.StatusInternalServerError)
		return false
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		writeError(w, errors.New(string(body)), resp.StatusCode)
		log.Printf("Access Denied from %s: %s", r.RemoteAddr, errors.New(string(body)))
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
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig: nil,
	}
	http_client = &http.Client{Transport: tr}

	router := httprouter.New()
	router.GET("/quotes/:id", New().Apply(httplog).Gate(authenticate).Apply(getQuotesForUser).Handle())
	router.POST("/quotes/:id", New().Apply(httplog).Apply(addQuotes).Handle())
	router.DELETE("/quotes/:id", New().Apply(httplog).Apply(removeQuotes).Handle())

	log.Fatal(http.ListenAndServe(":8080", router))
}
