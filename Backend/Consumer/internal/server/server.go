package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LootNex/OrderService/Consumer/configs"
	"github.com/LootNex/OrderService/Consumer/internal/cache"
	"github.com/LootNex/OrderService/Consumer/internal/db/postgresql"
	"github.com/LootNex/OrderService/Consumer/internal/handlers"
	"github.com/LootNex/OrderService/Consumer/internal/kafka/consumer"
	"github.com/LootNex/OrderService/Consumer/internal/logger"
	"github.com/LootNex/OrderService/Consumer/internal/repository"
	"github.com/LootNex/OrderService/Consumer/internal/service"
	"github.com/gorilla/mux"
)

func StartServer() error {

	logger, err := logger.InitLogger()
	if err != nil {
		return err
	}

	cfg, err := config.InitConfig()
	if err != nil {
		return err
	}

	postg, err := postgresql.InitPosgres(cfg, logger)
	if err != nil {
		return err
	}

	pgstorage := repository.NewPGStorage(postg)
	cache := cache.NewCacheStorage()
	serv := service.NewOrderService(pgstorage, cache)
	handler := handlers.NewHandler(serv)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	go serv.LoadCache(ctx)
	go consumer.StartConsumer(ctx, cfg.Kafka.Topic, cfg.Kafka.Brokers, serv, logger)

	r := mux.NewRouter()

	server := http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handlers.CORS(r),
	}

	r.HandleFunc("/order/{id}", handler.GetOrder).Methods("GET")

	go func() {

		logger.Info("server running on port" + cfg.Server.Port)

		if err := server.ListenAndServe(); err != nil {
			logger.Fatal(fmt.Sprintf("cannot start http server err:%v", err))
			stop()
		}
	}()

	<-ctx.Done()

	logger.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("cannot stop server err:%v", err)
	}

	logger.Info("Server stopped")
	return nil

}
