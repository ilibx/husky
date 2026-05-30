package feishu

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) GetUserInfo(ctx context.Context, userID string) (map[string]interface{}, error) {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doGet(ctx, "/contact/v3/users/"+userID, token)
	if err != nil {
		return nil, fmt.Errorf("feishu get user: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}
