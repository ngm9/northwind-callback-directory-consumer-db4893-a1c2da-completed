package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"northwind/callbackdirectory/service"
)

type Handler struct {
	svc *service.DirectoryService
}

func NewHandler(svc *service.DirectoryService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) PhoneNumbers(w http.ResponseWriter, r *http.Request) {
	country := r.URL.Query().Get("country")
	phone := r.URL.Query().Get("phone")
	if strings.TrimSpace(country) == "" {
		log.Printf("validation failed for %s %s: country is missing or blank", r.Method, r.URL.Path)
		writeError(w, http.StatusBadRequest, "'country' is required and cannot be blank.")
		return
	}
	if strings.TrimSpace(phone) == "" {
		log.Printf("validation failed for %s %s: phone is missing or blank", r.Method, r.URL.Path)
		writeError(w, http.StatusBadRequest, "'phone' is required and cannot be blank.")
		return
	}

	resolved, err := h.svc.GetPhoneNumbers(country, phone)
	if err != nil {
		log.Printf("failed to resolve callback number for country=%q phone=%q: %v", country, phone, err)
		writeError(w, http.StatusInternalServerError, "failed to resolve callback number.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"phone": resolved})
}

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/phone-numbers", h.PhoneNumbers)
	return mux
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		http.Error(w, "failed to serialize response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(buf.Bytes()); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
