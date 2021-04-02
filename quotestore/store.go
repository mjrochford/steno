package quotestore

import (
	"time"
	"encoding/json"

	"github.com/google/uuid"

	"steno/discord"
)

type Quote struct {
	ID string `json:"id"`
	StenographerID string `json:"stenographer_id"`
	AuthorID string `json:"author_id"`

	Str string `json:"str"`
	Date string `json:"date"`
}

func (q Quote) MarshalBinary() ([]byte, error) {
	out, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (q *Quote) UnmarshalBinary(data []byte) error {
	err := json.Unmarshal(data, q)
	if err != nil {
		return err
	}
	return nil
}

func (q *Quote) String() string {
	return q.Str
}

func ISO8601Date(t time.Time) string {
	return t.Format(time.RFC3339)
}

func QuoteNew(author, stenographer discord.User, q string) Quote {
	return Quote {
		ID: uuid.NewString(),
		Date: ISO8601Date(time.Now()),
		StenographerID: stenographer.ID,
		AuthorID: author.ID,
	}
}

func QuoteFromJSON(j string) (Quote, error) {
	var q Quote
	err := json.Unmarshal([]byte(j), &q)
	if err != nil {
		return Quote{}, err
	}

	if q.ID == "" {
		q.ID = uuid.NewString()
	}

	return q, nil
}

type QuoteStore interface {
	GetAll(guildID, userID string) ([]Quote, error)
	GetRandom(guildID, userID string) (Quote, error)
	Search(guildID, userID, pattern string) ([]Quote, error)

	Push(guildID, userID string, quote Quote) error
	Rm(guildID, userID string, quote Quote) error
}

