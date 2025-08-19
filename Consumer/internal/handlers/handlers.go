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
