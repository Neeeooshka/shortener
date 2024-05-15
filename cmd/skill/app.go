package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Neeeooshka/alice-skill.git/pkg/logger/zap"
	"net/http"
	"strings"
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

	// текст ответа навыка
	var text string

	switch true {
	// пользователь попросил отправить сообщение
	case strings.HasPrefix(req.Request.Command, "Отправь"):
		// гипотетическая функция parseSendCommand вычленит из запроса логин адресата и текст сообщения
		username, message := parseSendCommand(req.Request.Command)

		// найдём внутренний идентификатор адресата по его логину
		recipientID, err := a.store.FindRecipient(ctx, username)
		if err != nil {
			log.Debug("cannot find recipient by username", log.String("username", username), log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// сохраняем новое сообщение в СУБД, после успешного сохранения оно станет доступно для прослушивания получателем
		err = a.store.SaveMessage(ctx, recipientID, store.Message{
			Sender:  req.Session.User.UserID,
			Time:    time.Now(),
			Payload: message,
		})
		if err != nil {
			log.Debug("cannot save message", log.String("recipient", recipientID), log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Оповестим отправителя об успешности операции
		text = "Сообщение успешно отправлено"

	// пользователь попросил прочитать сообщение
	case strings.HasPrefix(req.Request.Command, "Прочитай"):
		// гипотетическая функция parseReadCommand вычленит из запроса порядковый номер сообщения в списке доступных
		messageIndex := parseReadCommand(req.Request.Command)

		// получим список непрослушанных сообщений пользователя
		messages, err := a.store.ListMessages(ctx, req.Session.User.UserID)
		if err != nil {
			log.Debug("cannot load messages for user", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		text = "Для вас нет новых сообщений."
		if len(messages) < messageIndex {
			// пользователь попросил прочитать сообщение, которого нет
			text = "Такого сообщения не существует."
		} else {
			// получим сообщение по идентификатору
			messageID := messages[messageIndex].ID
			message, err := a.store.GetMessage(ctx, messageID)
			if err != nil {
				log.Debug("cannot load message", log.Int64("id", messageID), log.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// передадим текст сообщения в ответе
			text = fmt.Sprintf("Сообщение от %s, отправлено %s: %s", message.Sender, message.Time, message.Payload)
		}

	// пользователь хочет зарегистрироваться
	case strings.HasPrefix(req.Request.Command, "Зарегистрируй"):
		// гипотетическая функция parseRegisterCommand вычленит из запроса
		// желаемое имя нового пользователя
		username := parseRegisterCommand(req.Request.Command)

		// регистрируем пользователя
		err := a.store.RegisterUser(ctx, req.Session.User.UserID, username)
		// наличие неспецифичной ошибки
		if err != nil && !errors.Is(err, store.ErrConflict) {
			log.Debug("cannot register user", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// определяем правильное ответное сообщение пользователю
		text = fmt.Sprintf("Вы успешно зарегистрированы под именем %s", username)
		if errors.Is(err, store.ErrConflict) {
			// ошибка специфична для случая конфликта имён пользователей
			text = "Извините, такое имя уже занято. Попробуйте другое."
		}
	// если не поняли команду, просто скажем пользователю, сколько у него новых сообщений
	default:
		messages, err := a.store.ListMessages(ctx, req.Session.User.UserID)
		if err != nil {
			log.Debug("cannot load messages for user", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		text = "Для вас нет новых сообщений."
		if len(messages) > 0 {
			text = fmt.Sprintf("Для вас %d новых сообщений.", len(messages))
		}

		// первый запрос новой сессии
		if req.Session.New {
			// обработаем поле Timezone запроса
			tz, err := time.LoadLocation(req.Timezone)
			if err != nil {
				log.Debug("cannot parse timezone")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// получим текущее время в часовом поясе пользователя
			now := time.Now().In(tz)
			hour, minute, _ := now.Clock()

			// формируем новый текст приветствия
			text = fmt.Sprintf("Точное время %d часов, %d минут. %s", hour, minute, text)
		}
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

func parseReadCommand(command string) int {
	return 1
}

func parseSendCommand(command string) (string, string) {
	return "alice", "test message"
}

func parseRegisterCommand(command string) string {
	return "Дима"
}
