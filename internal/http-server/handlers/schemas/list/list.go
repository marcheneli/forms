package save

import (
	"net/http"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	resp "github.com/marcheneli/forms/internal/lib/api/response"
	"github.com/marcheneli/forms/internal/lib/logger/sl"

	"github.com/marcheneli/forms/internal/storage/sqlite/schemas"
)

type Response struct {
	resp.Response
	Schemas []schemas.Schema
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type SchemasLister interface {
	GetList() ([]schemas.Schema, error)
}

func New(log *slog.Logger, schemasLister SchemasLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.schemas.list.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		schemas, err := schemasLister.GetList()
		if err != nil {
			log.Error("failed to list schemas", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to list schemas"))

			return
		}

		log.Info("schemas listed", slog.Int("schemas length", len(schemas)))

		responseOK(w, r, schemas)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, schemas []schemas.Schema) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Schemas:  schemas,
	})
}
