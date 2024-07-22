package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"

	"github.com/bendbennett/go-api-demo/internal/user"
	"github.com/redis/go-redis/v9"
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
	telemetryEnabled bool,
) (*userCache, io.Closer, error) {
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     redisConf.Addr,
			Password: redisConf.Password,
		},
	)

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return nil, nil, err
	}

	if telemetryEnabled {
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
			return errors.Errorf("%s", err)
		}

		usrMap[fmt.Sprintf("%v:%v", usr, u.ID)] = mUsr
	}

	err := c.cache.MSet(ctx, usrMap).Err()
	if err != nil {
		return errors.Errorf("%s", err)
	}

	return nil
}

func (c *userCache) Read(ctx context.Context) ([]user.User, error) {
	iter := c.cache.Scan(
		ctx,
		0,
		fmt.Sprintf("%v:*", usr),
		0,
	).Iterator()

	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, errors.Errorf("%s", err)
	}

	if len(keys) == 0 {
		return nil, nil
	}

	mg := c.cache.MGet(
		ctx,
		keys...,
	)

	users := make([]user.User, len(mg.Val()))

	for k, v := range mg.Val() {
		u := user.User{}

		if _, ok := v.(string); !ok {
			return users, errors.New("could not assert user val as string")
		}

		if err := json.Unmarshal([]byte(v.(string)), &u); err != nil {
			return users, errors.Errorf("%s", err)
		}

		users[k] = u
	}

	return users, nil
}

type instrumentCache struct{}

func (ic instrumentCache) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := next(ctx, network, addr)

		return conn, err
	}
}

func (ic instrumentCache) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "redis: "+strings.ToUpper(cmd.Name()))
		defer span.End()

		err := next(ctx, cmd)

		return err
	}
}

func (ic instrumentCache) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		cmdNames := make([]string, len(cmds))

		for i, cmd := range cmds {
			cmdNames[i] = strings.ToUpper(cmd.Name())
		}

		ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "redis-pipeline: "+strings.Join(cmdNames, " --> "))
		defer span.End()

		err := next(ctx, cmds)

		return err
	}
}
