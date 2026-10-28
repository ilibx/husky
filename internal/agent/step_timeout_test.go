package agent

import (
	"context"
	"sync"
	"testing"
)

func TestHumanStepTimeoutMinutes(t *testing.T) {
	tests := []struct {
		name string
		cfg  humanStepConfig
		want int
	}{
		{"未配置走缺省 24 小时", humanStepConfig{}, 1440},
		{"显式配置生效", humanStepConfig{TimeoutMinutes: 30}, 30},
		{"负数回落到缺省", humanStepConfig{TimeoutMinutes: -5}, 1440},
	}

	for _, tt := range tests {
		if got := humanStepTimeoutMinutes(tt.cfg); got != tt.want {
			t.Errorf("%s: got %d, want %d", tt.name, got, tt.want)
		}
	}
}

func TestHumanStepStillWaiting(t *testing.T) {
	waiting := []string{"running", "pending"}
	finished := []string{"completed", "failed", "skipped", ""}

	for _, status := range waiting {
		if !humanStepStillWaiting(status) {
			t.Errorf("状态 %q 仍在等人工处理，应判定为可超时", status)
		}
	}
	for _, status := range finished {
		if humanStepStillWaiting(status) {
			t.Errorf("状态 %q 已结束，不该再判超时", status)
		}
	}
}

// TestHumanStepTimeoutPredicateRegression 是一个回归测试：
// 超时判定曾经写成「到点时状态 != pending 就放弃」，而 executeHumanStep 在启动
// 定时器之前就把状态置成了 running，于是谓词恒成立、timeout_minutes 永远不生效。
func TestHumanStepTimeoutPredicateRegression(t *testing.T) {
	// 定时器到点时，步骤的真实状态
	const statusAtTimeout = "running"

	if !humanStepStillWaiting(statusAtTimeout) {
		t.Fatal("running 状态必须判定为「仍在等人工处理」，否则超时永远不会触发")
	}
}

func TestHumanStepTimers(t *testing.T) {
	var timers humanStepTimers

	cancelled := make(map[uint]int)
	var mu sync.Mutex
	mk := func(id uint) context.CancelFunc {
		return func() {
			mu.Lock()
			cancelled[id]++
			mu.Unlock()
		}
	}

	timers.register(1, mk(1))
	timers.register(2, mk(2))
	if got := timers.pending(); got != 2 {
		t.Fatalf("注册两个定时器后 pending = %d, want 2", got)
	}

	timers.cancel(1)
	if got := timers.pending(); got != 1 {
		t.Errorf("取消后 pending = %d, want 1", got)
	}
	if cancelled[1] != 1 {
		t.Errorf("cancel(1) 应调用一次 cancel func, got %d", cancelled[1])
	}

	// 重复取消同一个 ID 不应 panic，也不应重复调用
	timers.cancel(1)
	if cancelled[1] != 1 {
		t.Errorf("重复取消不该重复调用 cancel func, got %d", cancelled[1])
	}

	// 取消不存在的 ID 是空操作
	timers.cancel(999)
	if cancelled[999] != 0 {
		t.Errorf("取消不存在的 ID 不应触发任何 cancel func")
	}

	// 同一步骤重复注册（例如打回重跑）会覆盖旧定时器，旧的不再被追踪
	timers.register(2, mk(2))
	if got := timers.pending(); got != 1 {
		t.Errorf("重复注册同一 ID 后 pending = %d, want 1", got)
	}
	timers.cancel(2)
	if cancelled[2] != 1 {
		t.Errorf("应只调用最新的 cancel func 一次, got %d", cancelled[2])
	}

	if got := timers.pending(); got != 0 {
		t.Errorf("全部取消后 pending = %d, want 0", got)
	}

	// 并发注册 / 取消不应 panic 或丢状态
	var wg sync.WaitGroup
	for i := uint(100); i < 200; i++ {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			timers.register(id, mk(id))
		}(i)
	}
	wg.Wait()
	if got := timers.pending(); got != 100 {
		t.Fatalf("并发注册后 pending = %d, want 100", got)
	}
	for i := uint(100); i < 200; i++ {
		wg.Add(1)
		go func(id uint) {
			defer wg.Done()
			timers.cancel(id)
		}(i)
	}
	wg.Wait()
	if got := timers.pending(); got != 0 {
		t.Errorf("并发取消后 pending = %d, want 0", got)
	}
}
