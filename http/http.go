package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/fahmifan/covidbot"
	"github.com/go-chi/chi/v5"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

const defaultPort = "8000"

type Config struct {
	Port        string
	UserService covidbot.UserService

	router     chi.Router
	httpServer *http.Server
}

type Server struct {
	*Config
}

func NewServer(cfg *Config) *Server {
	if cfg.Port == "" {
		cfg.Port = defaultPort
	}
	s := &Server{cfg}
	s.init()
	return s
}

type respErr struct {
	Error string `json:"error"`
}

func wjson(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(covidbot.Dump(v))
}

// write json error
func werror(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(covidbot.Dump(respErr{Error: msg}))
}

func (s *Server) init() {
	s.router = chi.NewRouter()

	s.router.Get("/ping", ping)
	s.router.Post("/webhooks/telegram", s.handleTelegramWebhook)
}

func (s *Server) Run() error {
	logrus.Info("start http server at :" + s.Port)
	s.httpServer = &http.Server{Addr: ":" + s.Port, Handler: s.router}
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	logrus.Info("stopping http server")
	err := s.httpServer.Shutdown(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logrus.Info("http server stopped without error")
	return nil
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func (s *Server) handleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	update := tgbotapi.Update{}
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		logrus.Error(err)
		werror(w, http.StatusBadRequest, "failed precondition")
		return
	}

	user := &covidbot.User{
		TelegramUpdateID: strconv.Itoa(update.UpdateID),
		ChatID:           update.Message.Chat.ID,
		Username:         update.Message.Chat.UserName,
	}
	if update.UpdateID <= 0 || user.ChatID <= 0 || user.Username == "" {
		werror(w, http.StatusBadRequest, "failed precondition")
		return
	}

	err = s.UserService.Create(r.Context(), user)
	if err != nil {
		logrus.Error(err)
		werror(w, http.StatusInternalServerError, "system error")
		return
	}

	logrus.Info(covidbot.Dumps(user))
	wjson(w, http.StatusOK, user)
}
