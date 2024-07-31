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
	Page     int `json:"page" validate:"required"`
	PageSize int `json:"pageSize" validate:"required"`
}

type Response struct {
	resp.Response
	Fields []Field
}

type Field struct {
	Name string
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type SchemaLister interface {
	List() ([]Field, error)
}

func New(log *slog.Logger, schemaLister SchemaLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.fields.list.New"

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

		schemas, err := schemaLister.List()
		if err != nil {
			log.Error("failed to list fields", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to list fields"))

			return
		}

		log.Info("fields listed", slog.Int("fields length", len(schemas)))

		responseOK(w, r, schemas)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, fields []Field) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Fields:   fields,
	})
}
