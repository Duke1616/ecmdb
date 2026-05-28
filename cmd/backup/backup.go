package backup

import (
	"context"
	"fmt"

	"github.com/Duke1616/ecmdb/cmd/initial/backup"
	"github.com/Duke1616/ecmdb/cmd/initial/ioc"
	"github.com/spf13/cobra"
)

var (
	version     string
	description string
)

var Cmd = &cobra.Command{
	Use:   "backup",
	Short: "数据备份管理",
	Long:  "执行数据库备份和恢复操作",
}

// backupMongoCmd MongoDB 备份命令
var backupMongoCmd = &cobra.Command{
	Use:   "mongo [collection]",
	Short: "备份 MongoDB 集合",
	Long:  "备份指定的 MongoDB 集合数据",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		collectionName := args[0]

		// 初始化 Ioc 注册
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 创建备份管理器
		backupManager := backup.NewBackupManager(app)

		// 备份选项
		opts := backup.Options{
			Version:     version,
			Description: description,
			Tags: map[string]string{
				"type":       "mongo",
				"collection": collectionName,
			},
		}

		fmt.Printf("💾 开始备份 MongoDB 集合...\n")
		fmt.Printf("📊 集合名称: %s\n", collectionName)
		fmt.Printf("🏷️  版本: %s\n", opts.Version)
		fmt.Printf("==================================================\n")

		// 执行备份
		result, err := backupManager.BackupMongoCollection(context.Background(), collectionName, opts)
		cobra.CheckErr(err)

		fmt.Printf("✅ MongoDB 集合备份完成!\n")
		fmt.Printf("📋 备份ID: %s\n", result.BackupID)
		fmt.Printf("📊 记录数: %d\n", result.TotalRecords)
		fmt.Printf("📅 备份时间: %d\n", result.BackupTime)
	},
}

// backupMySQLCmd MySQL 备份命令
var backupMySQLCmd = &cobra.Command{
	Use:   "mysql [table]",
	Short: "备份 MySQL 表",
	Long:  "备份指定的 MySQL 表数据",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]

		// 初始化 Ioc 注册
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 创建备份管理器
		backupManager := backup.NewBackupManager(app)

		// 备份选项
		opts := backup.Options{
			Version:     version,
			Description: description,
			Tags: map[string]string{
				"type":  "mysql",
				"table": tableName,
			},
		}

		fmt.Printf("💾 开始备份 MySQL 表...\n")
		fmt.Printf("📊 表名称: %s\n", tableName)
		fmt.Printf("🏷️  版本: %s\n", opts.Version)
		fmt.Printf("==================================================\n")

		// 执行备份
		result, err := backupManager.BackupMySQLTable(context.Background(), tableName, opts)
		cobra.CheckErr(err)

		fmt.Printf("✅ MySQL 表备份完成!\n")
		fmt.Printf("📋 备份ID: %s\n", result.BackupID)
		fmt.Printf("📊 记录数: %d\n", result.TotalRecords)
		fmt.Printf("📅 备份时间: %d\n", result.BackupTime)
	},
}

// listBackupCmd 列出备份命令
var listBackupCmd = &cobra.Command{
	Use:   "list [source] [type]",
	Short: "列出备份记录",
	Long:  "列出指定源和类型的备份记录",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		sourceName := args[0]
		providerType := args[1]

		// 初始化 Ioc 注册
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 创建备份管理器
		backupManager := backup.NewBackupManager(app)

		fmt.Printf("📋 备份记录列表\n")
		fmt.Printf("==================================================\n")
		fmt.Printf("源名称: %s\n", sourceName)
		fmt.Printf("类型: %s\n", providerType)
		fmt.Printf("==================================================\n")

		// 列出备份
		backups, err := backupManager.ListBackups(context.Background(), sourceName, providerType)
		cobra.CheckErr(err)

		if len(backups) == 0 {
			fmt.Printf("📭 没有找到备份记录\n")
			return
		}

		for i, backup := range backups {
			fmt.Printf("%d. 备份ID: %s\n", i+1, backup.BackupID)
			fmt.Printf("   版本: %s\n", backup.Version)
			fmt.Printf("   时间: %s\n", backup.BackupDate)
			fmt.Printf("   描述: %s\n", backup.Description)
			fmt.Printf("   记录数: %d\n", backup.RecordCount)
			fmt.Printf("   状态: %s\n", backup.Status)
			fmt.Printf("   ----------------------------------------\n")
		}
	},
}

// restoreMongoCmd MongoDB 恢复命令
var restoreMongoCmd = &cobra.Command{
	Use:   "restore-mongo [collection] [backup-id]",
	Short: "恢复 MongoDB 集合",
	Long:  "从备份恢复指定的 MongoDB 集合数据",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		collectionName := args[0]
		backupID := args[1]

		// 初始化 Ioc 注册
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 创建备份管理器
		backupManager := backup.NewBackupManager(app)

		fmt.Printf("🔄 开始恢复 MongoDB 集合...\n")
		fmt.Printf("📊 集合名称: %s\n", collectionName)
		fmt.Printf("📋 备份ID: %s\n", backupID)
		fmt.Printf("==================================================\n")

		// 执行恢复
		err = backupManager.RestoreMongoCollection(context.Background(), collectionName, backupID)
		cobra.CheckErr(err)

		fmt.Printf("✅ MongoDB 集合恢复完成!\n")
	},
}

// restoreMySQLCmd MySQL 恢复命令
var restoreMySQLCmd = &cobra.Command{
	Use:   "restore-mysql [table] [backup-id]",
	Short: "恢复 MySQL 表",
	Long:  "从备份恢复指定的 MySQL 表数据",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		tableName := args[0]
		backupID := args[1]

		// 初始化 Ioc 注册
		app, err := ioc.InitApp()
		cobra.CheckErr(err)

		// 创建备份管理器
		backupManager := backup.NewBackupManager(app)

		fmt.Printf("🔄 开始恢复 MySQL 表...\n")
		fmt.Printf("📊 表名称: %s\n", tableName)
		fmt.Printf("📋 备份ID: %s\n", backupID)
		fmt.Printf("==================================================\n")

		// 执行恢复
		err = backupManager.RestoreMySQLTable(context.Background(), tableName, backupID)
		cobra.CheckErr(err)

		fmt.Printf("✅ MySQL 表恢复完成!\n")
	},
}

func init() {
	// 全局标志
	Cmd.PersistentFlags().StringVarP(&version, "version", "v", "v1.0.0", "备份版本号")
	Cmd.PersistentFlags().StringVarP(&description, "description", "d", "", "备份描述")

	// 添加子命令
	Cmd.AddCommand(backupMongoCmd)
	Cmd.AddCommand(backupMySQLCmd)
	Cmd.AddCommand(listBackupCmd)
	Cmd.AddCommand(restoreMongoCmd)
	Cmd.AddCommand(restoreMySQLCmd)
}
