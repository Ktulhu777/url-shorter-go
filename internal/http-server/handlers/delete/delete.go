package delete

import (
	"log/slog"
	"net/http"
	"strconv"
	resp "url-shorter/internal/lib/api/response"
	"url-shorter/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLDeleter interface {
	DeleteURL(id int) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.delete.New"

		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		idStr  := chi.URLParam(r, "id")
		 
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Info("invalid id")
			render.JSON(w, r, resp.Error("invalid request"))

			return
		}
		if id < 0 {
			log.Error("id is negative number")
			render.JSON(w, r, resp.Error("the number cannot be less than zero"))
			return
		}

		err = urlDeleter.DeleteURL(id)

		if err != nil {
			log.Error("deletion not completed", slog.Int64("id", int64(id)), sl.Err(err))
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		log.Info("deletion completed", slog.Int64("id", int64(id)))
		render.JSON(w, r, resp.OK())
	}
}
