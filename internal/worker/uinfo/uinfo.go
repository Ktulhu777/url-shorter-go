package uinfo

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"url-shorter/internal/lib/logger/sl"

	"github.com/avct/uasurfer"
)

var LogQueue = make(chan LogData, 100)

type LogData struct {
	UA  string
	R   *http.Request
	Log *slog.Logger
}

type ParsedUserInfo struct {
	Timestamp      string
	IP             string
	OSName         string
	Browser        string
	BrowserVersion string
	Device         string
	Platform       string
}

func init() {
	go worker()
}

func worker() {
	for data := range LogQueue {
		writeLog(data.UA, data.R, data.Log)
	}
}

func getIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	return ip
}

func writeLog(ua string, r *http.Request, log *slog.Logger) {
	const fn = "middleware.uinfo.writeLog"
	log = log.With(slog.String("fn", fn))

	// Парсим данные пользователя
	userInfo := parseUserInfo(ua, r)

	// Запись данных в файл
	file, err := os.OpenFile("user_info.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Error("Ошибка при открытии файла: %v", sl.Err(err)) // Убедись, что правильно обрабатываешь ошибки
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(fmt.Sprintf(
		"Timestamp: %s;\nip: %s;\nOS: %s;\nBrowser: %s (%s);\nDevice: %s;\nPlatform: %s\n\n",
		userInfo.Timestamp, userInfo.IP, userInfo.OSName, userInfo.Browser, userInfo.BrowserVersion, userInfo.Device, userInfo.Platform,
	))
	if err != nil {
		log.Error("Ошибка при записи данных: %v", sl.Err(err))
		return
	}

	writer.Flush()
}

func parseUserInfo(ua string, r *http.Request) *ParsedUserInfo {
	userAgent := uasurfer.Parse(ua)
	return &ParsedUserInfo{
		Timestamp:      time.Now().Format(time.RFC3339),
		Browser:        userAgent.Browser.Name.String(),
		BrowserVersion: fmt.Sprintf("%d.%d.%d", userAgent.Browser.Version.Major, userAgent.Browser.Version.Minor, userAgent.Browser.Version.Patch),
		Device:         userAgent.DeviceType.String(),
		OSName:         userAgent.OS.Name.String(),
		Platform:       userAgent.OS.Platform.String(),
		IP:             getIP(r),
	}
}
