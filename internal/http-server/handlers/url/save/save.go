package save

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"
	"io"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/lib/sl_extension"
	"url-shortener/internal/storage"
)

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
	User  string `json:"user"`
}

type URLSaver interface {
	SaveURL(alias, urlToSave, user string) (int64, error)
}

const aliasLength = 10

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.save.New"

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

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl_extension.Error(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias

		if alias == "" {
			alias = random.GenerateRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(alias, req.URL, req.User)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}

		if err != nil {
			log.Error("failed to save url", sl_extension.Error(err))

			render.JSON(w, r, resp.Error("failed to save url"))

			return
		}

		log.Info("url successfully saved", slog.Int64("id", id))

		responseOK(w, r, alias)
	}
}
