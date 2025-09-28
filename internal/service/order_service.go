package service

import (
	"log"
	"wb-order-service/internal/cache"
	"wb-order-service/internal/model"
)

// OrderService определяет интерфейс для репозитория, который нам нужен.
// Это хорошая практика, чтобы не зависеть напрямую от конкретной реализации (PostgreSQL).
type OrderRepository interface {
	SaveOrder(order model.Order) error
}

type OrderService struct {
	repo  OrderRepository
	cache *cache.OrderCache
}

func NewOrderService(repo OrderRepository, cache *cache.OrderCache) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

// ProcessNewOrder обрабатывает новый заказ: сохраняет в БД и кэширует.
func (s *OrderService) ProcessNewOrder(order model.Order) {
	// 1. Сохраняем в базу данных
	err := s.repo.SaveOrder(order)
	if err != nil {
		log.Printf("ОШИБКА: не удалось сохранить заказ %s в БД: %v", order.OrderUID, err)
		return // Не кэшируем, если не смогли сохранить
	}

	// 2. Если сохранение в БД успешно, добавляем в кэш
	s.cache.Set(order.OrderUID, order)
	log.Printf("Заказ %s успешно обработан и кэширован.", order.OrderUID)
}

// GetOrderByUID получает заказ по его UID из кэша.
func (s *OrderService) GetOrderByUID(uid string) (model.Order, bool) {
	// В нашей архитектуре данные всегда должны быть в кэше.
	// Если их там нет, значит, что-то пошло не так, либо такого заказа нет.
	// Запрос в БД делать не будем, чтобы соответствовать логике "сначала кэш".
	return s.cache.Get(uid)
}