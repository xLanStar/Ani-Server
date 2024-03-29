package fetcher

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Data struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type Poster struct {
	Data Data
	Url  string
}

func (poster *Poster) Post() ([]byte, error) {
	jsonValue, _ := json.Marshal(poster.Data)

	resp, err := http.Post(poster.Url, "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(resp.Body)

	return bytes, err
}
