package authentication

import (
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
	signInHTML *x.HTMLTemplate
	signUpHTML *x.HTMLTemplate
)

func init() {
	compileTemplate()
}

func compileTemplate() {
	box := x.NewBox(packr.New("authentication", "./templates"))
	signInHTML = box.MustParseHTML("signin", "layout.html", "signin.html")
	signUpHTML = box.MustParseHTML("signup", "layout.html", "signup.html")
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
	KratosLoginURL() string
	KratosRegistrationURL() string
}

func New(r Registry, c Configuration) *Handler {
	return &Handler{
		r: r,
		c: c,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	sub := router.NewRoute().Subrouter()
	sub.Use(h.r.Middleware().ValidateFormRequest)
	sub.HandleFunc("/auth/signin", h.RenderSignInForm).Methods(http.MethodGet)
	sub.HandleFunc("/auth/signup", h.RenderSignUpForm).Methods(http.MethodGet)
}

func (h *Handler) RenderSignInForm(w http.ResponseWriter, r *http.Request) {
	requestCode := r.URL.Query().Get("flow")
	params := public.NewGetSelfServiceLoginFlowParams().WithID(requestCode)
	res, err := h.r.KratosClient().Public.GetSelfServiceLoginFlow(params)

	if err != nil {
		h.r.Logger().Errorf("fail to get login request from kratos: %s", err)
		http.Redirect(w, r, h.c.KratosLoginURL(), http.StatusFound)
		return
	}

	if res.Error() == "" {
		h.r.Logger().Errorf("fail to get login request from kratos: %s", res.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	form := res.GetPayload().Methods["password"].Config
	sso := res.GetPayload().Methods["oidc"].Config

	htmlValues := struct {
		Form *models.LoginFlowMethodConfig
		OIDC *models.LoginFlowMethodConfig
	}{
		Form: form,
		OIDC: sso,
	}

	if err := signInHTML.Render(w, &htmlValues); err != nil {
		h.r.Logger().Errorf("fail to render HTML: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) RenderSignUpForm(w http.ResponseWriter, r *http.Request) {
	requestCode := r.URL.Query().Get("flow")
	params := public.NewGetSelfServiceRegistrationFlowParams().WithID(requestCode)
	res, err := h.r.KratosClient().Public.GetSelfServiceRegistrationFlow(params)

	if err != nil {
		h.r.Logger().Errorf("fail to get registration request from kratos: %s", err)
		http.Redirect(w, r, h.c.KratosRegistrationURL(), http.StatusFound)
		return
	}

	if res.Error() == "" {
		h.r.Logger().Errorf("fail to get registration request from kratos: %s", res.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	form := res.GetPayload().Methods["password"].Config

	htmlValues := struct {
		Form *models.RegistrationFlowMethodConfig
	}{
		Form: form,
	}

	if err := signUpHTML.Render(w, &htmlValues); err != nil {
		h.r.Logger().Errorf("fail to render HTML: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
