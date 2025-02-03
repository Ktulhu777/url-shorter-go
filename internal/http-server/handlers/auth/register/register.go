package register

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	resp "url-shorter/internal/lib/api/response"
	"url-shorter/internal/lib/logger/sl"
	"url-shorter/internal/storage"
)

type Request struct {
	Username   string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
	PasswordRe string `json:"password_re" validate:"required,eqfield=Password"`
}

type Response struct {
	resp.Response
	ID int64 `json:"id"`
}

type UserSaver interface {
	SaveUser(username, email, password string) (int64, error)
}

func New(log *slog.Logger, userSaver UserSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.auth.register.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			log.Error("invalid request", sl.Err(err))
			validatorErr := err.(validator.ValidationErrors)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationErrorRegisterUser(validatorErr))
			return
		}

		id, err := userSaver.SaveUser(req.Username, req.Email, req.Password)

		if errors.Is(err, storage.ErrUsernamelExists) {
			log.Info("username already exists", slog.String("username", req.Username))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("username already exists"))
			return
		}

		if errors.Is(err, storage.ErrEmailExists) {
			log.Info("email already exists", slog.String("email", req.Email))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("email already exists"))
			return
		}

		if err != nil {
			log.Error("failed to add user", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to add user"))
			return
		}

		log.Info("user added", slog.Int64("id", id))

		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, Response{
			Response: resp.OK(),
			ID:       id,
		})
	}
}
