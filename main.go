package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sawadashota/kratos-frontend-go/driver"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	d := driver.NewDefaultDriver()
	router := mux.NewRouter()
	d.Registry().RegisterRoutes(router)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", d.Configuration().Port()), router); err != nil {
		logrus.Fatalf("fail to staring server: %s", err)
	}
}
