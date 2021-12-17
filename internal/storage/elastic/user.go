package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

const usrs = "users"

type search interface {
	Perform(request *http.Request) (*http.Response, error)
}

type userSearch struct {
	search search
}

type r struct {
	err        error
	statusCode int
	isError    bool
}

func NewUserSearch(
	esConf elasticsearch.Config,
	isTracingEnabled bool,
) (*userSearch, error) {
	es, err := elasticsearch.NewClient(esConf)
	if err != nil {
		return nil, err
	}

	_, err = es.Ping()
	if err != nil {
		return nil, err
	}

	us := userSearch{es}

	if isTracingEnabled {
		us = userSearch{&instrumentSearch{es}}
	}

	return &us, nil
}

type elasticUser struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

func (s *userSearch) Create(
	ctx context.Context,
	users ...user.User,
) error {
	var wg sync.WaitGroup
	responses := make(chan r)

	for _, u := range users {
		eU := elasticUser{
			ID:        u.ID,
			FullName:  fmt.Sprintf("%s %s", u.FirstName, u.LastName),
			FirstName: u.FirstName,
			LastName:  u.LastName,
			CreatedAt: u.CreatedAt,
		}

		j, err := json.Marshal(eU)
		if err != nil {
			return errors.Errorf("%s", err)
		}

		wg.Add(1)

		go func(uID string, j []byte) {
			defer wg.Done()

			req := esapi.IndexRequest{
				Index:      usrs,
				DocumentID: uID,
				Body:       strings.NewReader(string(j)),
				Refresh:    "false",
			}

			resp, err := req.Do(ctx, s.search)
			if err != nil {
				responses <- r{err: err}
				return
			}

			defer resp.Body.Close()

			responses <- r{
				statusCode: resp.StatusCode,
				isError:    resp.IsError(),
				err:        err,
			}
		}(u.ID, j)
	}

	go func() {
		wg.Wait()
		close(responses)
	}()

	var err error

	for resp := range responses {
		if resp.err != nil {
			switch err {
			case nil:
				err = resp.err
			default:
				err = fmt.Errorf(
					"%w, %s",
					resp.err,
					errors.Errorf("%s", err),
				)
			}
		}
	}

	return err
}

func (s *userSearch) Search(
	ctx context.Context,
	searchTerm string,
) ([]user.User, error) {
	req := esapi.SearchRequest{
		Index:          []string{usrs},
		DocvalueFields: []string{"first_name.keyword", "last_name.keyword"},
		Query:          fmt.Sprintf("*%s*", searchTerm),
	}

	resp, err := req.Do(ctx, s.search)
	if err != nil {
		return nil, errors.Errorf("%s", err)
	}

	if resp.IsError() {
		var e = e{}

		if err := json.NewDecoder(resp.Body).Decode(&e); err != nil {
			return nil, errors.Errorf("%s", err)
		}

		err := fmt.Errorf(
			"status: %d",
			e.Status,
		)

		if len(e.Err.RootCause) > 0 {
			err = fmt.Errorf(
				"%v, type: %v, reason: %v",
				err.Error(),
				e.Err.RootCause[0].Type,
				e.Err.RootCause[0].Reason,
			)
		}

		return nil, err
	}

	h := h{}

	if err = json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, errors.Errorf("%s", err)
	}

	var users []user.User

	for _, v := range h.Hits.HitsHits {
		u := user.User{
			CreatedAt: v.Source.CreatedAt,
			ID:        v.Source.ID,
			FirstName: v.Source.FirstName,
			LastName:  v.Source.LastName,
		}

		users = append(users, u)
	}

	return users, nil
}

type e struct {
	Err    err `json:"error"`
	Status int `json:"status"`
}

type err struct {
	RootCause []rootCause `json:"root_cause"`
}

type rootCause struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type h struct {
	Hits hh `json:"hits"`
}

type hh struct {
	HitsHits []s `json:"hits"`
}

type s struct {
	Source u `json:"_source"`
}

type u struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

type instrumentSearch struct {
	search search
}

func (s *instrumentSearch) Perform(request *http.Request) (*http.Response, error) {
	span, _ := opentracing.StartSpanFromContext(
		request.Context(),
		fmt.Sprintf(
			"HTTP %s: %s",
			request.Method,
			request.URL.Path,
		),
	)
	defer span.Finish()

	ext.Component.Set(span, "database/elasticsearch")
	ext.HTTPMethod.Set(span, request.Method)
	ext.HTTPUrl.Set(span, request.URL.Path)

	return s.search.Perform(request)
}
