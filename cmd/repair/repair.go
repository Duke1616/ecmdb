package repair

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/cmd/repair/ioc"
	"github.com/Duke1616/ecmdb/internal/attribute"
	"github.com/Duke1616/ecmdb/internal/model"
	"github.com/Duke1616/ecmdb/internal/resource"
	"github.com/ecodeclub/ekit/slice"
	"github.com/spf13/cobra"
)

var (
	dryRun  bool
	execute bool
)

var Cmd = &cobra.Command{
	Use:   "repair",
	Short: "修复数据加密字段",
	Long:  "对历史数据进行字段加密处理，将未加密但需要加密的字段进行加密",
	RunE:  runRepair,
}

func init() {
	Cmd.Flags().BoolVar(&dryRun, "dry-run", true, "干跑模式，不实际修改数据 (默认开启)")
	Cmd.Flags().BoolVar(&execute, "execute", false, "实际执行修复，会修改数据")
}

// runRepair 执行修复命令
func runRepair(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// 确定是否使用干跑模式
	// 如果指定了 --execute，则覆盖默认的 dry-run 模式
	actualDryRun := dryRun && !execute

	// 初始化服务
	app, err := ioc.InitApp()
	if err != nil {
		return fmt.Errorf("初始化服务失败: %w", err)
	}

	// 创建修复器
	repairer := NewFieldEncryptionRepairer(app.ModelSvc, app.AttrSvc, app.ResourceSvc, actualDryRun)

	// 执行修复
	return repairer.Repair(ctx)
}

// FieldEncryptionRepairer 字段加密修复器
type FieldEncryptionRepairer struct {
	modelSvc    model.Service
	attrSvc     attribute.Service
	resourceSvc resource.EncryptedSvc
	dryRun      bool
}

// NewFieldEncryptionRepairer 创建字段加密修复器
func NewFieldEncryptionRepairer(
	modelSvc model.Service,
	attrSvc attribute.Service,
	resourceSvc resource.EncryptedSvc,
	dryRun bool,
) *FieldEncryptionRepairer {
	return &FieldEncryptionRepairer{
		modelSvc:    modelSvc,
		attrSvc:     attrSvc,
		resourceSvc: resourceSvc,
		dryRun:      dryRun,
	}
}

// Repair 执行修复
func (r *FieldEncryptionRepairer) Repair(ctx context.Context) error {
	fmt.Println("🔧 开始执行字段加密修复...")
	if r.dryRun {
		fmt.Println("🔍 运行在干跑模式，不会实际修改数据")
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		fmt.Printf("✅ 字段加密修复完成，耗时: %v\n", duration)
	}()

	// 获取需要处理的模型
	models, err := r.getModelsWithSecureFields(ctx)
	if err != nil {
		return fmt.Errorf("获取模型信息失败: %w", err)
	}

	if len(models) == 0 {
		fmt.Println("ℹ️  没有找到需要处理的模型，跳过修复")
		return nil
	}

	// 处理每个模型
	stats := &StatsRepair{}
	for modelUid, secureFields := range models {
		modelStats, err := r.processModel(ctx, modelUid, secureFields)
		if err != nil {
			fmt.Printf("❌ 处理模型 %s 失败: %v\n", modelUid, err)
			continue
		}
		stats.Add(modelStats)
	}

	// 输出统计信息
	stats.PrintSummary()
	return nil
}

// getModelsWithSecureFields 获取需要处理的模型
func (r *FieldEncryptionRepairer) getModelsWithSecureFields(ctx context.Context) (map[string][]string, error) {
	fmt.Println("📋 正在获取所有模型...")
	models, err := r.modelSvc.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取模型列表失败: %w", err)
	}

	if len(models) == 0 {
		return nil, nil
	}

	// 获取模型UID列表
	modelUIds := slice.Map(models, func(idx int, src model.Model) string {
		return src.UID
	})

	// 获取加密字段配置
	fmt.Println("🔐 正在获取加密字段配置...")
	secureFields, err := r.attrSvc.SearchAttributeFieldsBySecure(ctx, modelUIds)
	if err != nil {
		return nil, fmt.Errorf("获取加密字段配置失败: %w", err)
	}

	// 过滤出有加密字段的模型
	modelsWithSecureFields := make(map[string][]string)
	for modelUid, fields := range secureFields {
		if len(fields) > 0 {
			modelsWithSecureFields[modelUid] = fields
			fmt.Printf("📝 模型 %s 有 %d 个加密字段: %v\n", modelUid, len(fields), fields)
		}
	}

	fmt.Printf("🎯 共找到 %d 个模型需要处理加密字段\n", len(modelsWithSecureFields))
	return modelsWithSecureFields, nil
}

