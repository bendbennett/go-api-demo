package schema

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/linkedin/goavro/v2"
	"github.com/pkg/errors"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type schemaClient struct {
	httpClient httpClient
	codecs     map[string]*goavro.Codec
	mu         sync.Mutex
}

func NewClient(clientTimeout time.Duration) *schemaClient {
	c := http.Client{
		Timeout: clientTimeout,
	}

	return &schemaClient{
		httpClient: &c,
		codecs:     make(map[string]*goavro.Codec),
	}
}

type decoder interface {
	Decode([]byte) (interface{}, error)
}

type d struct {
	codec *goavro.Codec
}

var _ decoder = (*d)(nil)

// Decode accepts a slice of bytes, discarding the first 5 bytes because of their usage
// in the Confluent schema registry.
// See https://stackoverflow.com/questions/40548909/consume-kafka-avro-messages-in-go
func (d *d) Decode(msg []byte) (interface{}, error) {
	nMsg, _, err := d.codec.NativeFromBinary(msg[5:])
	if err != nil {
		return nil, errors.Wrap(err, "could not decode msg")
	}

	return nMsg, nil
}

func (c *schemaClient) GetDecoder(endpoint string) (*d, error) {
	codec, err := c.getSchemaCodec(endpoint)
	if err != nil {
		return nil, err
	}

	return &d{codec}, nil
}

func (c *schemaClient) getSchemaCodec(endpoint string) (*goavro.Codec, error) {
	if codec, ok := c.codecs[endpoint]; ok {
		return codec, nil
	}

	req, err := http.NewRequest(
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("new request error: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("schema registry request error: %w", err)
	}
	defer resp.Body.Close()

	type schema struct {
		Schema interface{} `json:"schema"`
	}

	s := schema{}

	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		return nil, fmt.Errorf("decoding response from schema registry error: %w", err)
	}

	codec, err := goavro.NewCodec(fmt.Sprintf("%v", s.Schema))
	if err != nil {
		return nil, fmt.Errorf("new codec error: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.codecs[endpoint] = codec

	return codec, nil
}
