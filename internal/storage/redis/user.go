package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/go-redis/redis/v8"
)

const usr = "user"

type cache interface {
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	MSet(ctx context.Context, values ...interface{}) *redis.StatusCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
}

type userCache struct {
	cache cache
}

func NewUserCache(
	redisConf redis.Options,
	isTracingEnabled bool,
) (*userCache, io.Closer, error) {
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     redisConf.Addr,
			Password: redisConf.Password,
		},
	)

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}

	if isTracingEnabled {
		rdb.AddHook(instrumentCache{})
	}

	return &userCache{
		cache: rdb,
	}, rdb, nil
}

func (c *userCache) Create(ctx context.Context, users ...user.User) error {
	if len(users) == 0 {
		return nil
	}

	usrMap := make(map[string]interface{}, len(users))

	for _, u := range users {
		mUsr, err := json.Marshal(u)
		if err != nil {
			return err
		}

		usrMap[fmt.Sprintf("%v:%v", usr, u.ID)] = mUsr
	}

	return c.cache.MSet(ctx, usrMap).Err()
}

func (c *userCache) Read(ctx context.Context) ([]user.User, error) {
	var (
		users []user.User
		keys  []string
	)

	iter := c.cache.Scan(
		ctx,
		0,
		fmt.Sprintf("%v:*", usr),
		0,
	).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return users, err
	}

	if len(keys) == 0 {
		return users, nil
	}

	mg := c.cache.MGet(
		ctx,
		keys...,
	)

	for _, v := range mg.Val() {
		u := user.User{}

		if _, ok := v.(string); !ok {
			return users, errors.New("could not assert user val as string")
		}

		if err := json.Unmarshal([]byte(v.(string)), &u); err != nil {
			return users, err
		}

		users = append(users, u)
	}

	return users, nil
}

type instrumentCache struct{}

func (ic instrumentCache) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span, sCtx := opentracing.StartSpanFromContext(ctx, "redis:cmd")
	ext.DBType.Set(span, "redis")
	ext.DBStatement.Set(span, strings.ToUpper(cmd.Name()))

	return sCtx, nil
}

func (ic instrumentCache) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.Finish()
	}

	return nil
}

func (ic instrumentCache) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	span, sCtx := opentracing.StartSpanFromContext(ctx, "redis:pipeline-cmd")
	ext.DBType.Set(span, "redis")
	cmdNames := make([]string, len(cmds))
	for i, cmd := range cmds {
		cmdNames[i] = strings.ToUpper(cmd.Name())
	}
	ext.DBStatement.Set(span, strings.Join(cmdNames, " --> "))

	return sCtx, nil
}

func (ic instrumentCache) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.Finish()
	}

	return nil
}
