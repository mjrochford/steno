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

var stenoStore quotestore.RedisStore
var httpClient *http.Client

func addQuotes(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")
	quote, err := quotestore.QuoteFromJSON(r.FormValue("quote"))
	if err != nil {
		return http.StatusBadRequest, errors.New("invalid request, No quote provided")
	}

	err = stenoStore.Push(guildID, userID, quote)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("add quote failed/%s", err)
	}

	return http.StatusOK, nil
}

func removeQuotes(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")
	quote, err := quotestore.QuoteFromJSON(r.FormValue("quote"))
	if err != nil {
		return http.StatusBadRequest, errors.New("invalid request, No quote provided")
	}

	err = stenoStore.Rm(guildID, userID, quote)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("rm quote failed/%s", err)
	}

	return http.StatusOK, nil
}

func getQuotesForUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")

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
		return http.StatusInternalServerError, fmt.Errorf("get quotes failed/%s", err)
	}

	if len(quotes) <= 0 {
		return http.StatusNotFound, fmt.Errorf("no quotes for user/%s", err)
	}

	quotesJSON, err := json.Marshal(quotes)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("json marshal failed/%s", err)
	}

	fmt.Fprint(w, string(quotesJSON))
	return http.StatusOK, nil
}

func discordRequest(method, url, auth string) ([]byte, error) {
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Authorization", auth)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("discord requst not ok: %d", resp.StatusCode)
	}

	out, _ := ioutil.ReadAll(resp.Body)
	return out, nil
}

func authenticate(_ http.ResponseWriter, r *http.Request, _ httprouter.Params) (int, error) {
	authorization := r.Header["Authorization"]
	if len(authorization) == 0 {
		return http.StatusBadRequest, errors.New("invalid request, No Authorization")
	}
	auth := authorization[0]
	tokenType := strings.Split(auth, " ")[0] // Bearer ...
	// token := strings.Split(auth, " ")[1]   // ... {token}

	if tokenType != "Bot" {
		return http.StatusBadRequest, errors.New("invalid request, Bad token")
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
		return http.StatusInternalServerError, fmt.Errorf("discord request failed: %s", err)
	}
	var discordApp discord.Application
	err = json.Unmarshal(respBody, &discordApp)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("disord request parsing failed: %s", err)
	}

	return http.StatusOK, nil
}

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
