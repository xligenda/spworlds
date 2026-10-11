package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/xligenda/spworlds"
)

type app struct {
	spw      *spworlds.Client
	redirect string
	webhook  string

	mu           sync.Mutex
	lastDonation *spworlds.PaymentData
}

func newApp(c *spworlds.Client, redirect string, webhook string) *app {
	return &app{
		spw:      c,
		redirect: redirect,
		webhook:  webhook,
	}
}

func (a *app) donateWebhook(w http.ResponseWriter, r *http.Request) {
	if passed, _ := a.spw.ValidateRequest(r); !passed {
		http.Error(w, "Проверка не пройдена", http.StatusForbidden)
	}

	data, err := a.spw.ParsePaymentData(r)
	if err != nil {
		http.Error(w, "Ошибка парсинга", http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	a.lastDonation = data
	a.mu.Unlock()

	log.Printf("Успешный платеж! ReqID (Payload): %s", data.Payload)
	w.WriteHeader(http.StatusOK)
}

func (a *app) latestDonate(w http.ResponseWriter, r *http.Request) {
	a.mu.Lock()
	// copy from ref
	donate := a.lastDonation
	a.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if donate == nil {
		w.Write([]byte(`{}`))
		return
	}

	if err := json.NewEncoder(w).Encode(donate); err != nil {
		log.Printf("Ошибка отправки JSON: %v", err)
		http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func (a *app) donateLink(w http.ResponseWriter, r *http.Request) {
	amountStr := r.PathValue("amount")

	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 || amount > 9999 {
		http.Error(w, "Некорректная сумма доната", http.StatusBadRequest)
		return
	}

	reqID := generateRequestID()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	p, err := a.spw.CreatePayment(ctx, spworlds.NewCreatePaymentOptions(
		[]spworlds.Product{
			{
				Name:  "Донат!",
				Count: 1,
				Price: amount,
			},
		}, a.redirect, a.webhook,
	).SetPayload(reqID))

	if err != nil {
		log.Printf("Ошибка создания платежа (ReqID: %s): %v", reqID, err)
		http.Error(w, "Ошибка при создании платежа", http.StatusInternalServerError)
		return
	}

	log.Printf("Платеж успешно создан (ReqID: %s, Сумма: %d)", reqID, amount)
	http.Redirect(w, r, p.URL, http.StatusFound)
}
