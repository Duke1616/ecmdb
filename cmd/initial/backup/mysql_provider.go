package backup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/bson"
)

// MySQLBackupProvider MySQL 备份提供者
type MySQLBackupProvider struct {
	app    *ioc.App
	logger *elog.Component
}

// NewMySQLBackupProvider 创建 MySQL 备份提供者
func NewMySQLBackupProvider(app *ioc.App) Provider {
	return &MySQLBackupProvider{
		app:    app,
		logger: elog.DefaultLogger,
	}
}

// GetProviderType 获取提供者类型
func (p *MySQLBackupProvider) GetProviderType() string {
	return "mysql"
}

// Backup 备份 MySQL 表
func (p *MySQLBackupProvider) Backup(ctx context.Context, tableName string, opts Options) (*Result, error) {
	p.logger.Info("开始备份 MySQL 表",
		elog.String("表", tableName),
		elog.String("版本", opts.Version))

	// 使用分离的备份集合
	metaCollectionName := MetaCollectionName
	dataCollectionName := DataCollectionName
	metaCollection := p.app.DB.Collection(metaCollectionName)
	dataCollection := p.app.DB.Collection(dataCollectionName)

	backupTime := time.Now()
	backupID := p.generateBackupID(opts.Version, tableName)

	// 使用 GORM 的链式调用一次性获取所有数据
	var allData []map[string]interface{}
	err := p.app.GormDB.WithContext(ctx).
		Table(tableName).
		Find(&allData).Error
	if err != nil {
		p.logger.Error("查询表数据失败",
			elog.String("表", tableName),
			elog.FieldErr(err))
		return nil, err
	}

	// 转换为 bson.M 格式
	var bsonData []bson.M
	for _, row := range allData {
		bsonRow := bson.M{}
		for k, v := range row {
			// 处理 time.Time 类型，移除时区信息
			if t, ok := v.(time.Time); ok {
				// 格式化为 MySQL 可接受的时间格式
				bsonRow[k] = t.Format("2006-01-02 15:04:05")
			} else {
				bsonRow[k] = v
			}
		}
		bsonData = append(bsonData, bsonRow)
	}

	recordCount := int64(len(bsonData))

	// 创建备份元数据
	backupMeta := Meta{
		BackupID:     backupID,
		Version:      opts.Version,
		BackupTime:   backupTime.UnixMilli(),
		BackupDate:   backupTime.Format("2006-01-02 15:04:05"),
		Description:  opts.Description,
		SourceName:   tableName,
		BackupType:   BackupTypeMySQLTable,
		TotalRecords: recordCount,
		Status:       BackupStatusCompleted,
		Tags:         opts.Tags,
		CreatedAt:    backupTime,
	}

	// 插入元数据
	_, err = metaCollection.InsertOne(ctx, backupMeta)
	if err != nil {
		p.logger.Error("插入备份元数据失败",
			elog.String("元数据集合", metaCollectionName),
			elog.FieldErr(err))
		return nil, err
	}

	if recordCount == 0 {
		p.logger.Info("表为空，只保存元数据", elog.String("表", tableName))
		return &Result{
			BackupID:     backupID,
			BackupTime:   backupTime.UnixMilli(),
			Collections:  []string{metaCollectionName},
			TotalRecords: 0,
		}, nil
	}

	// 创建数据记录 - 生成 SQL 语句
	var dataDocs []interface{}
	for _, rowData := range bsonData {
		// 生成 INSERT 语句
		sqlStatement := p.generateInsertSQL(tableName, rowData)
		
		dataDoc := MySQLBackupData{
			BackupID:     backupID,
			OriginalID:   rowData["id"],
			SQLStatement: sqlStatement,
			CreatedAt:    backupTime,
		}
		dataDocs = append(dataDocs, dataDoc)
	}

	// 批量插入数据
	_, err = dataCollection.InsertMany(ctx, dataDocs)
	if err != nil {
		p.logger.Error("插入备份数据失败",
			elog.String("数据集合", dataCollectionName),
			elog.FieldErr(err))
		return nil, err
	}

	p.logger.Info("MySQL 表备份完成",
		elog.String("表", tableName),
		elog.String("元数据集合", metaCollectionName),
		elog.String("数据集合", dataCollectionName),
		elog.Int64("记录数", recordCount))

	return &Result{
		BackupID:     backupID,
		BackupTime:   backupTime.UnixMilli(),
		Collections:  []string{metaCollectionName, dataCollectionName},
		TotalRecords: recordCount,
	}, nil
}

