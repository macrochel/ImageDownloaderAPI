package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DownloadRequest struct {
	URL string `json:"url"`
}

func main() {
	app := fiber.New()

	app.Post("/download/FromUrl", func(c *fiber.Ctx) error {
		var req DownloadRequest
		if err := c.BodyParser(&req); err != nil {
			return err
		}
		fileName := downloadImage(req.URL, false)
		return c.SendString("http://192.168.0.1:2050/static/" + fileName)
	})

	app.Post("/download/FromUrlWithProxy", func(c *fiber.Ctx) error {
		var req DownloadRequest
		if err := c.BodyParser(&req); err != nil {
			return err
		}
		fileName := downloadImage(req.URL, true)
		return c.SendString("http://192.168.0.1:2050/static/" + fileName)
	})

	app.Post("/download/FromFile", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return err
		}
		fileName := uuid.New().String() + ".jpg"
		destination := "./statics/" + fileName
		if err := c.SaveFile(file, destination); err != nil {
			return err
		}
		return c.SendString("http://192.168.0.1:2050/static/" + fileName)
	})

	app.Get("/clear", func(c *fiber.Ctx) error {
		cleanStatics()
		customResponse := map[string]interface{}{
			"status": "success",
		}

		return c.JSON(customResponse)
	})

	app.Static("/static", "./statics")

	log.Fatal(app.Listen(":2050"))
}

func downloadImage(urlString string, useProxy bool) string {
	var client *http.Client

	if useProxy {
		proxyURL := ""
		proxy := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(proxyURL)
		}

		transport := &http.Transport{
			Proxy: proxy,
		}

		client = &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		}
	} else {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	request, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	hash := md5.New()
	hash.Write([]byte(urlString))
	hashInBytes := hash.Sum(nil)
	fileName := hex.EncodeToString(hashInBytes) + ".jpg"

	file, err := os.Create(filepath.Join("./statics", fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return fileName
}

func cleanStatics() {
	os.RemoveAll("./statics")
	os.MkdirAll("./statics", 0700)
}
