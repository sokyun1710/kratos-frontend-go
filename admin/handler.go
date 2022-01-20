package admin

import (
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/ory/kratos-client-go/client"
	"github.com/ory/kratos-client-go/client/admin"
	"github.com/ory/kratos-client-go/models"
	"github.com/sawadashota/kratos-frontend-go/middleware"
	"github.com/sawadashota/kratos-frontend-go/x"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	identitiesHTML *x.HTMLTemplate
)

func init() {
	compileTemplate()
}

func compileTemplate() {
	box := x.NewBox(packr.New("admin", "./templates"))
	identitiesHTML = box.MustParseHTML("identities", "layout.html", "identities.html")
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
	sub.Use(h.r.Middleware().Authorize)
	sub.HandleFunc("/identities", h.RenderIdentities).Methods(http.MethodGet)
}

func (h *Handler) RenderIdentities(w http.ResponseWriter, r *http.Request) {
	params := admin.NewListIdentitiesParams().WithDefaults()
	res, err := h.r.KratosClient().Admin.ListIdentities(params)

	if err != nil {
		h.r.Logger().Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if res.Error() == "" {
		h.r.Logger().Error(res.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	htmlValues := struct {
		LogoutURL string
		List      []*models.Identity
	}{
		LogoutURL: h.c.KratosLogoutURL(),
		List:      res.GetPayload(),
	}

	if err := identitiesHTML.Render(w, &htmlValues); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
