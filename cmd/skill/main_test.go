package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Neeeooshka/alice-skill.git/internal/store"
	"github.com/Neeeooshka/alice-skill.git/internal/store/mock"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor"
	gz "github.com/Neeeooshka/alice-skill.git/pkg/compressor/gzip"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhook(t *testing.T) {
	// создадим конроллер моков и экземпляр мок-хранилища
	ctrl := gomock.NewController(t)
	s := mock.NewMockStore(ctrl)

	// определим, какой результат будем получать от «хранилища»
	messages := []store.Message{
		{
			Sender:  "411419e5-f5be-4cdb-83aa-2ca2b6648353",
			Time:    time.Now(),
			Payload: "Hello!",
		},
	}

	// установим условие: при любом вызове метода ListMessages возвращать массив messages без ошибки
	s.EXPECT().
		ListMessages(gomock.Any(), gomock.Any()).
		Return(messages, nil)

	// создадим экземпляр приложения и передадим ему «хранилище»
	appInstance := newApp(s)

	handler := http.HandlerFunc(appInstance.AliceSkill)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	testCases := []struct {
		name         string // добавляем название тестов
		method       string
		body         string // добавляем тело запроса в табличные тесты
		expectedCode int
		expectedBody string
	}{
		{
			name:         "method_get",
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_put",
			method:       http.MethodPut,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_delete",
			method:       http.MethodDelete,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_post_without_body",
			method:       http.MethodPost,
			expectedCode: http.StatusInternalServerError,
			expectedBody: "",
		},
		{
			name:         "method_post_unsupported_type",
			method:       http.MethodPost,
			body:         `{"request": {"type": "idunno", "command": "do something"}, "version": "1.0"}`,
			expectedCode: http.StatusUnprocessableEntity,
			expectedBody: "",
		},
		{
			name:         "method_post_success",
			method:       http.MethodPost,
			body:         `{"request": {"type": "SimpleUtterance", "command": "sudo do something"}, "session": {"new": true}, "version": "1.0"}`,
			expectedCode: http.StatusOK,
			expectedBody: `Точное время .* часов, .* минут. Для вас 1 новых сообщений.`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL

			if len(tc.body) > 0 {
				req.SetHeader("Content-Type", "application/json")
				req.SetBody(tc.body)
			}

			resp, err := req.Send()
			assert.NoError(t, err, "error making HTTP request")

			assert.Equal(t, tc.expectedCode, resp.StatusCode(), "Response code didn't match expected")
			// проверяем корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				// сравниваем тело ответа с ожидаемым шаблоном
				assert.Regexp(t, tc.expectedBody, string(resp.Body()))
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	// создадим конроллер моков и экземпляр мок-хранилища
	ctrl := gomock.NewController(t)
	s := mock.NewMockStore(ctrl)

	// определим, какой результат будем получать от «хранилища»
	messages := []store.Message{
		{
			Sender:  "411419e5-f5be-4cdb-83aa-2ca2b6648353",
			Time:    time.Now(),
			Payload: "Hello!",
		},
	}

	// установим условие: при любом вызове метода ListMessages возвращать массив messages без ошибки
	s.EXPECT().
		ListMessages(gomock.Any(), gomock.Any()).
		Return(messages, nil).
		MaxTimes(2)

	// создадим экземпляр приложения и передадим ему «хранилище»
	appInstance := newApp(s)

	handler := http.HandlerFunc(compressor.IncludeCompressor(appInstance.AliceSkill, gz.NewGzipCompressor()))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	requestBody := `{"request": {"type": "SimpleUtterance", "command": "sudo do something"}, "session": {"new": true}, "version": "1.0"}`

	// ожидаемое содержимое тела ответа при успешном запросе
	successBody := `Точное время .* часов, .* минут. Для вас 1 новых сообщений.`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Regexp(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		assert.Regexp(t, successBody, string(b))
	})
}
