package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "url-shorter/internal/lib/api/response"
	"url-shorter/internal/lib/logger/sl"
	"url-shorter/internal/storage"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        const fn = "handlers.redirect.New"

        log := log.With(
            slog.String("fn", fn),
            slog.String("request_id", middleware.GetReqID(r.Context())),
            slog.String("method", r.Method),
            slog.String("url", r.URL.String()),
        )

        log.Info("New handler called", slog.String("method", r.Method), slog.String("url", r.URL.String()))

        alias := chi.URLParam(r, "alias")
        if alias == "" {
            log.Info("alias is empty")
            render.JSON(w, r, resp.Error("invalid request"))
            return
        }

        resURL, err := urlGetter.GetURL(alias)
        if errors.Is(err, storage.ErrURLNotFound) {
            log.Info("url not found", slog.String("alias", alias))
            render.JSON(w, r, resp.Error("not found"))
            return
        }

        if err != nil {
            log.Error("failed to get url", sl.Err(err))
            render.JSON(w, r, resp.Error("internal error"))
            return
        }

        log.Info("got url", slog.String("url", resURL))

        // redirect to found url
        http.Redirect(w, r, resURL, http.StatusFound)
    }
}

