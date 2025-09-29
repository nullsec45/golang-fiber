package main

import (
	"github.com/gofiber/fiber/v2"
	"time"
	"fmt"
)

func main() {
	app := fiber.New(fiber.Config{
		IdleTimeout:time.Second * 5,
		WriteTimeout: time.Second * 5,
		ReadTimeout: time.Second * 5,
		Prefork:true,
	})

	app.Use("/api",func(ctx *fiber.Ctx) error {
		fmt.Println("I'm middleware before proccessing request")
		err := ctx.Next()
		fmt.Println("I'm middleware after proccessing request")
		return err
	})

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})

	app.Get("/api", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World From route /api")
	})

	app.Get("/api/hello", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World From route /api/hello")
	})


	if fiber.IsChild() {
		fmt.Println("I'm child proccess")
	}else{
		fmt.Println("I'm parent proccess")
	}

	err := app.Listen("localhost:3000")
	
	if err != nil {
		panic(err)
	}


}