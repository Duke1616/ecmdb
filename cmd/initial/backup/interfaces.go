package backup

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/gotomicro/ego/core/elog"
)

// Provider 备份提供者接口
type Provider interface {
	// Backup 备份数据
	Backup(ctx context.Context, sourceName string, opts Options) (*Result, error)

	// Restore 恢复数据
	Restore(ctx context.Context, sourceName string, backupID string) error

	// ListBackups 列出备份
	ListBackups(ctx context.Context, sourceName string) ([]Info, error)

	// CleanupOldBackups 清理旧备份
	CleanupOldBackups(ctx context.Context, sourceName string, keepDays int) error

	// GetProviderType 获取提供者类型
	GetProviderType() string
}

// BackupManager 备份管理器
type BackupManager struct {
	providers map[string]Provider
	logger    *elog.Component
}

// NewBackupManager 创建备份管理器
func NewBackupManager(app *ioc.App) *BackupManager {
	providers := make(map[string]Provider)

	// 注册各种备份提供者
	providers["mongo"] = NewMongoBackupProvider(app)
	providers["mysql"] = NewMySQLBackupProvider(app)

	return &BackupManager{
		providers: providers,
		logger:    elog.DefaultLogger,
	}
}

// BackupMongoCollection 备份 MongoDB 集合
func (bm *BackupManager) BackupMongoCollection(ctx context.Context, collectionName string, opts Options) (*Result, error) {
	provider, exists := bm.providers["mongo"]
	if !exists {
		return nil, fmt.Errorf("MongoDB 备份提供者未注册")
	}
	return provider.Backup(ctx, collectionName, opts)
}

// BackupMySQLTable 备份 MySQL 表
func (bm *BackupManager) BackupMySQLTable(ctx context.Context, tableName string, opts Options) (*Result, error) {
	provider, exists := bm.providers["mysql"]
	if !exists {
		return nil, fmt.Errorf("MySQL 备份提供者未注册")
	}
	return provider.Backup(ctx, tableName, opts)
}

// RestoreMongoCollection 恢复 MongoDB 集合
func (bm *BackupManager) RestoreMongoCollection(ctx context.Context, collectionName string, backupID string) error {
	provider, exists := bm.providers["mongo"]
	if !exists {
		return fmt.Errorf("MongoDB 备份提供者未注册")
	}
	return provider.Restore(ctx, collectionName, backupID)
}

// RestoreMySQLTable 恢复 MySQL 表
func (bm *BackupManager) RestoreMySQLTable(ctx context.Context, tableName string, backupID string) error {
	provider, exists := bm.providers["mysql"]
	if !exists {
		return fmt.Errorf("MySQL 备份提供者未注册")
	}
	return provider.Restore(ctx, tableName, backupID)
}

// ListBackups 列出备份
func (bm *BackupManager) ListBackups(ctx context.Context, sourceName string, providerType string) ([]Info, error) {
	provider, exists := bm.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("备份提供者 %s 未注册", providerType)
	}
	return provider.ListBackups(ctx, sourceName)
}

// CleanupOldBackups 清理旧备份
func (bm *BackupManager) CleanupOldBackups(ctx context.Context, sourceName string, providerType string, keepDays int) error {
	provider, exists := bm.providers[providerType]
	if !exists {
		return fmt.Errorf("备份提供者 %s 未注册", providerType)
	}
	return provider.CleanupOldBackups(ctx, sourceName, keepDays)
}
