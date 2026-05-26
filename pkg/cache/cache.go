package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/husky/husky/internal/config"
)

// NewRedis 创建 Redis 连接
func NewRedis(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect redis: %w", err)
	}

	return client, nil
}

// Cache Redis 缓存封装
type Cache struct {
	client *redis.Client
}

// NewCache 创建缓存实例
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// Get 获取缓存数据
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Set 设置缓存数据
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, string(data), expiration).Err()
}

// Delete 删除缓存
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func (c *Cache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetJSON 获取 JSON 格式缓存数据
func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// SetJSON 设置 JSON 格式缓存数据
func (c *Cache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, string(data), expiration).Err()
}

// Increment 自增计数器
func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Decrement 自减计数器
func (c *Cache) Decrement(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// Expire 设置过期时间
func (c *Cache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余生存时间
func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Keys 获取匹配的键
func (c *Cache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// FlushDB 清空当前数据库
func (c *Cache) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// CacheKey 生成缓存键的辅助函数
type CacheKey struct{}

// User 用户相关缓存键
func (CacheKey) User(id uint) string {
	return fmt.Sprintf("user:%d", id)
}

// UserByEmail 按邮箱查询用户的缓存键
func (CacheKey) UserByEmail(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

// Ticket 工单相关缓存键
func (CacheKey) Ticket(id uint) string {
	return fmt.Sprintf("ticket:%d", id)
}

// TicketByNo 按工单号查询的缓存键
func (CacheKey) TicketByNo(ticketNo string) string {
	return fmt.Sprintf("ticket:no:%s", ticketNo)
}

// Knowledge 知识文章相关缓存键
func (CacheKey) Knowledge(id uint) string {
	return fmt.Sprintf("knowledge:%d", id)
}

// Agent Agent 相关缓存键
func (CacheKey) Agent(id uint) string {
	return fmt.Sprintf("agent:%d", id)
}

// Counter 计数器相关缓存键
func (CacheKey) Counter(name string) string {
	return fmt.Sprintf("counter:%s", name)
}

// Lock 分布式锁相关缓存键
func (CacheKey) Lock(name string) string {
	return fmt.Sprintf("lock:%s", name)
}

var CK = CacheKey{}
