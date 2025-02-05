package uinfo

import (
	"log/slog"
	"net/http"

	workerUInfo "url-shorter/internal/worker/uinfo"

	"github.com/go-chi/chi/v5/middleware"
)

func GetUserInfo(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const fn = "middleware.uinfo.GetUserInfo"

			log = log.With(
				slog.String("fn", fn),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			select {
			case workerUInfo.LogQueue <- workerUInfo.LogData{
				UA:  r.UserAgent(),
				R:   r,
				Log: log,
			}:
			default:
				log.Warn("Очередь логов заполнена, пропускаем запись")
			}

			next.ServeHTTP(w, r)
		})
	}
}
