package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"steno/discord"
	"steno/httptools"
	"steno/quotestore"
)

func httpStringError(w http.ResponseWriter, err string, status int) {
	httpError(w, errors.New(err), status)
}

func httpError(w http.ResponseWriter, err error, status int) {
	http.Error(w, fmt.Sprintf("steno: %s", err.Error()), status)
	log.Printf("ERROR/steno/%s\n", err.Error())
}

func addQuotes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")
	quote, err := quotestore.QuoteFromJSON(r.FormValue("quote"))
	if err != nil {
		httpStringError(w, "invalid request, No quote provided", http.StatusBadRequest)
		return
	}

	err = stenoStore.Push(guildID, userID, quote)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `"{success: true}"`)
}

func removeQuotes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")
	quote, err := quotestore.QuoteFromJSON(r.FormValue("quote"))
	if err != nil {
		httpStringError(w, "invalid request, No quote provided", http.StatusBadRequest)
		return
	}

	err = stenoStore.Rm(guildID, userID, quote)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getQuotesForUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")
	log.Printf("Getting user quotes for guild:%s user:%s \n", guildID, userID)

	random := strings.Compare(strings.ToLower(r.FormValue("random")), "true") == 0
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = math.MaxInt64
	}

	searchStr := r.FormValue("search")

	var quotes []quotestore.Quote
	if len(searchStr) > 0 {
		quotes, err = stenoStore.Search(guildID, userID, searchStr)
	} else if random {
		var quote quotestore.Quote
		quote, err = stenoStore.GetRandom(guildID, userID)
		quotes = []quotestore.Quote{quote}
	} else {
		quotes, err = stenoStore.GetAll(guildID, userID)
	}

	limit = int(math.Min(float64(len(quotes)), float64(limit)))
	quotes = quotes[0:limit]

	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}

	if len(quotes) <= 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `"{success: false}"`)
		return
	}

	quotesJSON, err := json.Marshal(quotes)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
	}
	fmt.Fprint(w, string(quotesJSON))
}

func discordRequest(method, url, auth string) ([]byte, error) {
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Authorization", auth)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}

	out, _ := ioutil.ReadAll(resp.Body)
	return out, nil
}

func authenticate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) bool {
	authorization := r.Header["Authorization"]
	if len(authorization) == 0 {
		httpStringError(w, "invalid request, No Authorization", http.StatusBadRequest)
		return false
	}
	auth := authorization[0]
	tokenType := strings.Split(auth, " ")[0] // Bearer ...
	// token := strings.Split(auth, " ")[1]   // ... {token}

	if tokenType != "Bot" {
		httpStringError(w, "invalid request, Bad token", http.StatusBadRequest)
		return false
	}

	// Error conditions for http.NewRequest are
	// - Invalid method (taken care of using http.MethodGet)
	// - Nil context (using background context)
	// - Invalid Url (assume url is valid)
	// - Invalid Body (body is nil here)
	// [see here](https://go.googlesource.com/go/+/go1.16.2/src/net/http/request.go#853)
	respBody, err := discordRequest(http.MethodGet,
		"https://discord.com/api/v8/oauth2/applications/@me", auth)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return false
	}
	var discordApp discord.Application
	err = json.Unmarshal(respBody, &discordApp)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return false
	}

	return true
}

var stenoStore quotestore.RedisStore
var httpClient *http.Client
func main() {

	log.SetOutput(os.Stdout)

	stenoStore = quotestore.Connect(os.Getenv("STENO_REDIS_ADDR"), "", 0)
	// stenoStore.LoadSavedData("")

	httpClient = &http.Client{}
	baseRoute := httptools.RouteNew().Log().Gate(authenticate)

	router := httprouter.New()
	router.GET("/quotes/:guild_id/:user_id", baseRoute.Clone().Finish(getQuotesForUser))
	router.POST("/quotes/:guild_id/:user_id", baseRoute.Clone().Finish(addQuotes))
	router.DELETE("/quotes/:guild_id/:user_id", baseRoute.Clone().Finish(removeQuotes))

	log.Fatal(http.ListenAndServe(":8080", router))
}
