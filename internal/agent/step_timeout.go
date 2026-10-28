package agent

import (
	"context"
	"sync"
)

// defaultHumanStepTimeoutMinutes 是人工步骤缺省超时：24 小时。
const defaultHumanStepTimeoutMinutes = 1440

// humanStepConfig 是 human 类型步骤的配置。
type humanStepConfig struct {
	TimeoutMinutes int `json:"timeout_minutes"`
}

// humanStepTimeoutMinutes 返回人工步骤的超时分钟数，非正数回落到缺省值。
func humanStepTimeoutMinutes(cfg humanStepConfig) int {
	if cfg.TimeoutMinutes <= 0 {
		return defaultHumanStepTimeoutMinutes
	}
	return cfg.TimeoutMinutes
}

// humanStepStillWaiting 判断超时到点时步骤是否仍在等人工处理。
//
// executeHumanStep 在启动定时器之前就把状态置成了 running，所以到点时状态是
// running 或 pending；其余状态（completed / failed / skipped）说明这一步已经被
// 审批、驳回或跳过，不该再判超时。
func humanStepStillWaiting(status string) bool {
	return status == "running" || status == "pending"
}

// humanStepTimers 管理人工步骤的超时定时器。
//
// 除了「注册 + 取消」，它还负责一件事：审批打回重跑时必须取消上一个定时器，
// 否则旧定时器到点后会把新一轮尝试判成超时。用结构体而不是裸 sync.Map，
// 是为了让「取消」只有一处实现，不会在某个分支漏掉。
type humanStepTimers struct {
	mu sync.Mutex
	m  map[uint]context.CancelFunc
}

func (t *humanStepTimers) register(stepID uint, cancel context.CancelFunc) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.m == nil {
		t.m = make(map[uint]context.CancelFunc)
	}
	t.m[stepID] = cancel
}

func (t *humanStepTimers) cancel(stepID uint) {
	t.mu.Lock()
	cancel, ok := t.m[stepID]
	if ok {
		delete(t.m, stepID)
	}
	t.mu.Unlock()
	if ok && cancel != nil {
		cancel()
	}
}

// pending 返回当前登记在册的定时器数量，仅供测试断言。
func (t *humanStepTimers) pending() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.m)
}
