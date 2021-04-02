// Package quotestore redis backend for discord quote storage
package quotestore

// redis is a bad choice for how i have the data setup,
// i just used this because i wanted to see what redis is like

// redis could be used as a caching front layer to a more
// proper relational database to lessen the join queries on the db

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"regexp"

	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	ctx context.Context
	db  *redis.Client
}

func quotesURI(guildID, userID string) string {
	return fmt.Sprintf("%s:%s:quotes", guildID, userID)
}

func Connect(addr string, pass string, nDB int) RedisStore {
	var ctx = context.Background()
	var db = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       nDB,
	})

	return RedisStore{ctx: ctx, db: db}
}

func (store RedisStore) Push(guildID, userID string, quote Quote) error {
	uri := quotesURI(guildID, userID)

	quoteJSON, err := json.Marshal(quote)
	if err != nil {
		return err
	}

	_, err = store.db.RPush(store.ctx, uri, quoteJSON).Result()
	return err
}

func (store RedisStore) Rm(guildID, userID string, quote Quote) error {
	uri := quotesURI(guildID, userID)
	quoteJSON, err := json.Marshal(quote)
	if err != nil {
		return err
	}

	_, err = store.db.LRem(store.ctx, uri, 0, quoteJSON).Result()
	return err

}

func (store RedisStore) Search(guildID, userID, pattern string) ([]Quote, error) {
	// very unoptimal searching but redis is unoptimal for this application
	list, err := store.GetAll(guildID, userID)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(fmt.Sprintf(".*%s.*", pattern))
	if err != nil {
		return nil, err
	}

	outList := make([]Quote, 0, len(list))
	for _, quote := range list {
		if re.Match([]byte(quote.String())) {
			outList = append(outList, quote)
		}
	}

	return outList, nil
}

func quotesFromDB(quotes []string) []Quote {
	out := make([]Quote, 0, len(quotes))
	for _, q := range quotes {
		quote, err := QuoteFromJSON([]byte(q))
		if err == nil {
			out = append(out, quote)
		} else {
			log.Printf("redisstore: error unmarshaling json from db -- %s\n", err)
		}
	}
	return out
}

func (store RedisStore) GetAll(guildID, userID string) ([]Quote, error) {
	uri := quotesURI(guildID, userID)

	quotes, err := store.db.LRange(store.ctx, uri, 0, -1).Result()
	if err == redis.Nil || len(quotes) == 0 {
		return nil, fmt.Errorf("redisstore: No quotes for guildID:%s userID:%s", guildID, userID)
	} else if err != nil {
		return nil, err
	}

	return quotesFromDB(quotes), nil
}

func (store RedisStore) GetRandom(guildID, userID string) (Quote, error) {
	quoteList, err := store.GetAll(guildID, userID)
	if err != nil {
		return Quote{}, err
	}

	choice := rand.Intn(len(quoteList))
	return quoteList[choice], nil
}

func (store RedisStore) Import(data map[string][]Quote) {
	for user, quotes := range data {
		store.db.Del(store.ctx, user)
		for _, quote := range quotes {
			_, err := store.db.RPush(store.ctx, user, quote).Result()
			if err != nil {
				log.Printf("redisstore: error importing key [%s := %s] %s",
					user, quote, err)
			}
		}

	}
}

func (store RedisStore) WriteJSON(w io.Writer) {
	out := make(map[string][]Quote)
	keys, err := store.db.Keys(store.ctx, "*:quotes").Result()
	if err != nil {
		log.Fatalf("redis_store: err retriveing keys %s\n", err)
		return
	}

	for _, k := range keys {
		quotes, err := store.db.LRange(store.ctx, k, 0, -1).Result()
		if err != nil {
			log.Fatalf("redis_store: err reading key %s --%s\n", k, err)
			return
		}

		out[k] = quotesFromDB(quotes)
	}

	outBytes, err := json.Marshal(out)
	if err != nil {
		log.Fatalf("redis_store: err creating json obj %s\n", err)
		return
	}

	w.Write(outBytes)
}

func (store RedisStore) LoadSavedData(saveLocation string) {
	var savedData map[string][]Quote
	file, err := os.OpenFile(saveLocation, os.O_RDONLY, 0644)
	if err != nil {
		log.Printf("error opening saved file %s", saveLocation)
	}
	buf, _ := io.ReadAll(file)
	json.Unmarshal(buf, &savedData)

	store.Import(savedData)
}
