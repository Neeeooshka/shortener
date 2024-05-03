package main

import (
	"encoding/json"
	"fmt"
	"github.com/Neeeooshka/alice-skill.git/pkg/logger/zap"
	"net/http"
	"time"

	"github.com/Neeeooshka/alice-skill.git/internal/models"
	"github.com/Neeeooshka/alice-skill.git/internal/store"
)

// app инкапсулирует в себя все зависимости и логику приложения
type app struct {
	store store.Store
}

// newApp принимает на вход внешние зависимости приложения и возвращает новый объект app
func newApp(s store.Store) *app {
	return &app{store: s}
}

func (a *app) AliceSkill(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := zap.Log

	if r.Method != http.MethodPost {
		log.Debug("got request with bad method", log.String("method", r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// десериализуем запрос в структуру модели
	log.Debug("decoding request")
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		log.Debug("cannot decode request JSON body", log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// проверяем, что пришёл запрос понятного типа
	if req.Request.Type != models.TypeSimpleUtterance {
		log.Debug("unsupported request type", log.String("type", req.Request.Type))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// получаем список сообщений для текущего пользователя
	messages, err := a.store.ListMessages(ctx, req.Session.User.UserID)
	if err != nil {
		log.Debug("cannot load messages for user", log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	text := "Для вас нет новых сообщений."
	if len(messages) > 0 {
		text = fmt.Sprintf("Для вас %d новых сообщений.", len(messages))
	}

	// первый запрос новой сессии
	if req.Session.New {
		// обрабатываем поле Timezone запроса
		tz, err := time.LoadLocation(req.Timezone)
		if err != nil {
			log.Debug("cannot parse timezone")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// получаем текущее время в часовом поясе пользователя
		now := time.Now().In(tz)
		hour, minute, _ := now.Clock()

		// формируем текст ответа
		text = fmt.Sprintf("Точное время %d часов, %d минут. %s", hour, minute, text)
	}

	// заполняем модель ответа
	resp := models.Response{
		Response: models.ResponsePayload{
			Text: text, // Алиса проговорит новый текст
		},
		Version: "1.0",
	}

	w.Header().Set("Content-Type", "application/json")

	// сериализуем ответ сервера
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		log.Debug("error encoding response", log.Error(err))
		return
	}
	log.Debug("sending HTTP 200 response")
}
