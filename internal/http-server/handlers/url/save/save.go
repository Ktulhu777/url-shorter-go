package save

import (
	"errors"
	"log/slog"
	"net/http"

	resp "url-shorter/internal/lib/api/response"
	"url-shorter/internal/lib/logger/sl"
	"url-shorter/internal/lib/random"
	"url-shorter/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error) 
	IsAliasExists(alias string) (bool, error)
}

type Request struct {
	URL string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 8

func New(log *slog.Logger, URLSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.save.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w,r, resp.Error("failed to decode request"))

			return
		}
		
		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))
			validatorErr := err.(validator.ValidationErrors)
			render.JSON(w,r, resp.ValidationError(validatorErr))
			return
		}
		
		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}
		if exists, err := URLSaver.IsAliasExists(alias); err != nil {
			log.Error("failed to check alias uniqueness", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to check alias uniqueness"))
			return
		} else if exists {
			log.Info("alias already exists", slog.String("alias", alias))
			render.JSON(w, r, resp.Error("alias already exists"))
			return
		}

		id, err := URLSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			render.JSON(w, r, resp.Error("url already exists"))
			return
		}

		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to add url"))
			return			
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias: alias,
		})
	}
}