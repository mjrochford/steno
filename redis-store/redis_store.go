package redis_store

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"

	"math/rand"
	"regexp"
)

type RedisStore struct {
	ctx context.Context
	db  *redis.Client
}

func Connect(addr string, pass string, n_db int) RedisStore {
	var ctx = context.Background()
	var db = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       n_db,
	})

	return RedisStore{ctx: ctx, db: db}
}

func (store RedisStore) Push(user_id string, quote string) error {
	key := fmt.Sprintf("%s:quotes", user_id)

	_, err := store.db.RPush(store.ctx, key, quote).Result()
	return err
}

func (store RedisStore) Rm(user_id string, quote string) error {
	key := fmt.Sprintf("%s:quotes", user_id)

	_, err := store.db.LRem(store.ctx, key, 0, quote).Result()
	return err

}

func (store RedisStore) Search(user_id string, pattern string) ([]string, error) {
	list, err := store.GetAll(user_id)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(fmt.Sprintf(".*%s.*", pattern))
	if err != nil {
		return nil, err
	}

	out_list := make([]string, 0, len(list))
	for _, quote := range list {
		if re.Match([]byte(quote)) {
			out_list = append(out_list, quote)
		}
	}

	return out_list, nil
}

func (store RedisStore) GetAll(user_id string) ([]string, error) {
	key := fmt.Sprintf("%s:quotes", user_id)

	quotes, err := store.db.LRange(store.ctx, key, 0, -1).Result()
	if err == redis.Nil || len(quotes) == 0 {
		return nil, errors.New(fmt.Sprintf("No quotes for user_id: %s", user_id))
	} else if err != nil {
		return nil, err
	}

	return quotes, nil
}

func (store RedisStore) GetRandom(user_id string) (string, error) {
	quote_list, err := store.GetAll(user_id)
	if err != nil {
		return "", err
	}

	choice := rand.Intn(len(quote_list))
	return quote_list[choice], nil
}
