// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package model

// 基础设置
type NoaConfig struct {
	DataBaseType string   `json:"dataBaseType" yaml:"DataBaseType"` // 数据库类型
	Dsn          string   `json:"dsn" yaml:"DSN"`                   // 数据库连接 DSN 信息
	SyncPath     []string `json:"syncPath" yaml:"SyncPath"`         // 需要迁移的路径，写ID内路径（web 上面能看到的）
	UserID       []string `json:"userID" yaml:"UserId"`             // 迁移用户的 ID
	S3Endpoint   string   `json:"s3Endpoint" yaml:"S3Endpoint"`     // S3 存储的 Endpoint
	S3AccessKey  string   `json:"s3AccessKey" yaml:"S3AccessKey"`   // S3 存储的 Access Key
	S3SecretKey  string   `json:"s3SecretKey" yaml:"S3SecretKey"`   // S3 存储的 Secret Key
	S3Bucket     string   `json:"s3Bucket" yaml:"S3Bucket"`         // S3 存储桶，需要提前手动创建好
	Cron         string   `json:"cron" yaml:"Cron"`                 // 同步任务开始的 Cron 表达式
	APIPort      string   `json:"apiport" yaml:"APIPort"`           // Prometheus Exporter http port
}

type CheckPoint struct {
	StatusCode int    `json:"statusCode"`
	AppName    string `json:"appName"`
	Version    string `json:"version"`
}

type FlieInfo struct {
	FilePath           string
	FileSize           uint64
	FileType, FileMIME string
	CheckSumMD5        string // SHA256 ,用来和 MinIO侧校验

	// 来自 Cloudreve 的信息
	Size     uint64 `xorm:"'size'"`      // 数据库内的文件大小
	PolicyId int    `xorm:"'policy_id'"` // 关联的存储策略
}
type Policies struct {
	Id   int
	Type string // 存储策略，本地 -> local
	Name string // 存储策略的名称
}

type CloudreveFile struct {
	Id         int
	Name       string
	SourceName string
	S3Path     string `xorm:"-"`
	UserId     int
}
