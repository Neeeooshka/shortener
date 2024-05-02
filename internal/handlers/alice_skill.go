package handlers

import (
	"encoding/json"
	"github.com/Neeeooshka/alice-skill.git/internal/logger"
	"github.com/Neeeooshka/alice-skill.git/internal/models"
	"go.uber.org/zap"
	"net/http"
)

func AliceSkill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Log.Debug("got request with bad method", zap.String("method", r.Method))
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// десериализуем запрос в структуру модели
	logger.Log.Debug("decoding request")
	var req models.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// проверяем, что пришёл запрос понятного типа
	if req.Request.Type != models.TypeSimpleUtterance {
		logger.Log.Debug("unsupported request type", zap.String("type", req.Request.Type))
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	text := "Для вас нет новых сообщений."

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
			Text: "Извините, я пока ничего не умею",
		},
		Version: "1.0",
	}

	w.Header().Set("Content-Type", "application/json")

	// сериализуем ответ сервера
	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}
	logger.Log.Debug("sending HTTP 200 response")
}
