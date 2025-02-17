package driver

import (
	"net/url"

	"github.com/gorilla/mux"
	"github.com/ory/kratos-client-go/client"
	"github.com/sawadashota/kratos-frontend-go/account"
	"github.com/sawadashota/kratos-frontend-go/admin"
	"github.com/sawadashota/kratos-frontend-go/authentication"
	"github.com/sawadashota/kratos-frontend-go/driver/configuration"
	"github.com/sawadashota/kratos-frontend-go/err"
	"github.com/sawadashota/kratos-frontend-go/internal/jwt"
	"github.com/sawadashota/kratos-frontend-go/middleware"
	"github.com/sawadashota/kratos-frontend-go/salary"
	"github.com/sirupsen/logrus"
)

var _ Registry = new(RegistryDefault)

// RegistryDefault .
type RegistryDefault struct {
	l logrus.FieldLogger
	c configuration.Provider
	r Registry

	jwtParser             *jwt.Parser
	mw                    *middleware.Middleware
	kratosClient          *client.OryKratos
	kratosPublicClient    *client.OryKratos
	authenticationHandler *authentication.Handler
	accountHandler        *account.Handler
	salaryHandler         *salary.Handler
	adminHandler          *admin.Handler
	errHandler            *err.Handler
}

func (r *RegistryDefault) JWTParser() *jwt.Parser {
	if r.jwtParser == nil {
		r.jwtParser = jwt.New(r, r.c)
	}
	return r.jwtParser
}

func (r *RegistryDefault) AccountHandler() *account.Handler {
	if r.accountHandler == nil {
		r.accountHandler = account.New(r, r.c)
	}
	return r.accountHandler
}

func (r *RegistryDefault) AuthenticationHandler() *authentication.Handler {
	if r.authenticationHandler == nil {
		r.authenticationHandler = authentication.New(r, r.c)
	}
	return r.authenticationHandler
}

func (r *RegistryDefault) SalaryHandler() *salary.Handler {
	if r.salaryHandler == nil {
		r.salaryHandler = salary.New(r, r.c)
	}
	return r.salaryHandler
}

func (r *RegistryDefault) AdminHandler() *admin.Handler {
	if r.adminHandler == nil {
		r.adminHandler = admin.New(r, r.c)
	}
	return r.adminHandler
}

func (r *RegistryDefault) ErrHandler() *err.Handler {
	if r.errHandler == nil {
		r.errHandler = err.New(r, r.c)
	}
	return r.errHandler
}

func (r *RegistryDefault) RegisterRoutes(router *mux.Router) {
	r.AuthenticationHandler().RegisterRoutes(router)
	r.AccountHandler().RegisterRoutes(router)
	r.SalaryHandler().RegisterRoutes(router)
	r.AdminHandler().RegisterRoutes(router)
	r.ErrHandler().RegisterRoutes(router)
}

func NewRegistryDefault(c configuration.Provider) *RegistryDefault {
	return &RegistryDefault{
		c: c,
	}
}

// Logger .
func (r *RegistryDefault) Logger() logrus.FieldLogger {
	if r.l == nil {
		r.l = r.newLogger()
	}
	return r.l
}

func (r *RegistryDefault) newLogger() logrus.FieldLogger {
	l := logrus.New()

	l.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(r.c.LogLevel())
	if err == nil {
		l.SetLevel(level)
	}
	return l
}

func (r *RegistryDefault) Middleware() *middleware.Middleware {
	if r.mw == nil {
		r.mw = middleware.New(r, r.c)
	}
	return r.mw
}

func (r *RegistryDefault) KratosClient() *client.OryKratos {
	if r.kratosClient == nil {
		u, _ := url.Parse(r.c.KratosAdminURL())
		cl := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
			Host:     u.Host,
			BasePath: "/",
			Schemes:  []string{u.Scheme},
		})
		r.kratosClient = cl
	}
	return r.kratosClient
}

func (r *RegistryDefault) KratosPublicClient() *client.OryKratos {
	if r.kratosPublicClient == nil {
		u, _ := url.Parse("http://kratos:4433/")
		cl := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
			Host:     u.Host,
			BasePath: "/",
			Schemes:  []string{u.Scheme},
		})
		r.kratosPublicClient = cl
	}
	return r.kratosPublicClient
}
