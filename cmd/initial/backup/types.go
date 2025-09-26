package backup

import (
	"time"
)

// Meta 备份元数据
type Meta struct {
	BackupID     string            `bson:"backup_id"`      // 备份批次ID
	Version      string            `bson:"version"`        // 版本号
	BackupTime   int64             `bson:"backup_time"`    // 备份时间戳
	BackupDate   string            `bson:"backup_date"`    // 备份日期
	Description  string            `bson:"description"`    // 备份描述
	SourceName   string            `bson:"source_name"`    // 源名称（表名或集合名）
	BackupType   string            `bson:"backup_type"`    // 备份类型
	TotalRecords int64             `bson:"total_records"`  // 总记录数
	Status       string            `bson:"status"`         // 备份状态
	Tags         map[string]string `bson:"tags,omitempty"` // 额外标签
	CreatedAt    time.Time         `bson:"created_at"`     // 创建时间
}

// MongoBackupData MongoDB 备份数据
type MongoBackupData struct {
	BackupID   string      `bson:"backup_id"`   // 关联的备份ID
	OriginalID interface{} `bson:"original_id"` // 原始文档ID
	Data       interface{} `bson:"data"`        // 原始数据
	CreatedAt  time.Time   `bson:"created_at"`  // 创建时间
}

// MySQLBackupData MySQL 备份数据
type MySQLBackupData struct {
	BackupID     string      `bson:"backup_id"`      // 关联的备份ID
	OriginalID   interface{} `bson:"original_id"`    // 原始记录ID
	Data         interface{} `bson:"data"`           // 原始数据（保留兼容性）
	SQLStatement string      `bson:"sql_statement"`  // SQL 语句
	CreatedAt    time.Time   `bson:"created_at"`     // 创建时间
}

// Result 备份结果
type Result struct {
	BackupID     string   // 备份ID
	BackupTime   int64    // 备份时间戳
	Collections  []string // 备份的集合列表
	TotalRecords int64    // 总记录数
}

// Options 备份选项
type Options struct {
	Version     string            // 版本号
	Description string            // 备份描述
	Tags        map[string]string // 额外标签
}

// Info 备份信息（用于列表显示）
type Info struct {
	BackupID    string `bson:"backup_id"`    // 备份ID
	Version     string `bson:"version"`      // 版本号
	BackupTime  int64  `bson:"backup_time"`  // 备份时间戳
	BackupDate  string `bson:"backup_date"`  // 备份日期
	Description string `bson:"description"`  // 备份描述
	SourceName  string `bson:"source_name"`  // 源名称
	BackupType  string `bson:"backup_type"`  // 备份类型
	RecordCount int64  `bson:"record_count"` // 记录数
	Status      string `bson:"status"`       // 备份状态
}

// 常量定义
const (
	// 备份类型
	BackupTypeMongoCollection = "mongo_collection"
	BackupTypeMySQLTable      = "mysql_table"

	// 备份状态
	BackupStatusCompleted = "completed"
	BackupStatusFailed    = "failed"
	BackupStatusPartial   = "partial"

	// 集合名称
	MetaCollectionName = "c_upgrade_backup_meta"
	DataCollectionName = "c_upgrade_backup_data"
)
