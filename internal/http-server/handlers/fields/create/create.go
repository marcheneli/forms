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
	Name     string `json:"name" validate:"required"`
	SchemaId int    `json:"schemaId" validate:"required"`
}

type Response struct {
	resp.Response
	Id int
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type FieldCreator interface {
	Create(name string, schemaId int) (int64, error)
}

func New(log *slog.Logger, fieldCreator FieldCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.fields.create.New"

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

		id, err := fieldCreator.Create(req.Name, req.SchemaId)
		if err != nil {
			log.Error("failed to add field", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to add field"))

			return
		}

		log.Info("field added", slog.Int64("id", id))

		responseOK(w, r, int(id))
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, id int) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Id:       id,
	})
}
