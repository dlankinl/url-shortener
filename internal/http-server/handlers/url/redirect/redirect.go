package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/sl_extension"
	"url-shortener/internal/storage"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, getter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.redirect.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("not found"))

			return
		}

		resURL, err := getter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.String("alias", alias))

			render.JSON(w, r, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get url", sl_extension.Error(err))

			render.JSON(w, r, resp.Error("internal error: failed to get url"))

			return
		}

		log.Info("url was got", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
