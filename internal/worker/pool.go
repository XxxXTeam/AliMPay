// Package worker 提供Worker池管理
// @author AliMPay Team
// @description 实现基于Worker池的任务处理模式，避免创建过多goroutine
package worker

import (
	"context"
	"sync"

	"alimpay-go/pkg/logger"

	"go.uber.org/zap"
)

// Task 定义任务接口
// @description 所有需要在Worker池中执行的任务都需要实现此接口
type Task interface {
	// Execute 执行任务
	// @param ctx 上下文，用于控制任务的生命周期
	// @return error 执行错误
	Execute(ctx context.Context) error
}

// Pool Worker池
// @description 管理固定数量的Worker goroutine，处理任务队列
type Pool struct {
	workerCount int              // Worker数量
	taskQueue   chan Task        // 任务队列
	wg          sync.WaitGroup   // 等待组，用于优雅关闭
	ctx         context.Context  // 上下文
	cancel      context.CancelFunc // 取消函数
	started     bool             // 是否已启动
	mu          sync.RWMutex     // 读写锁
}

// NewPool 创建Worker池
// @description 创建指定数量Worker的池
// @param workerCount Worker数量
// @param queueSize 任务队列大小
// @return *Pool Worker池实例
func NewPool(workerCount, queueSize int) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		workerCount: workerCount,
		taskQueue:   make(chan Task, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动Worker池
// @description 启动所有Worker goroutine开始处理任务
func (p *Pool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		logger.Warn("Worker pool already started")
		return
	}

	p.started = true

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	logger.Success("Worker pool started",
		zap.Int("worker_count", p.workerCount),
		zap.Int("queue_size", cap(p.taskQueue)))
}

// worker Worker协程
// @description 从任务队列中取出任务并执行
// @param id Worker ID
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	logger.Info("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-p.ctx.Done():
			logger.Info("Worker stopped", zap.Int("worker_id", id))
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				logger.Info("Task queue closed, worker exiting",
					zap.Int("worker_id", id))
				return
			}

			// 执行任务
			if err := task.Execute(p.ctx); err != nil {
				logger.Error("Task execution failed",
					zap.Int("worker_id", id),
					zap.Error(err))
			}
		}
	}
}

// Submit 提交任务到队列
// @description 将任务添加到任务队列，由Worker池处理
// @param task 要执行的任务
// @return error 如果队列已满或池已停止则返回错误
func (p *Pool) Submit(task Task) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.started {
		logger.Error("Cannot submit task: worker pool not started")
		return ErrPoolNotStarted
	}

	select {
	case <-p.ctx.Done():
		return ErrPoolStopped
	case p.taskQueue <- task:
		return nil
	default:
		// 队列已满，记录警告
		logger.Warn("Task queue is full, task rejected")
		return ErrQueueFull
	}
}

// TrySubmit 尝试提交任务（非阻塞）
// @description 尝试将任务添加到队列，如果队列满则立即返回
// @param task 要执行的任务
// @return bool 是否成功提交
func (p *Pool) TrySubmit(task Task) bool {
	select {
	case p.taskQueue <- task:
		return true
	default:
		return false
	}
}

// Stop 停止Worker池
// @description 停止接收新任务，等待所有Worker完成当前任务后退出
func (p *Pool) Stop() {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return
	}
	p.started = false
	p.mu.Unlock()

	logger.Info("Stopping worker pool...")

	// 取消上下文，通知所有Worker
	p.cancel()

	// 关闭任务队列
	close(p.taskQueue)

	// 等待所有Worker完成
	p.wg.Wait()

	logger.Success("Worker pool stopped")
}

// GetStats 获取池统计信息
// @description 返回Worker池的当前状态统计
// @return map[string]interface{} 统计信息
func (p *Pool) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"worker_count": p.workerCount,
		"queue_size":   cap(p.taskQueue),
		"queue_length": len(p.taskQueue),
		"started":      p.started,
	}
}

// 定义错误
var (
	ErrPoolNotStarted = &PoolError{"worker pool not started"}
	ErrPoolStopped    = &PoolError{"worker pool stopped"}
	ErrQueueFull      = &PoolError{"task queue is full"}
)

// PoolError Worker池错误
type PoolError struct {
	msg string
}

func (e *PoolError) Error() string {
	return e.msg
}