// Restore 恢复 MySQL 表
func (p *MySQLBackupProvider) Restore(ctx context.Context, tableName string, backupID string) error {
	p.logger.Info("开始恢复 MySQL 表",
		elog.String("表", tableName),
		elog.String("备份ID", backupID))

	dataCollection := p.app.DB.Collection(DataCollectionName)

	// 查询指定备份ID的表数据
	filter := bson.M{
		"backup_id": backupID,
	}

	cursor, err := dataCollection.Find(ctx, filter)
	if err != nil {
		p.logger.Error("查询备份数据失败",
			elog.String("表", tableName),
			elog.FieldErr(err))
		return err
	}
	defer cursor.Close(ctx)

	var backupDocs []bson.M
	if err = cursor.All(ctx, &backupDocs); err != nil {
		p.logger.Error("解码备份数据失败",
			elog.String("表", tableName),
			elog.FieldErr(err))
		return err
	}

	if len(backupDocs) == 0 {
		p.logger.Info("没有找到备份数据", elog.String("表", tableName))
		return nil
	}

	// 恢复数据 - 将 INSERT 语句转换为 REPLACE 语句
	for _, backupDoc := range backupDocs {
		sqlStatement, ok := backupDoc["sql_statement"].(string)
		if !ok {
			p.logger.Warn("跳过无效的备份数据：缺少 SQL 语句")
			continue
		}

		// 将 INSERT 转换为 REPLACE
		replaceSQL := strings.Replace(sqlStatement, "INSERT INTO", "REPLACE INTO", 1)

		// 执行 REPLACE 语句
		if err = p.app.GormDB.WithContext(ctx).Exec(replaceSQL).Error; err != nil {
			p.logger.Error("执行 REPLACE 语句失败",
				elog.String("表", tableName),
				elog.String("SQL", replaceSQL),
				elog.FieldErr(err))
			return err
		}
	}

	p.logger.Info("MySQL 表恢复完成",
		elog.String("表", tableName),
		elog.Int("恢复数量", len(backupDocs)))

	return nil
}

// ListBackups 列出备份
func (p *MySQLBackupProvider) ListBackups(ctx context.Context, sourceName string) ([]Info, error) {
	metaCollection := p.app.DB.Collection(MetaCollectionName)

	// 查询指定源的备份元数据
	filter := bson.M{
		"source_name": sourceName,
	}

	cursor, err := metaCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var backups []Meta
	if err = cursor.All(ctx, &backups); err != nil {
		return nil, err
	}

	// 转换为统一格式
	var result []Info
	for _, backup := range backups {
		result = append(result, Info{
			BackupID:    backup.BackupID,
			Version:     backup.Version,
			BackupTime:  backup.BackupTime,
			BackupDate:  backup.BackupDate,
			Description: backup.Description,
			SourceName:  backup.SourceName,
			BackupType:  backup.BackupType,
			RecordCount: backup.TotalRecords,
			Status:      backup.Status,
		})
	}

	return result, nil
}

// CleanupOldBackups 清理旧备份
func (p *MySQLBackupProvider) CleanupOldBackups(ctx context.Context, sourceName string, keepDays int) error {
	p.logger.Info("开始清理旧备份",
		elog.String("源", sourceName),
		elog.Int("保留天数", keepDays))

	cutoffTime := time.Now().AddDate(0, 0, -keepDays).UnixMilli()
	metaCollection := p.app.DB.Collection(MetaCollectionName)
	dataCollection := p.app.DB.Collection(DataCollectionName)

	// 查询需要清理的备份元数据
	filter := bson.M{
		"$and": []bson.M{
			{"source_name": sourceName},
			{"backup_time": bson.M{"$lt": cutoffTime}},
		},
	}

	// 先获取要删除的备份ID列表
	cursor, err := metaCollection.Find(ctx, filter)
	if err != nil {
		p.logger.Error("查询旧备份元数据失败",
			elog.String("源", sourceName),
			elog.FieldErr(err))
		return err
	}
	defer cursor.Close(ctx)

	var backupMetas []Meta
	if err = cursor.All(ctx, &backupMetas); err != nil {
		p.logger.Error("解码备份元数据失败",
			elog.String("源", sourceName),
			elog.FieldErr(err))
		return err
	}

	if len(backupMetas) == 0 {
		p.logger.Info("没有需要清理的旧备份")
		return nil
	}

	// 收集备份ID
	var backupIDs []string
	for _, meta := range backupMetas {
		backupIDs = append(backupIDs, meta.BackupID)
	}

	// 删除元数据
	metaResult, err := metaCollection.DeleteMany(ctx, filter)
	if err != nil {
		p.logger.Error("删除旧备份元数据失败",
			elog.String("源", sourceName),
			elog.FieldErr(err))
		return err
	}

	// 删除对应的数据
	dataFilter := bson.M{"backup_id": bson.M{"$in": backupIDs}}
	dataResult, err := dataCollection.DeleteMany(ctx, dataFilter)
	if err != nil {
		p.logger.Error("删除旧备份数据失败",
			elog.String("源", sourceName),
			elog.FieldErr(err))
		return err
	}

	p.logger.Info("清理旧备份完成",
		elog.Int64("清理元数据数量", metaResult.DeletedCount),
		elog.Int64("清理数据数量", dataResult.DeletedCount))

	return nil
}


// generateInsertSQL 生成 INSERT SQL 语句
func (p *MySQLBackupProvider) generateInsertSQL(tableName string, rowData bson.M) string {
	var columns []string
	var values []string

	for col, val := range rowData {
		columns = append(columns, col)
		
		// 简单处理，直接转换
		switch v := val.(type) {
		case string:
			// 转义单引号
			escaped := strings.ReplaceAll(v, "'", "''")
			values = append(values, fmt.Sprintf("'%s'", escaped))
		case int, int32, int64:
			values = append(values, fmt.Sprintf("%v", v))
		case float32, float64:
			values = append(values, fmt.Sprintf("%v", v))
		case bool:
			if v {
				values = append(values, "1")
			} else {
				values = append(values, "0")
			}
		case nil:
			values = append(values, "NULL")
		default:
			// 其他类型转换为字符串
			escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
			values = append(values, fmt.Sprintf("'%s'", escaped))
		}
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(values, ", "))
}


// generateBackupID 生成备份批次ID
func (p *MySQLBackupProvider) generateBackupID(version, sourceName string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s_%s", version, sourceName, timestamp)
}
