package handlers

import "net/http"

func NewHandler() *Handler {
	return &Handler{}
}

type Handler struct {
}

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {

}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {

}
