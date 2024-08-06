package agent

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
)

func (s *server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		apiKey, ok := os.LookupEnv("API_TOKEN")
		if !ok {
			panic("API_TOKEN env variable is not set")
		}
		return key == apiKey, nil
	}))

	r := e.Group("/v1")
	api := r.Group("/agent")

	api.GET("/info", s.GetInformationHandler)
	api.POST("/recipe", s.ExecuteRecipeHandler)

	return e
}

func (s *server) GetInformationHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *server) ExecuteRecipeHandler(c echo.Context) error {
	var r ExecuteRecipeRequest

	err := json.NewDecoder(c.Request().Body).Decode(&r)
	if err != nil {
		return err
	}

	resp := map[string]string{
		"name":              r.Name,
		"kernel_parameters": r.KernelParameters[0].Value,
	}
	return c.JSON(http.StatusOK, resp)
}
