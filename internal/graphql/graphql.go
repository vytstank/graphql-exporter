package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/vinted/graphql-exporter/internal/config"
)

var funcMap = template.FuncMap{
	"NOW": func(t string) (string, error) {
		d, err := time.ParseDuration(t)
		return time.Now().UTC().Add(d).Format(time.RFC3339), err
	},
}

func GraphqlQuery(ctx context.Context, query string) ([]byte, error) {
	type GraphQLRequest struct {
		Query string `json:"query"`
	}

	tpl, err := template.New("query").Funcs(funcMap).Parse(query)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	var templateBuffer bytes.Buffer
	err = tpl.Execute(&templateBuffer, nil)
	if err != nil {
		return nil, fmt.Errorf("template error %s", err)
	}

	reqBody, err := json.Marshal(GraphQLRequest{Query: templateBuffer.String()})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query to JSON: %v", err)
	}

	u, err := url.ParseRequestURI(config.Config.GraphqlURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing URL: %s", err)
	}

	urlStr := u.String()
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("HTTP request error: %s", err)
	}

	for _, header := range config.Config.GraphqlCustomHeaders {
		req.Header.Add(header.Key, header.Value)
	}
	r, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	slog.Info(fmt.Sprintf("http req params: %v", string(reqBody)))
	if r.StatusCode != 200 {
		return nil, fmt.Errorf(r.Status)
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	slog.Info(fmt.Sprintf("response body: %v", string(body)))
	return body, nil
}
