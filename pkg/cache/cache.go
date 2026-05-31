package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/husky/husky/internal/config"
)

// CacheInterface 缓存接口，支持 Redis 和 Noop 两种实现
type CacheInterface interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

// NoopCache 空缓存实现，Redis 不可用时降级使用
type NoopCache struct{}

func (c *NoopCache) Get(ctx context.Context, key string) (string, error) {
	return "", redis.Nil
}

func (c *NoopCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (c *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoopCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (c *NoopCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	return redis.Nil
}

func (c *NoopCache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}

func (c *NoopCache) Increment(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (c *NoopCache) Decrement(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (c *NoopCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (c *NoopCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

// NewRedis 创建 Redis 连接，非必需，host 为空时直接返回 nil
func NewRedis(cfg *config.Config) (*redis.Client, error) {
	if !cfg.Cache.Enable {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Cache.Host, cfg.Cache.Port),
		Password:     cfg.Cache.Password,
		DB:           cfg.Cache.DB,
		PoolSize:     cfg.Cache.PoolSize,
		MinIdleConns: cfg.Cache.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		client.Close()
		return nil, nil
	}

	return client, nil
}

// Cache Redis 缓存封装，实现 CacheInterface
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
