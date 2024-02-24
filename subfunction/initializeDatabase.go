// noaMigrator - cloudreve -> S3 同步推送工具
// @CreateTime		: 2024/02/07 21:04
// @LastModified	: 2024/02/07 21:04
// @Author			: Luckykeeper
// @Contact		    : https://github.com/luckykeeper | https://luckykeeper.site
// @Email			: luckykeeper@luckykeeper.site
// @Project		    : noaMigrator

package subfunction

import (
	"noaMigrator/model"
	"time"

	// _ "github.com/alexbrainman/odbc"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
	"xorm.io/xorm/caches"
)

// 项目数据库引擎
var CocoaDataEngine *xorm.Engine

// 初始化数据库
func InitializeDatabase(NoaConfig model.NoaConfig, debugMode bool) {
	// 防止空指针问题，这样声明 err
	// 见：https://stackoverflow.com/questions/56396386/xorm-example-not-working-runtime-error-invalid-memory-address-or-nil-pointer
	var err error
	CocoaDataEngine, err = xorm.NewEngine(NoaConfig.DataBaseType, NoaConfig.Dsn)
	if err != nil {
		logrus.Fatalln("数据库初始化失败：", err)
	}
	// 不需要 Close ，orm 会自己判断
	// defer DataEngine.Close()

	// 取消大小写敏感
	// https://pkg.go.dev/xorm.io/xorm@v1.0.1/dialects#QuotePolicy
	// https://www.cnblogs.com/bartggg/p/13066944.html
	CocoaDataEngine.Dialect().SetQuotePolicy(1)

	// 使用缓存
	cacher := caches.NewLRUCacher(caches.NewMemoryStore(), 1000)
	CocoaDataEngine.SetDefaultCacher(cacher)

	if debugMode {
		CocoaDataEngine.ShowSQL(true)
	}
	err = CocoaDataEngine.Ping()
	if err != nil {
		logrus.Panicln("数据库连接失败！检查数据库信息是否正确，程序返回原因为：", err)
	} else {
		CocoaDataEngine.SetConnMaxLifetime(time.Second * 60) // 最大连接存活时间
		CocoaDataEngine.SetMaxOpenConns(100)                 // 最大连接数
		CocoaDataEngine.SetMaxIdleConns(3)                   // 最大空闲连接数
		logrus.Infoln("连接到远程数据库成功！")
	}
}
