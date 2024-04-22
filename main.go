package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/fnk2077/assessment-tax/postgres"
	"github.com/fnk2077/assessment-tax/tax"
	auth "github.com/fnk2077/assessment-tax/middleware"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	e.POST("/tax/calculations", taxHandler.TaxCalculate)
	e.POST("/tax/calculations/upload-csv", taxHandler.ReadTaxCSV)

	g := e.Group("/admin")
	g.Use(middleware.BasicAuth(auth.AuthMiddleware))
	g.POST("/deductions/personal", taxHandler.ChangePersonalDeduction)
	g.POST("/deductions/k-receipt", taxHandler.ChangeKReciept)

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
