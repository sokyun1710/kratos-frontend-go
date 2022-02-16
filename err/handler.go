package err

import (
	"net/http"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/ory/kratos-client-go/client"
	"github.com/sawadashota/kratos-frontend-go/middleware"
	"github.com/sawadashota/kratos-frontend-go/x"
	"github.com/sirupsen/logrus"
)

var (
	errHTML *x.HTMLTemplate
)

func init() {
	compileTemplate()
}

func compileTemplate() {
	box := x.NewBox(packr.New("err", "./templates"))
	errHTML = box.MustParseHTML("err", "layout.html", "err.html")
}

type Handler struct {
	r Registry
	c Configuration
}

type Registry interface {
	Logger() logrus.FieldLogger
	Middleware() *middleware.Middleware
	KratosClient() *client.OryKratos
}

type Configuration interface {
	KratosLogoutURL() string
}

func New(r Registry, c Configuration) *Handler {
	return &Handler{
		r: r,
		c: c,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	sub := router.NewRoute().Subrouter()
	sub.HandleFunc("/err", h.RenderError).Methods(http.MethodGet)
}

func (h *Handler) RenderError(w http.ResponseWriter, r *http.Request) {
	htmlValues := struct {
		Message string
	}{
		Message: "You are not allowed to see this page!",
	}

	if err := errHTML.Render(w, &htmlValues); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
