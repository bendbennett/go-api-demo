package test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	hello "github.com/bendbennett/go-api-demo/generated"
	"github.com/bendbennett/go-api-demo/internal/bootstrap"
	"github.com/bendbennett/go-api-demo/internal/config"
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

	// HTTP client for REST requests.
	restClient := newRESTClient()
	// gRPC client for gRPC requests.
	grpcClient := newGRPCClient()

	// HTTP - Test 200 OK
	statusCode, _, _, err := restClient.doRequest(httpRequest{
		method: http.MethodGet,
		url:    fmt.Sprintf("%v/", restClient.baseURL),
	})

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)

	// gRPC - Test OK
	res, err := grpcClient.client.Hello(
		context.Background(),
		&hello.HelloRequest{
			Name: "world",
		})

	require.NoError(t, err)
	assert.Equal(t, "Hello world", res.Message)

	cancel()
	err = eg.Wait()
	require.NoError(t, err)
}

func newRESTClient() *restClient {
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

	return &restClient{
		http.DefaultTransport,
		baseURL,
	}
}

type restClient struct {
	transport http.RoundTripper
	baseURL   string
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
		client: hello.NewHelloClient(conn),
	}
}

type grpcClient struct {
	client hello.HelloClient
}

type httpRequest struct {
	method  string
	url     string
	headers map[string]string
	body    []byte
}

func (a *restClient) doRequest(payload httpRequest) (int, []byte, http.Header, error) {
	req, err := http.NewRequest(
		payload.method,
		payload.url,
		bytes.NewReader(payload.body),
	)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("new request: %w", err)
	}

	for name, value := range payload.headers {
		req.Header.Set(name, value)
	}

	resp, err := a.transport.RoundTrip(req)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("round trip: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return resp.StatusCode, nil, nil, fmt.Errorf(
			"%d response from gateway %s: %s",
			resp.StatusCode,
			req.URL.String(),
			body,
		)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("body read err: %v", err)
	}

	return resp.StatusCode, bodyBytes, resp.Header, err
}
