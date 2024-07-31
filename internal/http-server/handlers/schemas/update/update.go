package save

import (
	"errors"
	"io"
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	resp "github.com/marcheneli/forms/internal/lib/api/response"
	"github.com/marcheneli/forms/internal/lib/logger/sl"
)

type Request struct {
	Name string `json:"name" validate:"required"`
	Id   int    `json:"id" validate:"required"`
}

type Response struct {
	resp.Response
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=FieldUpdater
type FieldUpdater interface {
	Update(name string, id int) error
}

func New(log *slog.Logger, fieldUpdater FieldUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.fields.update.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			// Такую ошибку встретим, если получили запрос с пустым телом.
			// Обработаем её отдельно
			log.Error("request body is empty")

			render.JSON(w, r, resp.Error("empty request"))

			return
		}
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		updateErr := fieldUpdater.Update(req.Name, req.Id)
		if err != nil {
			log.Error("failed to update schema", sl.Err(updateErr))

			render.JSON(w, r, resp.Error("failed to update schema"))

			return
		}

		log.Info("schema updated", slog.Int("id", req.Id))

		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
