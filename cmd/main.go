package main

import (
	"log"
	"net/http"
	"order-service/internal/cache"    // <-- Импорт кэша
	"order-service/internal/config"
	"order-service/internal/repository"
)

func main() {
	// ... (загрузка конфигурации и подключение к БД)
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	repo, err := repository.NewOrderRepository(cfg.DBSource)
	if err != nil {
		log.Fatalf("Не удалось подключиться к БД: %v", err)
	}
	defer repo.Close()
	log.Println("Подключение к PostgreSQL установлено")


	// 3. Инициализация кэша
	orderCache := cache.New()
	log.Println("Кэш инициализирован")

	// 4. Восстановление кэша из БД
	log.Println("Восстановление кэша из базы данных...")
	orders, err := repo.GetAllOrders()
	if err != nil {
		log.Fatalf("Не удалось восстановить кэш: %v", err)
	}
	orderCache.LoadFromDB(orders) // Загружаем все найденные заказы в кэш
	
	// ... (остальной код)

	log.Println("Запуск HTTP-сервера на порту :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Ошибка запуска HTTP-сервера: %v", err)
	}
}