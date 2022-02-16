package admin

import (
	"net/http"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/ory/kratos-client-go/client"
	"github.com/ory/kratos-client-go/client/admin"
	"github.com/ory/kratos-client-go/client/public"
	"github.com/ory/kratos-client-go/models"
	"github.com/sawadashota/kratos-frontend-go/middleware"
	"github.com/sawadashota/kratos-frontend-go/x"
	"github.com/sirupsen/logrus"
)

var (
	identitiesHTML *x.HTMLTemplate
	createidHTML   *x.HTMLTemplate
)

func init() {
	compileTemplate()
}

func compileTemplate() {
	box := x.NewBox(packr.New("admin", "./templates"))
	identitiesHTML = box.MustParseHTML("identities", "layout.html", "identities.html")
	createidHTML = box.MustParseHTML("create-id", "layout.html", "create-id.html")
}

type Handler struct {
	r Registry
	c Configuration
}

type Registry interface {
	Logger() logrus.FieldLogger
	Middleware() *middleware.Middleware
	KratosClient() *client.OryKratos
	KratosPublicClient() *client.OryKratos
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
	sub.HandleFunc("/create-id", h.RenderCreateId).Methods(http.MethodGet, http.MethodPost)
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

func (h *Handler) RenderCreateId(w http.ResponseWriter, r *http.Request) {
	res, err := h.r.KratosPublicClient().Public.InitializeSelfServiceRegistrationViaAPIFlow(public.NewInitializeSelfServiceRegistrationViaAPIFlowParams())

	if err != nil {
		h.r.Logger().Errorf("fail to get registration request from kratos #1: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	action := *res.Payload.Methods["password"].Config.Action
	flowID := action[strings.Index(action, "=")+1:]
	params := public.NewCompleteSelfServiceRegistrationFlowWithPasswordMethodParams()
	params.SetFlow(&flowID)

	htmlValues := struct {
		LogoutURL string
	}{
		LogoutURL: h.c.KratosLogoutURL(),
	}

	if r.Method == "GET" {
		if err1 := createidHTML.Render(w, &htmlValues); err1 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		params.SetPayload(map[string]interface{}{
			"traits.email":      r.FormValue("email"),
			"traits.name.first": r.FormValue("firstname"),
			"traits.name.last":  r.FormValue("lastname"),
			"password":          r.FormValue("password"),
		})

		_, err = h.r.KratosPublicClient().Public.CompleteSelfServiceRegistrationFlowWithPasswordMethod(params)

		if err != nil {
			h.r.Logger().Errorf("fail to get registration request from kratos #2: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/create-id", 302)
	}
}
