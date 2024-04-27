package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/fnk2077/assessment-tax/docs"
	middlewares "github.com/fnk2077/assessment-tax/middleware"
	"github.com/fnk2077/assessment-tax/postgres"
	"github.com/fnk2077/assessment-tax/tax"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func main() {

	p, err := postgres.New()
	if err != nil {
		panic(err)
	}
	taxHandler := tax.New(p)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.POST("/tax/calculations", taxHandler.TaxCalculateHandler)
	e.POST("/tax/calculations/upload-csv", taxHandler.TaxCVSCalculateHandler)

	g := e.Group("/admin")
	g.Use(middleware.BasicAuth(middlewares.AuthMiddleware))
	g.POST("/deductions/:type", taxHandler.ChangeDeductionHandler)

	go func() {
		if err := e.Start(":" + os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	fmt.Println("shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
