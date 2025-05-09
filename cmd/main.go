package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/quyenle-97/init/cfg"
	"github.com/quyenle-97/init/internal/models"
	"github.com/quyenle-97/init/migrations"
	"github.com/quyenle-97/init/pkgs/cache"
	"github.com/quyenle-97/init/pkgs/log"
	"github.com/quyenle-97/init/pkgs/rdbms"
	"github.com/quyenle-97/init/server"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {

	c := cfg.LoadConfig()
	logger, err := log.NewMultiLogger(logrus.TraceLevel)
	if err != nil {
		panic(err)
	}
	db, err := rdbms.NewDB(c.DBDriver, c.DB, true)
	if err != nil {
		panic(err)
	}
	models.Init(db)
	migration := rdbms.NewMigrationTool(db)
	lists := migrations.MigrationLists()
	migration.Migrate(lists)

	cache, err := cache.NewRedis(cache.RConfig{
		Host:    c.RConfig.Host,
		Port:    c.RPort(),
		Pass:    c.RConfig.Pass,
		Index:   c.RIndex(),
		Cluster: c.RCluster(),
	}, logger)
	if err != nil {
		panic(err)
	}

	r := server.Routing(c, db, logger, cache)
	address := *flag.String("listen", ":"+strconv.Itoa(c.GetPort()), "Listen address.")
	httpServer := http.Server{
		Addr:    address,
		Handler: server.AppMiddleware(r, logger),
	}

	idleConnectionsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		logger.Info("start graceful shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err = httpServer.Shutdown(ctx); err != nil {
			panic(err)
		}
		close(idleConnectionsClosed)
	}()

	logger.Info(fmt.Sprintf("Listening at port %s", c.Server.Port))
	if err = httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}

	<-idleConnectionsClosed
}
