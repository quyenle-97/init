package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/quyenle-97/init/cfg"
	"github.com/quyenle-97/init/migrations"
	"github.com/quyenle-97/init/pkgs/log"
	"github.com/quyenle-97/init/pkgs/rdbms"
	"github.com/quyenle-97/init/server"
	"github.com/sirupsen/logrus"
)

func main() {
	// Tải cấu hình
	c := cfg.LoadConfig()

	// Khởi tạo logger
	logger, err := log.NewMultiLogger(logrus.TraceLevel)
	if err != nil {
		panic(err)
	}

	// Kết nối đến cơ sở dữ liệu
	db, err := rdbms.NewDB(c.DBDriver, c.DB, true)
	if err != nil {
		panic(err)
	}

	// Chạy migrations
	migration := rdbms.NewMigrationTool(db)
	lists := migrations.MigrationLists()
	migration.Migrate(lists)

	// Kết nối đến Redis
	//redisCache, err := cache.NewRedis(cache.RConfig{
	//	Host:    c.RConfig.Host,
	//	Port:    c.RPort(),
	//	Pass:    c.RConfig.Pass,
	//	Index:   c.RIndex(),
	//	Cluster: c.RCluster(),
	//}, logger)
	//if err != nil {
	//	panic(err)
	//}

	// Thiết lập router
	r := server.Routing(c, db, logger, nil)

	// Khởi tạo HTTP server
	address := *flag.String("listen", ":"+strconv.Itoa(c.GetPort()), "Listen address.")
	httpServer := http.Server{
		Addr:    address,
		Handler: server.AppMiddleware(r, logger),
	}

	// Xử lý graceful shutdown
	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		logger.Info("bắt đầu graceful shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err = httpServer.Shutdown(ctx); err != nil {
			panic(err)
		}
		close(idleConnectionsClosed)
	}()

	logger.Info(fmt.Sprintf("Đang lắng nghe tại cổng %s", c.Server.Port))
	if err = httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}

	<-idleConnectionsClosed
}