// processModel 处理单个模型
func (r *FieldEncryptionRepairer) processModel(ctx context.Context, modelUid string, secureFields []string) (*StatsRepair, error) {
	fmt.Printf("\n🔄 正在处理模型: %s\n", modelUid)

	// 创建字段处理器
	processor := NewFieldProcessor(r.resourceSvc, r.dryRun)

	// 批量处理资源
	stats, err := processor.ProcessResources(ctx, modelUid, secureFields)
	if err != nil {
		return nil, fmt.Errorf("处理资源失败: %w", err)
	}

	fmt.Printf("✅ 模型 %s 处理完成: 处理 %d 条，更新 %d 条\n", modelUid, stats.Processed, stats.Updated)
	return stats, nil
}

// FieldProcessor 字段处理器
type FieldProcessor struct {
	resourceSvc resource.EncryptedSvc
	dryRun      bool
}

// NewFieldProcessor 创建字段处理器
func NewFieldProcessor(resourceSvc resource.EncryptedSvc, dryRun bool) *FieldProcessor {
	return &FieldProcessor{
		dryRun:      dryRun,
		resourceSvc: resourceSvc,
	}
}

// ProcessResources 处理资源
func (p *FieldProcessor) ProcessResources(
	ctx context.Context,

	modelUid string,
	secureFields []string,
) (*StatsRepair, error) {
	const batchSize = 100
	offset := int64(0)
	stats := &StatsRepair{}

	for {
		resources, _, err := p.resourceSvc.ListResource(ctx, secureFields, modelUid, offset, batchSize)

		if err != nil {
			return stats, fmt.Errorf("获取资源列表失败: %w", err)
		}

		if len(resources) == 0 {
			break
		}

		// 处理这批资源
		batchStats := p.processBatch(ctx, resources)
		stats.Add(batchStats)

		// 如果不足一页，说明到末尾了
		if int64(len(resources)) < batchSize {
			break
		}

		offset += batchSize

		// 显示进度
		if stats.Processed%1000 == 0 {
			fmt.Printf("📊 已处理 %d 条资源...\n", stats.Processed)
		}
	}

	return stats, nil
}

// processBatch 处理一批资源
func (p *FieldProcessor) processBatch(
	ctx context.Context,
	resources []resource.Resource,
) *StatsRepair {
	stats := &StatsRepair{}

	if p.dryRun {
		// 干跑模式：只统计需要更新的资源
		for _, r := range resources {
			stats.Processed++
			// 简单检查：如果资源有数据，就认为需要更新
			if len(r.Data) > 0 {
				stats.Updated++
				fmt.Printf("🔍 [干跑模式] 资源 ID %d 将被处理\n", r.ID)
			}
		}
	} else {
		// 实际执行：使用批量更新
		stats.Processed = len(resources)

		// 直接使用 BatchUpdateResources，它会自动处理加密
		updated, err := p.resourceSvc.BatchUpdateResources(ctx, resources)
		if err != nil {
			fmt.Printf("⚠️  批量更新资源失败: %v\n", err)
		} else {
			stats.Updated = int(updated)
		}
	}

	return stats
}

// StatsRepair 修复统计信息
type StatsRepair struct {
	Processed int
	Updated   int
}

// Add 添加统计信息
func (s *StatsRepair) Add(other *StatsRepair) {
	s.Processed += other.Processed
	s.Updated += other.Updated
}

// PrintSummary 打印统计摘要
func (s *StatsRepair) PrintSummary() {
	fmt.Printf("\n🎉 修复完成! 总计处理 %d 条资源，更新 %d 条\n", s.Processed, s.Updated)
}
