package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/swallowarc/porker-rpc/internal/commons/errs"
	"github.com/swallowarc/porker-rpc/internal/commons/loggers"
	"github.com/swallowarc/porker-rpc/internal/interface_adapters/gateways"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
)

const (
	maxRetries = 5
)

type (
	redisClient struct {
		cli redis.Cmdable
	}
)

func NewRedisClient(config Config) gateways.MemDBClient {
	return &redisClient{
		cli: redis.NewClient(&redis.Options{
			Addr:       config.HostPort,
			Password:   config.Password, // no password set
			DB:         config.DB,       // use default DB
			MaxRetries: maxRetries,
		}),
	}
}

func (c *redisClient) Ping(ctx context.Context) error {
	if err := c.cli.Ping(ctx).Err(); err != nil {
		return xerrors.Errorf("failed to redis Ping: %w", err)
	}
	return nil
}

func (c *redisClient) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	if err := c.cli.Set(ctx, key, value, duration).Err(); err != nil {
		return xerrors.Errorf("failed to redis Set: %w", err)
	}
	return nil
}

func (c *redisClient) SetNX(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	if err := c.cli.SetNX(ctx, key, value, duration).Err(); err != nil {
		return xerrors.Errorf("failed to redis SetNX: %w", err)
	}
	return nil
}

func (c *redisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.cli.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errs.NewNotFoundError(fmt.Sprintf("%s does not exist", key))
	}
	if err != nil {
		return "", xerrors.Errorf("failed to redis Get: %w", err)
	}
	return val, nil
}

func (c *redisClient) Del(ctx context.Context, key string) error {
	err := c.cli.Del(ctx, key).Err()
	if err == redis.Nil {
		return errs.NewNotFoundError(fmt.Sprintf("%s does not exist", key))
	}
	if err != nil {
		return xerrors.Errorf("failed to redis Del: %w", err)
	}
	return nil
}

func (c *redisClient) SAdd(ctx context.Context, key string, values ...interface{}) error {
	if err := c.cli.SAdd(ctx, key, values...).Err(); err != nil {
		return xerrors.Errorf("failed to redis SAdd: %w", err)
	}
	return nil
}

func (c *redisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	err := c.cli.SRem(ctx, key, members...).Err()
	if err == redis.Nil {
		return errs.NewNotFoundError(fmt.Sprintf("not exist. key: %s, members: %v, ", key, members))
	}
	if err != nil {
		return xerrors.Errorf("failed to redis SRem: %w", err)
	}
	return nil
}

func (c *redisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	members, err := c.cli.SMembers(ctx, key).Result()
	if err != nil {
		return nil, xerrors.Errorf("failed to redis SMembers: %w", err)
	}
	if members == nil {
		return []string{}, nil
	}
	return members, nil
}

func (c *redisClient) PublishStream(ctx context.Context, streamKey string, messages map[string]interface{}) error {
	values := make([]interface{}, 0, len(messages)*2)
	for k, v := range messages {
		values = append(values, k, v)
	}

	if err := c.cli.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: 1,
		ID:     "*",
		Values: values,
	}).Err(); err != nil {
		return xerrors.Errorf("failed to redis XAdd: %w", err)
	}
	return nil
}

func (c *redisClient) ReadStream(ctx context.Context, streamKey, messageKey, previousID string) (id, message string, err error) {
	const subscribeDuration = 3 * time.Second

	cmd := c.cli.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamKey, previousID},
		Block:   subscribeDuration,
	})
	streams, err := cmd.Result()
	if err == redis.Nil {
		return "", "", errs.NewNotFoundError("response nil from stream")
	}
	if err != nil {
		return "", "", xerrors.Errorf("failed to redis XRead. err: %w, streamKey: %s, messageID: %s", err, streamKey, previousID)
	}

	stream := streams[0]
	msg := stream.Messages[len(stream.Messages)-1]
	v, ok := msg.Values[messageKey].(string)
	if !ok {
		loggers.Logger(ctx).Warn("cast to string from stream message failed", zap.Reflect("message", msg))
		return "", "", nil
	}

	return msg.ID, v, nil
}

func (c *redisClient) ReadStreamLatest(ctx context.Context, streamKey, messageKey string) (id, message string, err error) {
	return c.ReadStream(ctx, streamKey, messageKey, "0")
}

func (c *redisClient) Expire(ctx context.Context, key string, duration time.Duration) error {
	if err := c.cli.Expire(ctx, key, duration).Err(); err != nil {
		return xerrors.Errorf("failed to redis Expire: %w", err)
	}

	return nil
}
