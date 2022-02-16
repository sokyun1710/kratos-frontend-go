package driver

import (
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

// Registry .
type Registry interface {
	Logger() logrus.FieldLogger
	JWTParser() *jwt.Parser
	Middleware() *middleware.Middleware
	KratosClient() *client.OryKratos
	KratosPublicClient() *client.OryKratos
	AccountHandler() *account.Handler
	AuthenticationHandler() *authentication.Handler
	SalaryHandler() *salary.Handler
	AdminHandler() *admin.Handler
	ErrHandler() *err.Handler
	RegisterRoutes(router *mux.Router)
}

// NewRegistry .
func NewRegistry(c configuration.Provider) Registry {
	return NewRegistryDefault(c)
}
