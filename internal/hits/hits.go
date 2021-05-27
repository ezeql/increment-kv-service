package hits

import (
	"github.com/ezeql/appcues-increment-simple/internal/hits/handlers"
	"github.com/ezeql/appcues-increment-simple/internal/incrementsrv"
	"github.com/go-redis/redis/v7"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type HitsServer struct {
	Router    *chi.Mux
	Config    Config
	increment *incrementsrv.Publisher
}

type Config struct {
	RedisClient *redis.Client
}

func HitsHTTP(config Config) (*HitsServer, error) {
	h := &HitsServer{}
	r := chi.NewRouter()
	srv, err := incrementsrv.NewPublisher(config.RedisClient)
	if err != nil {
		return nil, err
	}

	h.Router = r
	h.increment = srv
	h.Config = config

	h.initMiddleware()

	hh := handlers.IncrementResource{Service: srv}

	r.Post("/increment", hh.CreateIncrement)
	return h, nil
}
func (h *HitsServer) initMiddleware() {
	h.Router.Use(middleware.RequestID)
	h.Router.Use(middleware.RealIP)
	h.Router.Use(middleware.Recoverer)
}
