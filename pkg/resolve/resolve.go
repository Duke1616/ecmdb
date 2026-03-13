package resolve

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/Duke1616/ecmdb/internal/user"
	"github.com/ecodeclub/ekit/slice"
	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
)

// Target 是一种分配策略抽象，代表向哪个维度(Type)分配哪些实体(Ids/Fields)
type Target struct {
	Type   string   `json:"type"`   // 匹配策略，例如: 'appoint', 'role', 'leader' 等
	Values []string `json:"values"` // 规则的目标值列表
}

// Resolver 策略接口：每个实现都应自我声明其 Name（唯一规则类型标识），以及对应的 Resolve 逻辑
type Resolver interface {
	// Name 返回该解析器所覆盖的规则唯一标识，用于匹配 Target.Type
	Name() string
	// Resolve 根据 Target 解析出对应的用户列表
	Resolve(ctx context.Context, target Target) ([]user.User, error)
}

// Engine 规则解析引擎接口，负责将各种维度的分配规则（Target）解析为具体的系统用户
//
//go:generate mockgen -source=./resolve.go -package=resolveblocks -destination=./mocks/resolve.mock.go -typed Engine
type Engine interface {
	// Register 注册一个或多个解析策略实现
	Register(resolvers ...Resolver) Engine
	// Resolve 并发解析一组目标，任一目标解析失败即返回错误（底层调用 ResolveWithErrorHandling 并开启 failFast）
	Resolve(ctx context.Context, targets []Target) ([]user.User, error)
	// ResolveWithErrorHandling 提供更细粒度的解析控制，支持配置 failFast（快速失败）模式
	// failFast 为 true 时，任一解析任务失败将立即取消其他进行中的任务并返回
	ResolveWithErrorHandling(ctx context.Context, targets []Target, failFast bool) ([]user.User, error)
}

// engine 规则解析引擎（提供策略注册、并发调度与错误聚合功能）
type engine struct {
	resolvers map[string]Resolver
}

// NewEngine 构造解析引擎
func NewEngine() Engine {
	return &engine{
		resolvers: make(map[string]Resolver),
	}
}

// Register 注册一个或多个解析器，Key 由 resolver.Name() 自动推断，无需手动传递 ruleType
func (e *engine) Register(resolvers ...Resolver) Engine {
	for _, r := range resolvers {
		e.resolvers[r.Name()] = r
	}
	return e
}

// resolveResult 解析结果载体
type resolveResult struct {
	users []user.User
	err   error
}

// Resolve 批量解析所有的分配规则（默认遇到错误直接短路）
func (e *engine) Resolve(ctx context.Context, targets []Target) ([]user.User, error) {
	return e.ResolveWithErrorHandling(ctx, targets, true)
}

// ResolveWithErrorHandling 批量解析所有的分配规则
// failFast 参数决定是否在任何一个规则解析出错时立即中断退回错误，
// 若为 false，则收集所有的解析错误并尽可能返回部分成功的人员列表。
func (e *engine) ResolveWithErrorHandling(ctx context.Context, targets []Target, failFast bool) ([]user.User, error) {
	if len(targets) == 0 {
		return nil, nil
	}

	results := e.resolveConcurrently(ctx, targets, failFast)

	allUsers := e.mergeResults(results)
	multiErr := e.collectErrors(results)

	if multiErr != nil && failFast {
		return allUsers, multiErr.ErrorOrNil()
	}

	return allUsers, nil
}

// resolveConcurrently 并发调度解析过程
func (e *engine) resolveConcurrently(ctx context.Context, targets []Target, failFast bool) []resolveResult {
	var wg sync.WaitGroup
	resultChan := make(chan resolveResult, len(targets))

	// 如果需要 failFast，衍生出可取消的 Context
	execCtx := ctx
	var cancel context.CancelFunc
	if failFast {
		execCtx, cancel = context.WithCancel(ctx)
		defer cancel()
	}

	for _, t := range targets {
		wg.Add(1)
		go func(target Target) {
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					resultChan <- resolveResult{
						err: fmt.Errorf("panic in rule [%s]: %v\nstack:\n%s", target.Type, r, debug.Stack()),
					}
					if failFast && cancel != nil {
						cancel()
					}
				}
			}()

			// 检查是否已被其他协程取消
			if failFast && execCtx.Err() != nil {
				return
			}

			resolver, ok := e.resolvers[target.Type]
			if !ok {
				err := fmt.Errorf("unsupported rule type: '%s' (values: %v)",
					lo.Ternary(target.Type != "", target.Type, "<EMPTY>"), target.Values)
				resultChan <- resolveResult{err: err}
				if failFast && cancel != nil {
					cancel()
				}
				return
			}

			users, err := resolver.Resolve(execCtx, target)
			resultChan <- resolveResult{users: users, err: err}

			if err != nil && failFast && cancel != nil {
				cancel()
			}
		}(t)
	}

	// 等待并关闭通道
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var results []resolveResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// mergeResults 合并并去重解析出来的用户实体
func (e *engine) mergeResults(results []resolveResult) []user.User {
	var allUsers []user.User
	for _, result := range results {
		if result.err != nil {
			continue
		}
		// 比较结构体属性并合并去重
		allUsers = slice.UnionSetFunc(allUsers, result.users, func(src, dst user.User) bool {
			return src.Id == dst.Id
		})
	}

	// 过滤并剔除空的无效数据对象
	return slice.FilterMap(allUsers, func(idx int, src user.User) (user.User, bool) {
		return src, src.Id != 0
	})
}

// collectErrors 采集所有的错误
func (e *engine) collectErrors(results []resolveResult) *multierror.Error {
	var multiErr *multierror.Error
	for _, result := range results {
		if result.err == nil {
			continue
		}
		multiErr = multierror.Append(multiErr, result.err)
	}
	return multiErr
}
