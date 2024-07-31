package save

import (
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "github.com/marcheneli/forms/internal/lib/api/response"
	"github.com/marcheneli/forms/internal/lib/logger/sl"

	"github.com/marcheneli/forms/internal/storage/sqlite/fields"
)

type Response struct {
	resp.Response
	Fields []fields.Field
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type FieldsLister interface {
	GetList() ([]fields.Field, error)
}

func New(log *slog.Logger, fieldsLister FieldsLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.fields.list.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		fields, err := fieldsLister.GetList()
		if err != nil {
			log.Error("failed to list fields", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to list fields"))

			return
		}

		log.Info("fields listed", slog.Int("fields length", len(fields)))

		responseOK(w, r, fields)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, fields []fields.Field) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Fields:   fields,
	})
}
