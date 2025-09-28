package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"wb-order-service/internal/model"

	_ "github.com/lib/pq"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(dataSourceName string) (*OrderRepository, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть соединение с БД: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}
	return &OrderRepository{db: db}, nil
}

func (r *OrderRepository) Close() {
	r.db.Close()
}

func (r *OrderRepository) SaveOrder(order model.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback()

	deliveryJSON, _ := json.Marshal(order.Delivery)
	paymentJSON, _ := json.Marshal(order.Payment)

	orderQuery := `INSERT INTO orders (order_uid, track_number, entry, delivery, payment, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err = tx.Exec(orderQuery, order.OrderUID, order.TrackNumber, order.Entry, deliveryJSON, paymentJSON, order.Locale, order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("не удалось вставить заказ: %w", err)
	}

	itemQuery := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, item := range order.Items {
		_, err = tx.Exec(itemQuery, order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("не удалось вставить товар (chrt_id: %d): %w", item.ChrtID, err)
		}
	}

	return tx.Commit()
}

// GetOrderByUID получает один заказ по его ID
func (r *OrderRepository) GetOrderByUID(uid string) (*model.Order, error) {
    var o model.Order
    var deliveryJSON, paymentJSON []byte

    row := r.db.QueryRow("SELECT order_uid, track_number, entry, delivery, payment, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders WHERE order_uid = $1", uid)
    if err := row.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &deliveryJSON, &paymentJSON, &o.Locale, &o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // Это не ошибка, просто заказ не найден
        }
        return nil, fmt.Errorf("ошибка сканирования заказа: %w", err)
    }

    json.Unmarshal(deliveryJSON, &o.Delivery)
    json.Unmarshal(paymentJSON, &o.Payment)

    items, err := r.getItemsByOrderUID(uid)
    if err != nil {
        return nil, err
    }
    
    // ИСПРАВЛЕНИЕ: Гарантируем, что если товаров нет, будет пустой массив [], а не null
    if items == nil {
        o.Items = []model.Item{}
    } else {
        o.Items = items
    }
    
    return &o, nil
}


func (r *OrderRepository) GetAllOrders() ([]model.Order, error) {
	rows, err := r.db.Query("SELECT order_uid, track_number, entry, delivery, payment, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders")
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заказов из БД: %w", err)
	}
	defer rows.Close()

	var orders []model.Order
	orderMap := make(map[string]*model.Order)

	for rows.Next() {
		var o model.Order
		// ИСПРАВЛЕНИЕ: Сразу инициализируем срез, чтобы избежать nil
		o.Items = []model.Item{}
		var deliveryJSON, paymentJSON []byte

		if err := rows.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &deliveryJSON, &paymentJSON, &o.Locale, &o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки заказа: %w", err)
		}

		json.Unmarshal(deliveryJSON, &o.Delivery)
		json.Unmarshal(paymentJSON, &o.Payment)
		
		orders = append(orders, o)
		orderMap[o.OrderUID] = &orders[len(orders)-1]
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по заказам: %w", err)
	}

	itemRows, err := r.db.Query("SELECT order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items")
	if err != nil {
		return nil, fmt.Errorf("ошибка получения всех товаров: %w", err)
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var i model.Item
		var orderUID string
		if err := itemRows.Scan(&orderUID, &i.ChrtID, &i.TrackNumber, &i.Price, &i.Rid, &i.Name, &i.Sale, &i.Size, &i.TotalPrice, &i.NmID, &i.Brand, &i.Status); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки товара: %w", err)
		}
		if order, ok := orderMap[orderUID]; ok {
			order.Items = append(order.Items, i)
		}
	}
	if err = itemRows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка после итерации по товарам: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) getItemsByOrderUID(uid string) ([]model.Item, error) {
    rows, err := r.db.Query("SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items WHERE order_uid = $1", uid)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    // ИСПРАВЛЕНИЕ: Инициализируем как пустой срез, а не nil
    items := []model.Item{}
    for rows.Next() {
        var i model.Item
        if err := rows.Scan(&i.ChrtID, &i.TrackNumber, &i.Price, &i.Rid, &i.Name, &i.Sale, &i.Size, &i.TotalPrice, &i.NmID, &i.Brand, &i.Status); err != nil {
            return nil, err
        }
        items = append(items, i)
    }
    return items, nil
}

// package repository

// import (
// 	"database/sql"
// 	"encoding/json" // <-- Добавлен импорт для работы с JSON
// 	"fmt"
// 	"wb-order-service/internal/model" // <-- Добавлен импорт ваших моделей

// 	_ "github.com/lib/pq" // Драйвер для PostgreSQL
// )

// type OrderRepository struct {
// 	db *sql.DB
// }

// // NewOrderRepository создает новый экземпляр репозитория
// func NewOrderRepository(dataSourceName string) (*OrderRepository, error) {
// 	db, err := sql.Open("postgres", dataSourceName)
// 	if err != nil {
// 		return nil, fmt.Errorf("не удалось открыть соединение с БД: %w", err)
// 	}

// 	if err := db.Ping(); err != nil {
// 		db.Close()
// 		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
// 	}

// 	return &OrderRepository{db: db}, nil
// }

// // Close закрывает соединение с БД
// func (r *OrderRepository) Close() {
// 	r.db.Close()
// }

// // SaveOrder сохраняет новый заказ и его товары в БД в рамках одной транзакции
// func (r *OrderRepository) SaveOrder(order model.Order) error {
// 	tx, err := r.db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("не удалось начать транзакцию: %w", err)
// 	}
// 	defer tx.Rollback()

// 	deliveryJSON, err := json.Marshal(order.Delivery)
// 	if err != nil {
// 		return fmt.Errorf("ошибка маршалинга delivery: %w", err)
// 	}
// 	paymentJSON, err := json.Marshal(order.Payment)
// 	if err != nil {
// 		return fmt.Errorf("ошибка маршалинга payment: %w", err)
// 	}

// 	orderQuery := `
// 		INSERT INTO orders (order_uid, track_number, entry, delivery, payment, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
// 	_, err = tx.Exec(orderQuery,
// 		order.OrderUID, order.TrackNumber, order.Entry, deliveryJSON, paymentJSON, order.Locale, order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard)
// 	if err != nil {
// 		return fmt.Errorf("не удалось вставить заказ: %w", err)
// 	}

// 	itemQuery := `
// 		INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
// 	for _, item := range order.Items {
// 		_, err = tx.Exec(itemQuery,
// 			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
// 		if err != nil {
// 			return fmt.Errorf("не удалось вставить товар (chrt_id: %d): %w", item.ChrtID, err)
// 		}
// 	}

// 	return tx.Commit()
// }

// // GetAllOrders загружает все заказы из базы данных для восстановления кэша
// func (r *OrderRepository) GetAllOrders() ([]model.Order, error) {
// 	rows, err := r.db.Query("SELECT order_uid, track_number, entry, delivery, payment, locale, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard FROM orders")
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка получения заказов из БД: %w", err)
// 	}
// 	defer rows.Close()

// 	var orders []model.Order
// 	orderMap := make(map[string]*model.Order) // Используем мапу для быстрого доступа

// 	for rows.Next() {
// 		var o model.Order
// 		var deliveryJSON, paymentJSON []byte

// 		if err := rows.Scan(&o.OrderUID, &o.TrackNumber, &o.Entry, &deliveryJSON, &paymentJSON, &o.Locale, &o.CustomerID, &o.DeliveryService, &o.ShardKey, &o.SmID, &o.DateCreated, &o.OofShard); err != nil {
// 			return nil, fmt.Errorf("ошибка сканирования строки заказа: %w", err)
// 		}

// 		if err := json.Unmarshal(deliveryJSON, &o.Delivery); err != nil {
// 			return nil, fmt.Errorf("ошибка анмаршалинга delivery: %w", err)
// 		}
// 		if err := json.Unmarshal(paymentJSON, &o.Payment); err != nil {
// 			return nil, fmt.Errorf("ошибка анмаршалинга payment: %w", err)
// 		}
// 		orders = append(orders, o)
// 		orderMap[o.OrderUID] = &orders[len(orders)-1]
// 	}
//     if err = rows.Err(); err != nil {
//         return nil, fmt.Errorf("ошибка после итерации по заказам: %w", err)
//     }

// 	// Теперь одним запросом получаем все товары
// 	itemRows, err := r.db.Query("SELECT order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status FROM items")
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка получения всех товаров: %w", err)
// 	}
// 	defer itemRows.Close()

// 	for itemRows.Next() {
// 		var i model.Item
// 		var orderUID string
// 		if err := itemRows.Scan(&orderUID, &i.ChrtID, &i.TrackNumber, &i.Price, &i.Rid, &i.Name, &i.Sale, &i.Size, &i.TotalPrice, &i.NmID, &i.Brand, &i.Status); err != nil {
// 			return nil, fmt.Errorf("ошибка сканирования строки товара: %w", err)
// 		}

// 		// Находим нужный заказ в мапе и добавляем к нему товар
// 		if order, ok := orderMap[orderUID]; ok {
// 			order.Items = append(order.Items, i)
// 		}
// 	}
//     if err = itemRows.Err(); err != nil {
//         return nil, fmt.Errorf("ошибка после итерации по товарам: %w", err)
//     }

// 	return orders, nil
// }