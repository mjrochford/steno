package quotestore

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"

	"steno/discord"
)

type Quote struct {
	ID       string `json:"id"`
	AuthorID string `json:"author_id"`
	Str      string `json:"str"`

	Date           string `json:"date"`            // optional
	StenographerID string `json:"stenographer_id"` // optional
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
	return Quote{
		ID:             uuid.NewString(),
		Date:           ISO8601Date(time.Now()),
		StenographerID: stenographer.ID,
		AuthorID:       author.ID,
	}
}

func QuoteFromJSON(j []byte) (Quote, error) {
	var q Quote
	err := json.Unmarshal(j, &q)
	if err != nil {
		return Quote{}, err
	}

	if q.Str == "" {
		return Quote{}, errors.New("no quote string provided")
	}

	// TODO assert that AuthorID and StenographerID are actually discord users
	if q.Date == "" {
		q.Date = ISO8601Date(time.Now())
	}

	if q.ID == "" {
		q.ID = uuid.NewString()
	}

	return q, nil
}

func QuoteFromReader(r io.Reader) (Quote, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return Quote{}, err
	}
	return QuoteFromJSON(buf)
}

type QuoteStore interface {
	GetAll(guildID, userID string) ([]Quote, error)
	GetRandom(guildID, userID string) (Quote, error)
	Search(guildID, userID, pattern string) ([]Quote, error)

	Push(guildID, userID string, quote Quote) error
	Rm(guildID, userID string, quote Quote) error
}
