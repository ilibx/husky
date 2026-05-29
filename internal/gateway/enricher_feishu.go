package gateway

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type feishuUserAPI interface {
	GetUserInfo(ctx context.Context, userID string) (map[string]interface{}, error)
}

type FeishuEnricher struct {
	cli      feishuUserAPI
	mu       sync.RWMutex
	cache    map[string]*UserContext
	cacheTTL time.Duration
	cacheAt  map[string]time.Time
}

func NewFeishuEnricher(cli feishuUserAPI) *FeishuEnricher {
	return &FeishuEnricher{
		cli:      cli,
		cache:    make(map[string]*UserContext),
		cacheTTL: 5 * time.Minute,
		cacheAt:  make(map[string]time.Time),
	}
}

func (e *FeishuEnricher) Enrich(ctx *UserContext) error {
	if ctx.ChannelUserID == "" {
		return fmt.Errorf("feishu enricher: empty user id")
	}

	if cached := e.getFromCache(ctx.ChannelUserID); cached != nil {
		ctx.OpenID = cached.OpenID
		ctx.UnionID = cached.UnionID
		ctx.Username = cached.Username
		ctx.DisplayName = cached.DisplayName
		ctx.Department = cached.Department
		ctx.Title = cached.Title
		ctx.AvatarURL = cached.AvatarURL
		return nil
	}

	info, err := e.cli.GetUserInfo(context.Background(), ctx.ChannelUserID)
	if err != nil {
		return fmt.Errorf("feishu get user info: %w", err)
	}

	if user, ok := info["data"].(map[string]interface{}); ok {
		if u, ok := user["user"].(map[string]interface{}); ok {
			ctx.OpenID = toString(u["open_id"])
			ctx.UnionID = toString(u["union_id"])
			ctx.Username = toString(u["name"])
			ctx.DisplayName = toString(u["name"])
			ctx.Department = flattenDepartment(u["department_ids"])
			ctx.Title = toString(u["title"])
			ctx.AvatarURL = getAvatarURL(u["avatar"])
			ctx.Email = toString(u["email"])
			ctx.Phone = toString(u["mobile"])

			e.storeInCache(ctx.ChannelUserID, ctx)
		}
	}

	return nil
}

func (e *FeishuEnricher) getFromCache(userID string) *UserContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cached, ok := e.cache[userID]
	if !ok {
		return nil
	}
	if time.Since(e.cacheAt[userID]) >= e.cacheTTL {
		return nil
	}
	return cached
}

func (e *FeishuEnricher) storeInCache(userID string, ctx *UserContext) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.cache[userID] = ctx
	e.cacheAt[userID] = time.Now()
}

func (e *FeishuEnricher) InvalidateCache(userID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.cache, userID)
	delete(e.cacheAt, userID)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func flattenDepartment(v interface{}) string {
	if v == nil {
		return ""
	}
	if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
		return fmt.Sprintf("%v", arr[0])
	}
	return fmt.Sprintf("%v", v)
}

func getAvatarURL(v interface{}) string {
	if v == nil {
		return ""
	}
	if avatar, ok := v.(map[string]interface{}); ok {
		if url, ok := avatar["avatar_origin"].(string); ok && url != "" {
			return url
		}
		if url, ok := avatar["avatar_72"].(string); ok && url != "" {
			return url
		}
		if url, ok := avatar["avatar_240"].(string); ok && url != "" {
			return url
		}
	}
	return ""
}
