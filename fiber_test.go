package main

import (
	_ "embed"
	"github.com/gofiber/fiber/v2"
	"testing"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"io"
	"net/http"
	"strings"
	"bytes"
	"mime/multipart"
	"encoding/json"
	"errors"
	"github.com/gofiber/template/mustache/v2"
	"fmt"
)

var engine=mustache.New("./template",".mustache")


var app = fiber.New(fiber.Config{
    Views: engine,
    ErrorHandler: func(ctx *fiber.Ctx, err error) error {
        code := fiber.StatusInternalServerError
        if e, ok := err.(*fiber.Error); ok {
            code = e.Code
        }
        return ctx.Status(code).SendString("Error: " + err.Error())
    },
})


func TestRoutingHelloWorld(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})

	request := httptest.NewRequest("GET","/", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t,"Hello World", string(bytes))
}


func TestCtx(t *testing.T) {
	app.Get("/hello", func(ctx *fiber.Ctx) error {
		name := ctx.Query("name","Guest")
		return ctx.SendString("Hello "+name)
	})

	request := httptest.NewRequest("GET","/hello?name=Fajar", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t,"Hello Fajar", string(bytes))

	request = httptest.NewRequest("GET","/hello", nil)
	response, err = app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err = io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t,"Hello Guest", string(bytes))
}

func TestHttpRequest(t *testing.T) {
	app.Get("/request", func(ctx *fiber.Ctx) error {
		first := ctx.Get("firstname")
		last := ctx.Cookies("lastname")
		return ctx.SendString("Hello "+first+" "+last)
	})

	request := httptest.NewRequest("GET","/request", nil)
	request.Header.Set("firstname","Rama")
	request.AddCookie(&http.Cookie{Name:"lastname",Value:"Fajar"})
	response, err := app.Test(request)
	assert.Nil(t, err)
	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t,"Hello Rama Fajar", string(bytes))
}

func TestRouteParameter(t *testing.T){
	app.Get("/users/:userId/orders/:orderId", func(ctx *fiber.Ctx) error {
		userId := ctx.Params("userId")
		orderId := ctx.Params("orderId")
		return ctx.SendString("Get Order "+orderId+" From User "+userId)
	})

	request := httptest.NewRequest("GET", "/users/1/orders/2", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Equal(t, "Get Order 2 From User 1", string(bytes))
}

func TestFormRequest(t *testing.T) {
	app.Post("/hello", func(ctx *fiber.Ctx) error  {
		name := ctx.FormValue("name")
		return ctx.SendString("Hello "+name)
	})

	body := strings.NewReader("name=Fajar")
	request := 	httptest.NewRequest("POST", "/hello", body)
	request.Header.Set("Content-Type","application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t,200,response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Fajar", string(bytes))
}

//go:embed source/contoh.txt
var contohFile []byte

func TestFormUpload(t *testing.T) {
	app.Post("/upload", func(ctx *fiber.Ctx) error {
		file, err := ctx.FormFile("file")

		if err != nil {
			return err
		}

		err = ctx.SaveFile(file, "./target/"+file.Filename)
		if err != nil {
			return err
		}

		return ctx.SendString("Upload Success")
	})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	file, _ := writer.CreateFormFile("file","contoh.txt")
	file.Write(contohFile)
	writer.Close()

	request := httptest.NewRequest("POST","/upload", body)
	request.Header.Set("Content-Type",writer.FormDataContentType())
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Upload Success", string(bytes))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestRequestBody(t *testing.T) {
	app.Post("/login", func(ctx *fiber.Ctx) error  {
		body := ctx.Body()

		request := new(LoginRequest)
		err := json.Unmarshal(body, request)

		if err != nil {
			return err
		}

		return ctx.SendString("Hello "+request.Username)
	})

	body := strings.NewReader(`{"username":"Fajar","password":"rahasia"}`)
	request := 	httptest.NewRequest("POST", "/login", body)
	request.Header.Set("Content-Type","application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t,200,response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Fajar", string(bytes))
}

type RegisterRequest struct {
	Username string `json:"username" xml:"username" form:"username"`
	Password string `json:"password" xml:"password" form:"password"`
	Name     string `json:"name" xml:"name" form:"name"`
}

func TestBodyParser(t *testing.T) {
	app.Post("/register", func(ctx *fiber.Ctx) error {
		request := new(RegisterRequest)
		err := ctx.BodyParser(request)
		if err != nil {
			return err
		}

		return ctx.SendString("Register Success "+request.Username)
	})
}


func TestBodyParserJSON(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`{"username":"Fajar","password":"rahasia","name":"Rama Fajar"}`)
	request := 	httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type","application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t,200,response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Success Fajar", string(bytes))
}

