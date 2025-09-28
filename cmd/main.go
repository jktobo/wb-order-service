package main

import (
	"log"
	"net/http"
	"wb-order-service/internal/broker"
	"wb-order-service/internal/cache"
	"wb-order-service/internal/config"
	"wb-order-service/internal/handler"
	"wb-order-service/internal/repository"
	"wb-order-service/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	// 1. Конфигурация
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// 2. Репозиторий (БД)
	repo, err := repository.NewOrderRepository(cfg.DBSource)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer repo.Close()

	// 3. Кэш
	orderCache := cache.New()

	// 4. Восстановление кэша
	orders, err := repo.GetAllOrders()
	if err != nil {
		log.Fatalf("Не удалось восстановить кэш: %v", err)
	}
	orderCache.LoadFromDB(orders)

	// 5. Сервис (Бизнес-логика)
	orderService := service.NewOrderService(repo, orderCache)

	// 6. Kafka Consumer
	consumer := broker.NewConsumer(orderService)
	consumer.Start(cfg.KafkaBroker, cfg.KafkaTopic)

	// 7. HTTP Хендлеры и роутер
	orderHandler := handler.NewOrderHandler(orderService)
	router := mux.NewRouter()
	router.HandleFunc("/order/{order_uid}", orderHandler.GetOrderByUID).Methods("GET")

	// Сервируем статические файлы (наш index.html)
	// Важно: путь должен быть относительным от корня проекта, где запускается бинарник.
	// В Dockerfile мы копируем web/static в /app/web/static.
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static/")))

	// 8. Запуск HTTP-сервера
	log.Println("Запуск HTTP-сервера на http://localhost:8081")
	if err := http.ListenAndServe(":8081", router); err != nil {
		log.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
	}
}