package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Yamashou/gqlgenc/graphqljson"
	"golang.org/x/xerrors"
)

type HTTPRequestOption func(req *http.Request)

type Client struct {
	Client             *http.Client
	BaseURL            string
	HTTPRequestOptions []HTTPRequestOption
}

// Request represents an outgoing GraphQL request
type Request struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

func NewClient(client *http.Client, baseURL string, options ...HTTPRequestOption) *Client {
	return &Client{
		Client:             client,
		BaseURL:            baseURL,
		HTTPRequestOptions: options,
	}
}

func (c *Client) newRequest(ctx context.Context, query string, vars map[string]interface{}, httpRequestOptions []HTTPRequestOption) (*http.Request, error) {
	r := &Request{
		Query:         query,
		Variables:     vars,
		OperationName: "",
	}

	requestBody, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("encode: %s", err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}

	for _, httpRequestOption := range c.HTTPRequestOptions {
		httpRequestOption(req)
	}
	for _, httpRequestOption := range httpRequestOptions {
		httpRequestOption(req)
	}

	return req, nil
}

// Post sends a http POST request to the graphql endpoint with the given query then unpacks
// the response into the given object.
func (c *Client) Post(ctx context.Context, query string, respData interface{}, vars map[string]interface{}, httpRequestOptions ...HTTPRequestOption) error {
	req, err := c.newRequest(ctx, query, vars, httpRequestOptions)
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return xerrors.Errorf(": %w", err)
	}
	defer resp.Body.Close()

	if err := graphqljson.Unmarshal(resp.Body, respData); err != nil {
		return xerrors.Errorf(": %w", err)
	}

	if resp.StatusCode < 200 || 299 < resp.StatusCode {
		return xerrors.Errorf("http status code: %v", resp.StatusCode)
	}

	return nil
}