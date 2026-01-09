package service

import (
	"backend/internal/model"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService interface {
	CacheMessage(key string, value model.MessageView) error
	GetMessage(key string, limit int, offset int) ([]model.MessageView, error)
}

type redisService struct {
	client *redis.Client
}

func NewRedisService(client *redis.Client) RedisService {
	return &redisService{
		client: client,
	}
}

func (r *redisService) CacheMessage(key string, value model.MessageView) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)

	defer cancel()

	jsonValue, err := json.Marshal(value)

	zData := redis.Z{
		Score:  float64(value.CreatedAt.Unix()),
		Member: jsonValue,
	}

	err = r.client.ZAdd(ctx, key, zData).Err()
	if err != nil {
		return err
	}

	err = r.client.Expire(ctx, key, time.Hour*24).Err()

	if err != nil {
		return err
	}

	return nil
}

func (r *redisService) GetMessage(key string, limit int, offset int) ([]model.MessageView, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)

	defer cancel()
	val, err := r.client.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    "0",
		Max:    "+inf",
		Count:  int64(limit),
		Offset: int64(offset),
	}).Result()

	if err != nil {
		return []model.MessageView{}, err
	}

	messages := make([]model.MessageView, len(val))
	for i, v := range val {
		var msg model.MessageView
		err := json.Unmarshal([]byte(v), &msg)
		if err != nil {
			return []model.MessageView{}, err
		}
		messages[i] = msg
	}
	return messages, nil
}
