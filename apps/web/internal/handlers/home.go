package handlers

import (
	"net/http"

	"github.com/terjelafton/platform/apps/web/internal/templates"
)

func HandleHomePage(w http.ResponseWriter, r *http.Request) {
	templates.HomePage().Render(r.Context(), w)
}
