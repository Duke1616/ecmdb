package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/gotomicro/ego/core/elog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoBackupProvider MongoDB 备份提供者
type MongoBackupProvider struct {
	app    *ioc.App
	logger *elog.Component
}

// NewMongoBackupProvider 创建 MongoDB 备份提供者
func NewMongoBackupProvider(app *ioc.App) Provider {
	return &MongoBackupProvider{
		app:    app,
		logger: elog.DefaultLogger,
	}
}

// GetProviderType 获取提供者类型
func (p *MongoBackupProvider) GetProviderType() string {
	return "mongo"
}

// Backup 备份 MongoDB 集合
func (p *MongoBackupProvider) Backup(ctx context.Context, collectionName string, opts Options) (*Result, error) {
	p.logger.Info("开始备份 MongoDB 集合",
		elog.String("集合", collectionName),
		elog.String("版本", opts.Version))

	// 使用分离的备份集合
	metaCollectionName := MetaCollectionName
	dataCollectionName := DataCollectionName

	// 获取原集合和备份集合
	sourceCollection := p.app.DB.Collection(collectionName)
	metaCollection := p.app.DB.Collection(metaCollectionName)
	dataCollection := p.app.DB.Collection(dataCollectionName)

	// 查询所有数据
	cursor, err := sourceCollection.Find(ctx, bson.M{})
	if err != nil {
		p.logger.Error("查询集合数据失败",
			elog.String("集合", collectionName),
			elog.FieldErr(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	// 使用 cursor.All() 获取所有数据
	var documents []bson.M
	if err = cursor.All(ctx, &documents); err != nil {
		p.logger.Error("解码集合数据失败",
			elog.String("集合", collectionName),
			elog.FieldErr(err))
		return nil, err
	}

	backupTime := time.Now()
	backupID := p.generateBackupID(opts.Version, collectionName)

	// 创建备份元数据
	backupMeta := Meta{
		BackupID:     backupID,
		Version:      opts.Version,
		BackupTime:   backupTime.UnixMilli(),
		BackupDate:   backupTime.Format("2006-01-02 15:04:05"),
		Description:  opts.Description,
		SourceName:   collectionName,
		BackupType:   BackupTypeMongoCollection,
		TotalRecords: int64(len(documents)),
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

	if len(documents) == 0 {
		p.logger.Info("集合为空，只保存元数据", elog.String("集合", collectionName))
		return &Result{
			BackupID:     backupID,
			BackupTime:   backupTime.UnixMilli(),
			Collections:  []string{metaCollectionName},
			TotalRecords: 0,
		}, nil
	}

	// 为每个文档创建数据记录
	var dataDocs []interface{}
	for _, doc := range documents {
		dataDoc := MongoBackupData{
			BackupID:   backupID,
			OriginalID: doc["_id"],
			Data:       doc,
			CreatedAt:  backupTime,
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

	p.logger.Info("MongoDB 集合备份完成",
		elog.String("原集合", collectionName),
		elog.String("元数据集合", metaCollectionName),
		elog.String("数据集合", dataCollectionName),
		elog.Int("备份数量", len(dataDocs)))

	return &Result{
		BackupID:     backupID,
		BackupTime:   backupTime.UnixMilli(),
		Collections:  []string{metaCollectionName, dataCollectionName},
		TotalRecords: int64(len(dataDocs)),
	}, nil
}

// Restore 恢复 MongoDB 集合
func (p *MongoBackupProvider) Restore(ctx context.Context, collectionName string, backupID string) error {
	p.logger.Info("开始恢复 MongoDB 集合",
		elog.String("集合", collectionName),
		elog.String("备份ID", backupID))

	dataCollection := p.app.DB.Collection(DataCollectionName)
	targetCollection := p.app.DB.Collection(collectionName)

	// 查询指定备份ID的数据
	filter := bson.M{
		"backup_id": backupID,
	}

	cursor, err := dataCollection.Find(ctx, filter)
	if err != nil {
		p.logger.Error("查询备份数据失败",
			elog.String("备份ID", backupID),
			elog.FieldErr(err))
		return err
	}
	defer cursor.Close(ctx)

	var backupDocs []bson.M
	if err = cursor.All(ctx, &backupDocs); err != nil {
		p.logger.Error("解码备份数据失败",
			elog.String("备份ID", backupID),
			elog.FieldErr(err))
		return err
	}

	if len(backupDocs) == 0 {
		p.logger.Info("没有找到备份数据", elog.String("备份ID", backupID))
		return nil
	}

	// 提取原始数据并准备批量替换操作
	var operations []mongo.WriteModel
	for _, backupDoc := range backupDocs {
		if data, ok := backupDoc["data"].(bson.M); ok {
			// 使用 _id 作为过滤条件进行替换
			filter := bson.M{"_id": data["_id"]}
			
			// 创建 ReplaceOneModel 操作
			operation := mongo.NewReplaceOneModel().
				SetFilter(filter).
				SetReplacement(data).
				SetUpsert(true)
			operations = append(operations, operation)
		}
	}

	// 执行批量替换操作
	if len(operations) > 0 {
		// 设置批量写入选项
		opts := options.BulkWrite().
			SetOrdered(false) // 允许并行执行，提高性能
		
		_, err = targetCollection.BulkWrite(ctx, operations, opts)
		if err != nil {
			p.logger.Error("批量替换文档失败",
				elog.String("目标集合", collectionName),
				elog.Int("操作数量", len(operations)),
				elog.FieldErr(err))
			return err
		}
	}

	p.logger.Info("MongoDB 集合恢复完成",
		elog.String("备份ID", backupID),
		elog.String("目标集合", collectionName),
		elog.Int("恢复数量", len(operations)))

	return nil
}

// ListBackups 列出备份
func (p *MongoBackupProvider) ListBackups(ctx context.Context, sourceName string) ([]Info, error) {
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
func (p *MongoBackupProvider) CleanupOldBackups(ctx context.Context, sourceName string, keepDays int) error {
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

// generateBackupID 生成备份批次ID
func (p *MongoBackupProvider) generateBackupID(version, sourceName string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s_%s", version, sourceName, timestamp)
}
