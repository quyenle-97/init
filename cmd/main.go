package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/Minh2009/pv_soa/cfg"
	"github.com/Minh2009/pv_soa/internal/models"
	"github.com/Minh2009/pv_soa/migrations"
	"github.com/Minh2009/pv_soa/pkgs/cache"
	"github.com/Minh2009/pv_soa/pkgs/log"
	"github.com/Minh2009/pv_soa/pkgs/rdbms"
	"github.com/Minh2009/pv_soa/server"
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

	mux := server.Routing(c, db, logger, cache)
	address := *flag.String("listen", ":"+strconv.Itoa(c.GetPort()), "Listen address.")
	httpServer := http.Server{
		Addr:    address,
		Handler: server.AppMiddleware(mux, logger),
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
