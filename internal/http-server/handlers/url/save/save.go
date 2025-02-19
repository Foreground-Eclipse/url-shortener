package save

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/foreground-eclipse/url-shortener/internal/lib/api/response"
	"github.com/foreground-eclipse/url-shortener/internal/lib/logger/sl"
	"github.com/foreground-eclipse/url-shortener/internal/lib/random"
	"github.com/foreground-eclipse/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

//go:generate go run github.com/vektra/mockery/v2 --name=URLSaver

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
	CheckIfAliasExists(alias string) (bool, error)
}

// TODO: move to config or db or whatever
const aliasLength = 5

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return // return is needed because render.JSON doesnt stop the handler
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias

		exists, err := urlSaver.CheckIfAliasExists(alias)
		if err != nil {
			log.Error("failed to save url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to save url"))

			return
		}

		if exists {
			log.Info("alias already exists", slog.String("alias", req.Alias))

			render.JSON(w, r, resp.Error("alias already exists"))

			return
		}

		if alias == "" {
			alias = random.NewRandomString(aliasLength)

			exists, err := urlSaver.CheckIfAliasExists(alias)
			if err != nil {
				log.Error("failed to save url", sl.Err(err))

				render.JSON(w, r, resp.Error("failed to save url"))

				return

			}

			if exists {
				alias = random.NewRandomString(aliasLength)
			}

			id, err := urlSaver.SaveURL(req.URL, alias)
			if errors.Is(err, storage.ErrURLExists) {
				log.Info("url already exists", slog.String("url", req.URL))

				render.JSON(w, r, resp.Error("url already exists"))

				return
			}

			if err != nil {

				log.Error("failed to save url", sl.Err(err))

				render.JSON(w, r, resp.Error("failed to save url"))

				return
			}

			log.Info("url added", slog.Int64("id", id))

			responseOK(w, r, alias)
		}
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
