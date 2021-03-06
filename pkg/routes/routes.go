package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ah8ad3/gateway/pkg/auth"
	"github.com/ah8ad3/gateway/pkg/logger"
	"github.com/ah8ad3/gateway/plugins/ratelimitter"
	"github.com/go-chi/chi"
)

func welcome(w http.ResponseWriter, r *http.Request) {
	_ = r
	_, _ = w.Write(getServices())
}

func init() {
	go ratelimitter.CleanupVisitors()
}

// V1 Route function for first method of routing
func V1() *chi.Mux {
	r := chi.NewRouter()
	// Rate limiter per user
	r.Use(ratelimitter.LimitMiddleware)

	r.Get("/", welcome)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.RegisterUser)
		r.Post("/sign", auth.SignJWT)
		r.Post("/check", auth.CheckJwt)
	})

	for _, val := range Services {
		r.Route(val.Path, func(r chi.Router) {
			for _, url := range val.Urls {
				switch url.Method {
				case "GET":
					r.Get(url.Path, func(writer http.ResponseWriter, request *http.Request) {

						// remove path form url and send to service and serve answer
						splitRoute := strings.Split(request.URL.Path, "/")
						route := strings.Join(splitRoute[2:], "/")
						if route == "" {
							route = "/"
						} else {
							route = "/" + route
						}

						logger.SetUserLog(logger.UserLog{Log: logger.Log{Event: "log"}, RequestURL: request.URL.Path,
							IP: request.RemoteAddr, Time: time.Now()})

						writer.Header().Set("Content-Type", "application/json")

						server := findService(splitRoute[1])
						body := GetService(server, route, request.URL.RawQuery)
						_, _ = writer.Write(body)
					})

				case "POST":
					r.Post(url.Path, func(writer http.ResponseWriter, request *http.Request) {
						// remove path form url and send to service and serve answer
						splitRoute := strings.Split(request.URL.Path, "/")[2:]
						route := strings.Join(splitRoute, "/")
						if route == "" {
							route = "/"
						} else {
							route = "/" + route
						}
						_ = request.ParseForm()

						m := make(map[string]interface{})
						for key, value := range request.Form {
							m[key] = strings.Join(value, "")
						}

						data, _ := json.Marshal(m)

						logger.SetUserLog(logger.UserLog{Log: logger.Log{Event: "log"}, RequestURL: request.URL.Path,
							IP: request.RemoteAddr, Time: time.Now()})

						writer.Header().Set("Content-Type", "application/json")
						server := findService(splitRoute[1])
						body := PostService(server, route, data)
						_, _ = writer.Write(body)
					})

				case "PUT":
					// not implemented now
					r.Put(url.Path, func(writer http.ResponseWriter, request *http.Request) {
						_, _ = writer.Write([]byte("hello"))
					})

				case "DELETE":
					// not implemented now
					r.Delete(url.Path, func(writer http.ResponseWriter, request *http.Request) {
						_, _ = writer.Write([]byte("hello"))
					})

				default:
					logger.SetSysLog(logger.SystemLog{Log: logger.Log{Event: "critical",
						Description: fmt.Sprintf("Bad url method in service %s", val.Name)},
						Pkg: "auth", Time: time.Now()})
					//log.Fatal("Bad url method in service ", val.Name)

				}

			}
		})
	}
	return r
}
