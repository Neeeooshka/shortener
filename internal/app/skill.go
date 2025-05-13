package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Neeeooshka/alice-skill.git/internal/models"
	"github.com/Neeeooshka/alice-skill.git/internal/store"
	"github.com/Neeeooshka/alice-skill.git/pkg/logger/zap"
)

// app инкапсулирует в себя все зависимости и логику приложения
type skillApp struct {
	store store.Store
	// канал для отложенной отправки новых сообщений
	msgChan chan store.Message
}

// newApp принимает на вход внешние зависимости приложения и возвращает новый объект app
func NewSkillApp(s store.Store) *skillApp {
	instance := &skillApp{
		store:   s,
		msgChan: make(chan store.Message, 1024), // установим каналу буфер в 1024 сообщения
	}

	// запустим горутину с фоновым сохранением новых сообщений
	go instance.flushMessages()

	return instance
}

func (a *skillApp) AliceSkill(w http.ResponseWriter, r *http.Request) {
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
		recepientID, err := a.store.FindRecipient(ctx, username)
		if err != nil {
			log.Debug("cannot find recepient by username", log.String("username", username), log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// отправим сообщение в очередь на сохранение
		a.msgChan <- store.Message{
			Sender:    req.Session.User.UserID,
			Recepient: recepientID,
			Time:      time.Now(),
			Payload:   message,
		}

		// оповестим отправителя об успешности операции
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

func parseReadCommand(_ string) int {
	return 1
}

func parseSendCommand(_ string) (string, string) {
	return "alice", "test message"
}

func parseRegisterCommand(_ string) string {
	return "alice"
}

// flushMessages постоянно сохраняет несколько сообщений в хранилище с определённым интервалом
func (a *skillApp) flushMessages() {

	log := zap.Log

	// будем сохранять сообщения, накопленные за последние 10 секунд
	ticker := time.NewTicker(10 * time.Second)

	var messages []store.Message

	for {
		select {
		case msg := <-a.msgChan:
			// добавим сообщение в слайс для последующего сохранения
			messages = append(messages, msg)
		case <-ticker.C:
			// подождём, пока придёт хотя бы одно сообщение
			if len(messages) == 0 {
				continue
			}
			// сохраним все пришедшие сообщения одновременно
			err := a.store.SaveMessages(context.TODO(), messages...)
			if err != nil {
				log.Debug("cannot save messages", log.Error(err))
				// не будем стирать сообщения, попробуем отправить их чуть позже
				continue
			}
			// сотрём успешно отосланные сообщения
			messages = nil
		}
	}
}
