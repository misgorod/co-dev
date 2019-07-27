package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/misgorod/co-dev/auth"
	"github.com/misgorod/co-dev/db"
	"github.com/misgorod/co-dev/middlewares"
	"github.com/misgorod/co-dev/post"
	"github.com/misgorod/co-dev/users"
	"gopkg.in/go-playground/validator.v9"
)

func main() {
	client, err := db.Connect()
	if err != nil {
		panic(err)
	}
	authHandler := auth.Handler{
		Client:   client,
		Validate: validator.New(),
	}
	postHandler := post.Handler{
		Client:   client,
		Validate: validator.New(),
	}
	usersHandler := users.Handler{
		Client: client,
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Logger, middleware.Recoverer)
	r.Route("/api", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Route("/users", func(r chi.Router) {
			r.Use(middlewares.Authenticate)
			r.Get("/{id}", usersHandler.Get)
		})

		r.Route("/posts", func(r chi.Router) {
			r.Get("/", postHandler.GetAll)
			r.Group(func(r chi.Router) {
				r.Use(middlewares.Authenticate)
				r.Post("/", postHandler.Post)
			})
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", postHandler.Get)
				r.Route("/members", func(r chi.Router) {
					r.Use(middlewares.Authenticate)
					r.Post("/", postHandler.MemberPost)
					r.Delete("/", postHandler.MemberDelete)
				})
				r.Route("/image", func(r chi.Router) {
					r.Use(middlewares.Authenticate)
					r.Post("/", postHandler.PostImage)
				})
			})
		})

		r.Get("/image/{id}", postHandler.GetImage)
	})
	log.Fatal(http.ListenAndServe(":8080", r))
}
