package handler

import (
	"encoding/json"
	"net/http"
	"wb-order-service/internal/service"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

// GetOrderByUID обрабатывает GET-запрос для получения заказа
func (h *OrderHandler) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	// Получаем order_uid из URL
	vars := mux.Vars(r)
	uid, ok := vars["order_uid"]
	if !ok {
		http.Error(w, "order_uid не найден в запросе", http.StatusBadRequest)
		return
	}

	// Ищем заказ в сервисе (который посмотрит в кэше)
	order, found := h.service.GetOrderByUID(uid)
	if !found {
		http.Error(w, "Заказ не найден", http.StatusNotFound)
		return
	}

	// Отдаем результат в виде JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}