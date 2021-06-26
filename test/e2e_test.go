package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	user "github.com/bendbennett/go-api-demo/generated"
	"github.com/bendbennett/go-api-demo/internal/bootstrap"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func Test_E2E(t *testing.T) {
	if err := os.Setenv("HTTP_PORT", "3001"); err != nil {
		t.Error(err)
	}

	if err := os.Setenv("GRPC_PORT", "1235"); err != nil {
		t.Error(err)
	}

	a := bootstrap.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return a.Run(egCtx)
	})

	httpClient := newHTTPClient()
	grpcClient := newGRPCClient()

	// User - Create
	_ = userCreateHTTP(t, httpClient)
	_ = userCreateGRPC(t, grpcClient)

	cancel()
	err := eg.Wait()
	require.NoError(t, err)
}

type httpClient struct {
	transport http.RoundTripper
	baseURL   string
}

func newHTTPClient() *httpClient {
	baseURL := fmt.Sprintf(
		"http://%s:%d",
		config.GetEnvAsString(
			"TEST_HOST",
			"localhost",
		),
		config.GetEnvAsInt(
			"TEST_HTTP_PORT",
			3001,
		),
	)

	return &httpClient{
		http.DefaultTransport,
		baseURL,
	}
}

type httpRequest struct {
	method  string
	url     string
	headers map[string]string
	body    []byte
}

func (a *httpClient) doRequest(payload httpRequest) (int, []byte, error) {
	req, err := http.NewRequest(
		payload.method,
		payload.url,
		bytes.NewReader(payload.body),
	)
	if err != nil {
		return 0, nil, fmt.Errorf("new request: %w", err)
	}

	for name, value := range payload.headers {
		req.Header.Set(name, value)
	}

	resp, err := a.transport.RoundTrip(req)
	if err != nil {
		return 0, nil, fmt.Errorf("round trip: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, fmt.Errorf("body read err: %v", err)
	}

	return resp.StatusCode, bodyBytes, err
}

type userHTTP struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	CreatedAt string `json:"created_at"`
}

type errResponseHTTP struct {
	Errors  map[string]string `json:"errors"`
	Message string            `json:"message"`
}

func userCreateHTTP(t *testing.T, httpClient *httpClient) userHTTP {
	// Malformed JSON
	statusCode, body, err := httpClient.doRequest(httpRequest{
		method: http.MethodPost,
		url:    fmt.Sprintf("%v/user", httpClient.baseURL),
		body: []byte(
			`{
				"first_name": "john
			}`),
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, statusCode)

	jsonErr := errResponseHTTP{}
	err = json.Unmarshal(body, &jsonErr)
	require.NoError(t, err)

	assert.Len(t, jsonErr.Errors, 1)
	assert.NotEmpty(t, jsonErr.Errors["body"])
	assert.NotEmpty(t, jsonErr.Message)

	// Invalid input - first name too short
	statusCode, body, err = httpClient.doRequest(httpRequest{
		method: http.MethodPost,
		url:    fmt.Sprintf("%v/user", httpClient.baseURL),
		body: []byte(
			`{
				"first_name": "j",
				"last_name": "smith"
			}`),
	})

	validationErr := errResponseHTTP{}
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, statusCode)

	err = json.Unmarshal(body, &validationErr)
	require.NoError(t, err)

	assert.Len(t, validationErr.Errors, 1)
	assert.NotEmpty(t, validationErr.Errors["first_name"])
	assert.NotEmpty(t, validationErr.Message)

	// Success
	statusCode, body, err = httpClient.doRequest(httpRequest{
		method: http.MethodPost,
		url:    fmt.Sprintf("%v/user", httpClient.baseURL),
		body: []byte(
			`{
				"first_name": "john",
				"last_name": "smith"
			}`),
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, statusCode)

	userHTTP := userHTTP{}
	err = json.Unmarshal(body, &userHTTP)
	require.NoError(t, err)

	assert.True(t, validate.IsUUID(userHTTP.ID))
	assert.Equal(t, "john", userHTTP.FirstName)
	assert.Equal(t, "smith", userHTTP.LastName)
	_, err = time.Parse(time.RFC3339, userHTTP.CreatedAt)
	assert.NoError(t, err)

	return userHTTP
}

type grpcClient struct {
	userClient user.UserClient
}

func newGRPCClient() *grpcClient {
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf(
			"%s:%d",
			config.GetEnvAsString(
				"TEST_HOST",
				"localhost",
			),
			config.GetEnvAsInt(
				"TEST_GRPC_PORT",
				1235,
			),
		),
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}

	return &grpcClient{
		userClient: user.NewUserClient(conn),
	}
}

func userCreateGRPC(t *testing.T, grpcClient *grpcClient) *user.CreateResponse {
	// Missing required field
	_, err := grpcClient.userClient.Create(
		context.Background(),
		&user.CreateRequest{
			FirstName: "john",
		})

	require.Error(t, err)

	// Invalid input - first name too short
	_, err = grpcClient.userClient.Create(
		context.Background(),
		&user.CreateRequest{
			FirstName: "j",
			LastName:  "smith",
		})

	require.Error(t, err)

	// Success
	userGRPC, err := grpcClient.userClient.Create(
		context.Background(),
		&user.CreateRequest{
			FirstName: "john",
			LastName:  "smith",
		})

	require.NoError(t, err)
	assert.True(t, validate.IsUUID(userGRPC.Id))
	assert.Equal(t, "john", userGRPC.FirstName)
	assert.Equal(t, "smith", userGRPC.LastName)
	_, err = time.Parse(time.RFC3339, userGRPC.CreatedAt)
	assert.NoError(t, err)

	return userGRPC
}
