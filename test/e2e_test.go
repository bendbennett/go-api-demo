package test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/joho/godotenv"

	"github.com/go-redis/redis/v8"

	user "github.com/bendbennett/go-api-demo/generated"
	"github.com/bendbennett/go-api-demo/internal/bootstrap"
	"github.com/bendbennett/go-api-demo/internal/config"
	"github.com/bendbennett/go-api-demo/internal/validate"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate"
	migratemysql "github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func Test_E2E(t *testing.T) {
	env(t)
	purge(t)

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
	userCreateHTTP(t, httpClient)
	userCreateGRPC(t, grpcClient)

	// User - Read
	userReadHTTP(t, httpClient)
	userReadGRPC(t, grpcClient)

	// User - Search
	userSearchHTTP(t, httpClient)
	userSearchGRPC(t, grpcClient)

	cancel()
}

func env(t *testing.T) {
	env := map[string]string{
		"HTTP_PORT":          "3001",
		"GRPC_PORT":          "1235",
		"TRACING_ENABLED":    "false",
		"LOGGING_PRODUCTION": "true",
		"METRICS_ENABLED":    "false",
	}

	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			t.Error(err)
		}
	}

	_ = godotenv.Load("../.env")
}

// nolint: gocyclo
func purge(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		db, err := sql.Open(
			"mysql",
			fmt.Sprintf(
				"%s:%s@tcp(%s:%s)/%s?parseTime=true",
				config.GetEnvAsString("MYSQL_USER", ""),
				config.GetEnvAsString("MYSQL_PASSWORD", ""),
				config.GetEnvAsString("MYSQL_HOST", ""),
				config.GetEnvAsString("MYSQL_PORT", ""),
				config.GetEnvAsString("MYSQL_DBNAME", ""),
			),
		)
		if err != nil {
			t.Error(err)
		}

		dbInstance, err := migratemysql.WithInstance(db, &migratemysql.Config{})
		if err != nil {
			panic(err)
		}

		m, err := migrate.NewWithDatabaseInstance(
			"file://../internal/storage/mysql/migrations",
			"mysql",
			dbInstance,
		)
		if err != nil {
			t.Error(err)
		}

		if err := m.Drop(); err != nil && err != migrate.ErrNoChange {
			t.Error(err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			t.Error(err)
		}
	}()

	go func() {
		defer wg.Done()

		rdb := redis.NewClient(
			&redis.Options{
				Addr: fmt.Sprintf("%s:%v",
					config.GetEnvAsString("REDIS_HOST", "localhost"),
					config.GetEnvAsInt("REDIS_PORT", 6379),
				),
				Password: config.GetEnvAsString("REDIS_PASSWORD", "pass"),
			},
		)

		rdb.FlushAll(context.Background())

		for {
			cmd := rdb.Keys(context.Background(), "*")
			keys, err := cmd.Result()
			if err != nil {
				t.Error(err)
			}

			if len(keys) == 0 {
				break
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()

		es, err := elasticsearch.NewClient(
			elasticsearch.Config{
				Addresses: config.GetEnvAsSliceOfStrings(
					"ELASTICSEARCH_ADDRESSES",
					",",
					[]string{"http://localhost:9200"}),
			})
		if err != nil {
			t.Error(err)
		}

		for {
			deleteReq := esapi.DeleteByQueryRequest{
				Index: []string{"users"},
				Body:  strings.NewReader(`{"query": {"match_all": {}}}`),
			}

			_, err = deleteReq.Do(context.Background(), es)
			if err != nil {
				t.Error(err)
			}

			searchReq := esapi.SearchRequest{
				Index: []string{"users"},
				Body:  strings.NewReader(`{"query": {"match_all": {}}}`),
			}

			resp, err := searchReq.Do(context.Background(), es)
			if err != nil {
				t.Error(err)
			}

			if resp.StatusCode == http.StatusNotFound {
				break
			}

			var r map[string]interface{}

			if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
				t.Error(err)
			}

			if _, ok := r["hits"]; !ok {
				break
			}

			if int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)) == 0 {
				break
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	wg.Wait()
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

type usersHTTP []userHTTP

type errResponseHTTP struct {
	Errors  map[string]string `json:"errors"`
	Message string            `json:"message"`
}

func userCreateHTTP(t *testing.T, httpClient *httpClient) {
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
}

func userReadHTTP(t *testing.T, httpClient *httpClient) {
	maxAttempts := 500
	usersHTTP := usersHTTP{}

	// Success
	// We require a for loop here as the length of time it takes
	// for the DB mutation events to be published into Kafka and
	// processed is non-deterministic.
	for i := 0; i < maxAttempts; i++ {
		statusCode, body, err := httpClient.doRequest(httpRequest{
			method: http.MethodGet,
			url:    fmt.Sprintf("%v/user", httpClient.baseURL),
		})

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode)

		err = json.Unmarshal(body, &usersHTTP)
		assert.NoError(t, err)

		if len(usersHTTP) == 2 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	assert.Len(t, usersHTTP, 2)
	assert.True(t, validate.IsUUID(usersHTTP[0].ID))
	assert.NotEmpty(t, usersHTTP[0].FirstName)
	assert.NotEmpty(t, usersHTTP[0].LastName)
	_, err := time.Parse(time.RFC3339, usersHTTP[0].CreatedAt)
	assert.NoError(t, err)
	assert.True(t, validate.IsUUID(usersHTTP[1].ID))
	assert.NotEmpty(t, usersHTTP[1].FirstName)
	assert.NotEmpty(t, usersHTTP[1].LastName)
	_, err = time.Parse(time.RFC3339, usersHTTP[1].CreatedAt)
	assert.NoError(t, err)
}

func userSearchHTTP(t *testing.T, httpClient *httpClient) {
	maxAttempts := 500
	usersHTTP := usersHTTP{}

	// Success
	// We require a for loop here as the length of time it takes
	// for the DB mutation events to be published into Kafka and
	// processed is non-deterministic.
	// We also require a check for a 404 as the index is only
	// created when the first user doc is created in elasticsearch.
	for i := 0; i < maxAttempts; i++ {
		statusCode, body, err := httpClient.doRequest(httpRequest{
			method: http.MethodGet,
			url:    fmt.Sprintf("%v/user/search/smith", httpClient.baseURL),
		})

		require.NoError(t, err)

		retryCodes := map[int]struct{}{http.StatusNotFound: {}, http.StatusInternalServerError: {}}

		if _, ok := retryCodes[statusCode]; ok {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		assert.Equal(t, http.StatusOK, statusCode)

		err = json.Unmarshal(body, &usersHTTP)
		assert.NoError(t, err)

		if len(usersHTTP) == 2 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	assert.Len(t, usersHTTP, 2)
	assert.True(t, validate.IsUUID(usersHTTP[0].ID))
	assert.NotEmpty(t, usersHTTP[0].FirstName)
	assert.NotEmpty(t, usersHTTP[0].LastName)
	_, err := time.Parse(time.RFC3339, usersHTTP[0].CreatedAt)
	assert.NoError(t, err)
	assert.True(t, validate.IsUUID(usersHTTP[1].ID))
	assert.NotEmpty(t, usersHTTP[1].FirstName)
	assert.NotEmpty(t, usersHTTP[1].LastName)
	_, err = time.Parse(time.RFC3339, usersHTTP[1].CreatedAt)
	assert.NoError(t, err)
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

func userCreateGRPC(t *testing.T, grpcClient *grpcClient) {
	// Missing required field
	_, err := grpcClient.userClient.Create(
		context.Background(),
		&user.CreateRequest{
			FirstName: "joanna",
		})

	require.Error(t, err)

	// Invalid input - first name too short
	_, err = grpcClient.userClient.Create(
		context.Background(),
		&user.CreateRequest{
			FirstName: "j",
			LastName:  "smithson",
		})

	require.Error(t, err)

	// Success
	userGRPC, err := grpcClient.userClient.Create(
		context.Background(),
		&user.CreateRequest{
			FirstName: "joanna",
			LastName:  "smithson",
		})

	require.NoError(t, err)
	assert.True(t, validate.IsUUID(userGRPC.Id))
	assert.Equal(t, "joanna", userGRPC.FirstName)
	assert.Equal(t, "smithson", userGRPC.LastName)
	_, err = time.Parse(time.RFC3339, userGRPC.CreatedAt)
	assert.NoError(t, err)
}

func userReadGRPC(t *testing.T, grpcClient *grpcClient) {
	maxAttempts := 500
	usersGRPC := &user.UsersResponse{}

	// Success
	// We require a for loop here as the length of time it takes
	// for the DB mutation events to be published into Kafka and
	// processed is non-deterministic.
	for i := 0; i < maxAttempts; i++ {
		var err error

		usersGRPC, err = grpcClient.userClient.Read(
			context.Background(),
			&user.ReadRequest{},
		)

		assert.NoError(t, err)

		if len(usersGRPC.Users) == 2 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	assert.Len(t, usersGRPC.Users, 2)

	assert.True(t, validate.IsUUID(usersGRPC.Users[0].Id))
	assert.NotEmpty(t, usersGRPC.Users[0].FirstName)
	assert.NotEmpty(t, usersGRPC.Users[0].LastName)
	createdAt, err := time.Parse(time.RFC3339, usersGRPC.Users[0].CreatedAt)
	assert.NoError(t, err)
	assert.True(t, !createdAt.IsZero())

	assert.True(t, validate.IsUUID(usersGRPC.Users[1].Id))
	assert.NotEmpty(t, usersGRPC.Users[1].FirstName)
	assert.NotEmpty(t, usersGRPC.Users[1].LastName)
	createdAt, err = time.Parse(time.RFC3339, usersGRPC.Users[1].CreatedAt)
	assert.NoError(t, err)
	assert.True(t, !createdAt.IsZero())
}

func userSearchGRPC(t *testing.T, grpcClient *grpcClient) {
	maxAttempts := 500
	usersGRPC := &user.UsersResponse{}

	// Success
	// We require a for loop here as the length of time it takes
	// for the DB mutation events to be published into Kafka and
	// processed is non-deterministic.
	for i := 0; i < maxAttempts; i++ {
		var err error

		usersGRPC, err = grpcClient.userClient.Search(
			context.Background(),
			&user.SearchRequest{SearchTerm: "smith"},
		)

		assert.NoError(t, err)

		if len(usersGRPC.Users) == 2 {
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	assert.Len(t, usersGRPC.Users, 2)

	assert.True(t, validate.IsUUID(usersGRPC.Users[0].Id))
	assert.NotEmpty(t, usersGRPC.Users[0].FirstName)
	assert.NotEmpty(t, usersGRPC.Users[0].LastName)
	createdAt, err := time.Parse(time.RFC3339, usersGRPC.Users[0].CreatedAt)
	assert.NoError(t, err)
	assert.True(t, !createdAt.IsZero())

	assert.True(t, validate.IsUUID(usersGRPC.Users[1].Id))
	assert.NotEmpty(t, usersGRPC.Users[1].FirstName)
	assert.NotEmpty(t, usersGRPC.Users[1].LastName)
	createdAt, err = time.Parse(time.RFC3339, usersGRPC.Users[1].CreatedAt)
	assert.NoError(t, err)
	assert.True(t, !createdAt.IsZero())
}
