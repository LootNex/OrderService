package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/LootNex/OrderService/Consumer/configs"
	"github.com/LootNex/OrderService/Consumer/internal/db/postgresql"
	"github.com/LootNex/OrderService/Consumer/internal/db/redis"
	"github.com/LootNex/OrderService/Consumer/internal/handlers"
	"github.com/LootNex/OrderService/Consumer/internal/kafka/consumer"
	"github.com/LootNex/OrderService/Consumer/internal/logger"
	"github.com/LootNex/OrderService/Consumer/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func StartServer() error {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log, err := logger.InitLogger()
	if err != nil {
		return err
	}

	cfg, err := config.InitConfig()
	if err != nil {
		return err
	}

	PgConn, err := postgresql.InitPostgres(cfg, log)
	if err != nil {
		return err
	}

	RedisConn, err := redis.InitRedis(ctx, cfg)
	if err != nil {
		return err
	}

	pgstorage := postgresql.NewPGStorage(PgConn, log)
	CacheStorage := redis.NewCacheStorage(RedisConn)
	serv := service.NewOrderService(pgstorage, CacheStorage, log)
	OrderHandler := handlers.NewHandler(serv, log)

	if err = serv.LoadCache(ctx); err != nil {
		return fmt.Errorf("cannot load cache: %w", err)
	}

	go consumer.StartConsumer(ctx, cfg.Kafka.Topic, cfg.Kafka.Brokers, serv, log)

	r := mux.NewRouter()

	HttpServer := http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handlers.CORS(r),
	}

	r.HandleFunc("/order/{id}", OrderHandler.GetOrder).Methods("GET")

	go func() {

		log.Info("server running on port" + cfg.Server.Port)

		if err := HttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("cannot start http server err:%v", zap.Error(err))
			stop()
		}
	}()

	<-ctx.Done()

	log.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := HttpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("cannot stop server err:%w", err)
	}

	if err := PgConn.Close(); err != nil {
		log.Error("cannot close postgres", zap.Error(err))
	}
	if err := RedisConn.Close(); err != nil {
		log.Error("cannot close redis", zap.Error(err))
	}

	log.Info("Server stopped")
	return nil

}
