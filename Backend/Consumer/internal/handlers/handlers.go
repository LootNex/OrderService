package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/LootNex/OrderService/Consumer/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	Serv service.ServiceManager
	log  *zap.Logger
}

func NewHandler(serv service.ServiceManager, logg *zap.Logger) *Handler {
	return &Handler{
		Serv: serv,
		log:  logg,
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h Handler) GetOrder(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	orderID := vars["id"]

	ctx := r.Context()

	order, err := h.Serv.GetOrderByID(ctx, orderID)
	if err != nil {
		h.log.Warn("order not found", zap.String("order_id", orderID), zap.Error(err))
		http.Error(w, "no such orderID", http.StatusBadRequest)
		return
	}

	resp, err := json.MarshalIndent(order, "", "   ")
	if err != nil {
		http.Error(w, "problems with server, try again later", http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		h.log.Error("failed to write response", zap.Error(err))
	}

}
