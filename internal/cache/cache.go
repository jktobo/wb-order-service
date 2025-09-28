package cache
package cache

import (
	"fmt"
	"order-service/internal/model"
	"sync" // Пакет для безопасной работы с данными из нескольких горутин
)

// OrderCache - это наш in-memory кэш для заказов
type OrderCache struct {
	mu    sync.RWMutex // Мьютекс для безопасного доступа к map
	items map[string]model.Order
}

// New создает новый экземпляр кэша
func New() *OrderCache {
	return &OrderCache{
		items: make(map[string]model.Order),
	}
}

// Set добавляет или обновляет заказ в кэше
func (c *OrderCache) Set(uid string, order model.Order) {
	c.mu.Lock()         // Блокируем для записи
	defer c.mu.Unlock() // Гарантируем разблокировку в конце функции
	c.items[uid] = order
}

// Get получает заказ из кэша
func (c *OrderCache) Get(uid string) (model.Order, bool) {
	c.mu.RLock()         // Блокируем для чтения (менее строгая блокировка)
	defer c.mu.RUnlock() // Гарантируем разблокировку
	item, found := c.items[uid]
	return item, found
}

// LoadFromDB заменяет все элементы в кэше на переданный срез
func (c *OrderCache) LoadFromDB(orders []model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Создаем новую мапу, чтобы очистить старые данные
	newItems := make(map[string]model.Order)
	for _, order := range orders {
		newItems[order.OrderUID] = order
	}
	c.items = newItems
	fmt.Printf("Кэш успешно восстановлен. Загружено %d заказов.\n", len(c.items))
}