package consume

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

type creatorMock struct {
	hasBeenCalled bool
}

func (m *creatorMock) Create(context.Context, ...user.User) error {
	m.hasBeenCalled = true
	return nil
}

func (m *creatorMock) HasBeenCalled() bool {
	return m.hasBeenCalled
}

// TODO: Test should confirm that Create() is called on creatorMock
// when before == empty user
func TestProcessor_Process(t *testing.T) {
	cases := map[string]struct {
		creator              *creatorMock
		data                 any
		creatorHasBeenCalled bool
		expectedErr          error
	}{
		"no processing required": {
			&creatorMock{},
			map[string]interface{}{
				"after": map[string]interface{}{
					"mysql.go_api_demo.users.Value": map[string]interface{}{
						"id": "1",
					},
				},
				"before": map[string]interface{}{
					"mysql.go_api_demo.users.Value": map[string]interface{}{
						"id": "1",
					},
				},
			},
			false,
			nil,
		},
		"create called": {
			&creatorMock{},
			map[string]interface{}{
				"after": map[string]interface{}{
					"mysql.go_api_demo.users.Value": map[string]interface{}{
						"id": "1",
					},
				},
				"before": nil,
			},
			true,
			nil,
		},
		"unimplemented error": {
			&creatorMock{},
			map[string]interface{}{
				"after": nil,
				"before": map[string]interface{}{
					"mysql.go_api_demo.users.Value": map[string]interface{}{
						"id": "1",
					},
				},
			},
			false,
			errors.New("not implemented"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			processor := NewProcessor(
				c.creator,
			)

			err := processor.Process(
				context.Background(),
				c.data,
			)

			assert.Equal(t, c.creatorHasBeenCalled, c.creator.HasBeenCalled())
			assert.Equal(t, c.expectedErr, err)
		})
	}
}

var msg = map[string]interface{}{
	"after": map[string]interface{}{
		"go_api_demo_db.go_api_demo.users.Value": map[string]interface{}{
			"created_at": 1639512014000,
			"first_name": "john",
			"id":         "673b3c8c-3589-4b77-af89-94dcda52a861",
			"last_name":  "smith",
		},
	},
	"before": nil,
	"op":     "c",
	"source": map[string]interface{}{
		"connector": "mysql",
		"db":        "go-api-demo",
		"file":      "binlog.000002",
		"gtid":      nil,
		"name":      "go_api_demo_db",
		"pos":       9548,
		"query":     nil,
		"row":       0,
		"sequence":  nil,
		"server_id": 1,
		"snapshot": map[string]interface{}{
			"string": "false",
		},
		"table": map[string]interface{}{
			"string": "users",
		},
		"thread":  nil,
		"ts_ms":   1639512013000,
		"version": "1.7.1.Final",
	},
	"transaction": nil,
	"ts_ms": map[string]interface{}{
		"long": 1639512013850,
	},
}

// nolint:gocyclo
func Benchmark_Extract(b *testing.B) {
	extractUser := func(msg interface{}, key string) (usr, error) {
		valKey := "go_api_demo_db.go_api_demo.users.Value"

		if _, ok := msg.(map[string]interface{}); !ok {
			return usr{}, errors.New("cannot assert msg as map[string]interface{}")
		}

		m := msg.(map[string]interface{})

		if _, ok := m[key]; !ok {
			return usr{}, fmt.Errorf("%v key missing from msg", key)
		}

		if m[key] == nil {
			return usr{}, nil
		}

		if _, ok := m[key].(map[string]interface{}); !ok {
			return usr{}, fmt.Errorf("cannot assert msg[%v] as map[string]interface{}", key)
		}

		m = m[key].(map[string]interface{})

		if _, ok := m[valKey]; !ok {
			return usr{}, fmt.Errorf("%v key missing from msg", valKey)
		}

		if _, ok := m[valKey].(map[string]interface{}); !ok {
			return usr{}, fmt.Errorf("cannot assert msg[%v][%v] as map[string]interface{}", key, valKey)
		}

		m = m[valKey].(map[string]interface{})

		reqKeys := []string{"id", "first_name", "last_name", "created_at"}

		for _, reqKey := range reqKeys {
			if _, ok := m[reqKey]; !ok {
				return usr{}, fmt.Errorf("%v key missing from msg", reqKey)
			}
		}

		if _, ok := m["id"].(string); !ok {
			return usr{}, errors.New("cannot assert id is string")
		}

		if _, ok := m["first_name"].(string); !ok {
			return usr{}, errors.New("cannot assert first_name is string")
		}

		if _, ok := m["last_name"].(string); !ok {
			return usr{}, errors.New("cannot assert last_name is string")
		}

		if _, ok := m["created_at"].(int64); !ok {
			return usr{}, errors.New("cannot assert created_at is int64")
		}

		return usr{
			ID:        m["id"].(string),
			FirstName: m["first_name"].(string),
			LastName:  m["last_name"].(string),
			CreatedAt: m["created_at"].(int64),
		}, nil
	}

	for n := 0; n < b.N; n++ {
		_, _ = extractUser(msg, "after")

	}
}

func Benchmark_MapExtract(b *testing.B) {
	m := beforeAfter{}

	for n := 0; n < b.N; n++ {
		_ = mapstructure.Decode(msg, &m)

	}
}
