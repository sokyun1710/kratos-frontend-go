package account

import (
	"encoding/json"
	"github.com/go-openapi/runtime"
	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
	"github.com/ory/kratos-client-go/client"
	"github.com/ory/kratos-client-go/client/public"
	"github.com/ory/kratos-client-go/models"
	"github.com/sawadashota/kratos-frontend-go/middleware"
	"github.com/sawadashota/kratos-frontend-go/x"
	"github.com/sirupsen/logrus"
	"net/http"
)

var (
	homeHTML     *x.HTMLTemplate
	settingsHTML *x.HTMLTemplate
)

func init() {
	compileTemplate()
}

func compileTemplate() {
	box := x.NewBox(packr.New("account", "./templates"))
	homeHTML = box.MustParseHTML("home", "layout.html", "home.html")
	settingsHTML = box.MustParseHTML("home", "layout.html", "settings.html")
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
	KratosSettingsURL() string
	JWKsURL() string
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
	sub.HandleFunc("/", h.RenderHome).Methods(http.MethodGet)
	setting := sub.NewRoute().Subrouter()
	setting.Use(h.r.Middleware().ValidateFormRequest)
	sub.HandleFunc("/settings", h.RenderSettingForms).Methods(http.MethodGet)
}

func (h *Handler) RenderHome(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.GetClaimsFromContext(r)

	if err != nil {
		h.r.Logger().Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	claimsJSON, err := json.MarshalIndent(claims, "", "  ")

	if err != nil {
		h.r.Logger().Errorf("fail to marshal claims: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	htmlValues := struct {
		LogoutURL  string
		ClaimsJSON string
	}{
		LogoutURL:  h.c.KratosLogoutURL(),
		ClaimsJSON: string(claimsJSON),
	}

	if err := homeHTML.Render(w, &htmlValues); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) RenderSettingForms(w http.ResponseWriter, r *http.Request) {
	requestCode := r.URL.Query().Get("flow")
	params := public.NewGetSelfServiceSettingsFlowParams().WithID(requestCode)
	authInfo := runtime.ClientAuthInfoWriterFunc(params.WriteToRequest)
	res, err := h.r.KratosClient().Public.GetSelfServiceSettingsFlow(params, authInfo)

	if err != nil {
		h.r.Logger().Errorf("fail to get login request from kratos: %s", err)
		http.Redirect(w, r, h.c.KratosSettingsURL(), http.StatusFound)
		return
	}

	if res.Error() == "" {
		h.r.Logger().Errorf("fail to get settings request from kratos: %s", res.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.r.Logger().Debugf("payload: %v", res.GetPayload())

	htmlValues := struct {
		LogoutURL string
		Password  models.SettingsFlowMethod
		Profile   models.SettingsFlowMethod
		OIDC      models.SettingsFlowMethod
	}{
		LogoutURL: h.c.KratosLogoutURL(),
		Password:  res.GetPayload().Methods["password"],
		Profile:   res.GetPayload().Methods["profile"],
		OIDC:      res.GetPayload().Methods["oidc"],
	}

	if err := settingsHTML.Render(w, &htmlValues); err != nil {
		h.r.Logger().Errorf("fail to render HTML: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