func TestBodyParserForm(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`username=Fajar&password=rahasia&name=Rama+Fajar`)
	request := 	httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type","application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t,200,response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Success Fajar", string(bytes))
}

func TestBodyParserXML(t *testing.T) {
	TestBodyParser(t)

	body := strings.NewReader(`
		<RegisterRequest>
			<username>Fajar</username>
			<password>rahasia</password>
		</RegisterRequest>
	`)
	request := 	httptest.NewRequest("POST", "/register", body)
	request.Header.Set("Content-Type","application/xml")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t,200,response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Success Fajar", string(bytes))
}

func TestResponseJSON(t *testing.T) {
	app.Get("/user", func(ctx *fiber.Ctx) error  {
		return ctx.JSON(fiber.Map{
			"username":"nullsec45",
			"name":"Fajar",
		})
	})

	request := 	httptest.NewRequest("GET", "/user",nil)
	request.Header.Set("Accept","application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t,200,response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.JSONEq(t, `{"username":"nullsec45","name":"Fajar"}`, string(bytes))
}

func TestDownloadFile(t *testing.T) {
	app.Get("/download", func(ctx *fiber.Ctx) error  {
		return ctx.Download("./source/example_upload.txt","upload.txt")
	})

	request := httptest.NewRequest("GET","/download",nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t,200,response.StatusCode)
	assert.Equal(t, "attachment; filename=\"upload.txt\"", response.Header.Get("Content-Disposition"))
	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "This is sample file for upload", string(bytes))

}

func TestRoutingGroup(t *testing.T) {
	helloWorld := func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	}

	api := app.Group("/api")
	api.Get("/hello", helloWorld)
	api.Get("/world", helloWorld)

	web := app.Group("/web")
	web.Get("/hello", helloWorld)
	web.Get("/world", helloWorld)

	request := httptest.NewRequest("GET","/api/hello",nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t,200,response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(bytes))
}

func TestStatic(t *testing.T) {
	app.Static("/public", "./source")

	request := httptest.NewRequest("GET", "/public/contoh.txt", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "contoh file yang ingin ditransfer", string(bytes))
}

func TestErrorHandling(t *testing.T) {
	app.Get("/error", func(ctx *fiber.Ctx) error {
		return errors.New("ups")
	})

	request := httptest.NewRequest("GET", "/error", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 500, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Error: ups", string(bytes))
}

func TestView(t *testing.T) {
	app.Get("/view", func(ctx *fiber.Ctx) error {
		return ctx.Render("index", fiber.Map{
			"title":"Hello Title",
			"header":"Hello Header",
			"content":"Hello Content",
		})
	})

	request := httptest.NewRequest("GET", "/view", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	bytes, err := io.ReadAll(response.Body)
	fmt.Println(string(bytes))
	assert.Nil(t, err)
	assert.Contains(t, string(bytes), "Hello Title")
	assert.Contains(t, string(bytes), "Hello Header")
	assert.Contains(t, string(bytes), "Hello Content")
}

func TestClient(t *testing.T) {
	client := fiber.AcquireClient()

	agent := client.Get("http://example.com")
	status, response, errors := agent.String()
	assert.Nil(t, errors)
	assert.Equal(t, 200, status)
	assert.Contains(t, response, "Example Domain")

	// fmt.Println(response)
	defer fiber.ReleaseClient(client)
}