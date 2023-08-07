package delete

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/sl_extension"
	"url-shortener/internal/storage"
)

type Request struct {
	Alias string `json:"alias"`
	User  string `json:"user"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias"`
}

type URLDeleter interface {
	DeleteAlias(alias, username string) error
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.delete.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("request body is empty"))

			return
		}
		if err != nil {
			log.Error("fail while decoding request body", sl_extension.Error(err))

			render.JSON(w, r, resp.Error("fail while decoding request body"))

			return
		}

		log.Info("request body was decoded", slog.Any("req", req))

		alias := req.Alias
		if alias == "" {
			log.Error("alias is empty")

			render.JSON(w, r, resp.Error("alias is empty"))

			return
		}

		user := req.User
		if user == "" {
			log.Error("user is empty")

			render.JSON(w, r, resp.Error("user is empty"))

			return
		}

		err = deleter.DeleteAlias(alias, user)
		if errors.Is(err, storage.ErrWrongUser) {
			log.Error("wrong user")

			render.JSON(w, r, resp.Error("wrong user"))

			return
		}
		if err != nil {
			log.Error("failed to delete alias", sl_extension.Error(err))

			render.JSON(w, r, resp.Error("failed to delete alias"))

			return
		}

		log.Info("alias successfully deleted", slog.String("alias", alias))

		responseOK(w, r, alias)
	}
}
