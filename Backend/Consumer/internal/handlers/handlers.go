package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/LootNex/OrderService/Consumer/internal/service"
	"github.com/gorilla/mux"
)

type Handler struct {
	Serv service.ServiceManager
}

func NewHandler(serv service.ServiceManager) *Handler {
	return &Handler{
		Serv: serv,
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
		http.Error(w, "no such orderID", http.StatusBadRequest)
	}

	resp, err := json.MarshalIndent(order, "", "   ")
	if err != nil {
		http.Error(w, "problems with server, try again later", http.StatusInternalServerError)
		return
	}

	w.Write(resp)

}
