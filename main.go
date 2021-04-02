package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"

	"steno/discord"
	"steno/httptools"
	"steno/quotestore"
)

func isJSONContent(contentType []string) bool {
	for _, str := range contentType {
		if !strings.Contains(str, "application/json") && !strings.Contains(str, "text/json") {
			return false
		}
	}
	return true
}

var stenoStore quotestore.RedisStore
var httpClient *http.Client

/** Handler for adding quotes to the store
 * @url_param guild_id
 * @url_param user_id
 *
 * @body json encoded quote object
NOTE:
 * Must set Content-Type header in order for the data to be read
*/
func addQuotes(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")
	if !isJSONContent(r.Header["Content-Type"]) {
		return http.StatusBadRequest, errors.New("expected json body")
	}

	quote, err := quotestore.QuoteFromReader(r.Body)
	r.Body.Close()

	if quote.AuthorID == "" {
		quote.AuthorID = userID
	}

	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid request, %s", err)
	}

	err = stenoStore.Push(guildID, userID, quote)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("add quote failed/%s", err)
	}

	return http.StatusOK, nil
}

/**
 * Handler for the removing quotes from the store
 *
 * @url_param guild_id string
 * @url_param user_id string
NOTE:
 * redisstore requires the consumer of the api to provide json that will encode and then decode
 * and match with the json stored in the database, does not just match quote.ID which it probably
 * should
*/
func removeQuotes(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")
	quote, err := quotestore.QuoteFromReader(r.Body)
	r.Body.Close()
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid request, %s", err)
	}

	err = stenoStore.Rm(guildID, userID, quote)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("rm quote failed/%s", err)
	}

	return http.StatusOK, nil
}

/**
 * Handler for the retriveing information about stored quotes
 * @url_param guild_id string
 * @url_param user_id string
 *
 * @query_params search string a string to search the users quotes for
 * @query_params limit uint limit the amount of results that can be returned default 100
 * @query_params random bool only return one random quote if true
 */
func getQuotesForUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error) {
	userID := ps.ByName("user_id")
	guildID := ps.ByName("guild_id")

	random := strings.Compare(strings.ToLower(r.FormValue("random")), "true") == 0
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		limit = 100
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

/** Wrapper function for making a request to the discord api
 *  @param method http.Method
 *  @param endpoint string discord api endpoint to request
 *  @param auth discord Authorization header
		of the form "Bearer {token}" or "Bot {token}"
*/
func discordRequest(method, endpoint, auth string) ([]byte, error) {
	const DISCORDBASEURI = "https://discord.com/api/v8"
	u, _ := url.Parse(DISCORDBASEURI)
	u.Path = path.Join(u.Path, endpoint)

	// Error conditions for http.NewRequest are
	// - Invalid method (taken care of using http.MethodGet)
	// - Nil context (using background context)
	// - Invalid Url (assume url is valid)
	// - Invalid Body (body is nil here)
	// [see here](https://go.googlesource.com/go/+/go1.16.2/src/net/http/request.go#853)
	req, _ := http.NewRequest(method, u.String(), nil)
	req.Header.Add("Authorization", auth)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("resp %s", resp.Status)
	}

	out, _ := ioutil.ReadAll(resp.Body)
	return out, nil
}

func hasGuildAccess(guildID, auth string) bool {
	respBody, err := discordRequest(http.MethodGet, "/users/@me/guilds", auth)
	if err != nil {
		log.Printf("ERROR: discord request failed %s", err)
		return false
	}
	var guilds []discord.Guild
	err = json.Unmarshal(respBody, &guilds)
	if err != nil {
		log.Printf("ERROR: discord request parse failed %s", err)
		return false
	}

	for _, guild := range guilds {
		if guild.ID == guildID {
			return true
		}
	}
	return false
}

/** Handler for authenticating a particular request against the discord api
 *
 *  @url_param guild_id string guildID that is being queried
 *
 *  @header Authorization discord Authorization header
		of the form "Bearer {token}" or "Bot {token}"
*/
func authenticate(_ http.ResponseWriter, r *http.Request, ps httprouter.Params) (int, error) {
	guildID := ps.ByName("guild_id")

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

	// respBody, err := discordRequest(http.MethodGet, "/oauth2/applications/@me", auth)
	// if err != nil {
	// 	return http.StatusInternalServerError, fmt.Errorf("discord request failed: %s", err)
	// }
	// var discordApp discord.Application
	// err = json.Unmarshal(respBody, &discordApp)
	// if err != nil {
	// 	return http.StatusInternalServerError, fmt.Errorf("disord request parsing failed: %s", err)
	// }

	if !hasGuildAccess(guildID, auth) {
		return http.StatusForbidden,
			errors.New("discord bearer token does not have access to that guild")
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
