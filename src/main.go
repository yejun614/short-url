package main

import (
	"context"
	"crypto/sha512"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/redis/go-redis/v9"
)

type UrlData struct {
	URL     string `json:"url" xml:"url" form:"url"`
	AdminPW string `json:"admin_pw" xml:"admin-pw" form:"admin_pw"`
}

var (
	err         error
	FlagRedis   string
	RedisClient *redis.Client
)

func parseFlag() {
	flag.StringVar(
		&FlagRedis,
		"redis",
		"redis://localhost:6379",
		"Redis connection string (redis://<user>:<pass>@localhost:6379/<db>)",
	)
	flag.Parse()
}

func SHA512(data string) string {
	hash := sha512.New()
	hash.Write([]byte(data))
	result := hash.Sum(nil)
	return fmt.Sprintf("%x", result)
}

func homePage(c *fiber.Ctx) error {
	return c.Render("index", WebConfig, "layouts/default")
}

func addPage(c *fiber.Ctx) error {
	return c.Render("add_url", WebConfig, "layouts/default")
}

func delPage(c *fiber.Ctx) error {
	return c.Render("del_url", WebConfig, "layouts/default")
}

func getShortURL(c *fiber.Ctx) error {
	key := c.Params("key")
	ctx := context.Background()

	val, err := RedisClient.Get(ctx, fmt.Sprintf("short:%s", key)).Result()
	if err == redis.Nil {
		return fiber.ErrNotFound
	} else {
		return c.Redirect(val)
	}
}

func postShortURL(c *fiber.Ctx) error {
	// get key
	key := c.Params("key")

	// body parser
	p := new(UrlData)
	if err := c.BodyParser(p); err != nil {
		panic(err)
	}

	// body validation
	if p.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URL is required",
		})
	} else if p.AdminPW == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Admin password is required",
		})
	} else if _, err := url.ParseRequestURI("http://test/" + key); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("key validation error (%v)", err),
		})
	} else if _, err := url.ParseRequestURI(p.URL); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("URL validation error (%v)", err),
		})
	}

	// redis context
	ctx := context.Background()

	// duplicate check
	err = RedisClient.Get(ctx, fmt.Sprintf("short:%s", key)).Err()
	if err != redis.Nil {
		// error
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Duplicated Key",
		})
	}

	// hash password
	hashAdminPW := SHA512(p.AdminPW)

	// add url
	err = RedisClient.Set(ctx, fmt.Sprintf("short:%s", key), p.URL, 0).Err()
	if err != nil {
		panic(err)
	}

	// set admin pw
	err = RedisClient.Set(ctx, fmt.Sprintf("admin:pw:%s", key), hashAdminPW, 0).Err()
	if err != nil {
		panic(err)
	}

	// OK
	return c.SendString("Success")
}

func deleteShortURL(c *fiber.Ctx) error {
	// get key
	key := c.Params("key")

	// body parser
	p := new(UrlData)
	if err := c.BodyParser(p); err != nil {
		panic(err)
	}

	// body validation
	if p.AdminPW == "" {
		// error
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Admin password is required",
		})
	}

	// redis context
	ctx := context.Background()

	// check exists
	_, err = RedisClient.Get(ctx, fmt.Sprintf("short:%s", key)).Result()
	if err == redis.Nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	// check password
	hashPW := SHA512(p.AdminPW)

	adminPW, err := RedisClient.Get(ctx, fmt.Sprintf("admin:pw:%s", key)).Result()
	if err == redis.Nil {
		panic(err)
	}

	if adminPW != hashPW {
		// error
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// delete url
	err = RedisClient.Del(ctx, fmt.Sprintf("short:%s", key)).Err()
	if err != nil {
		panic(err)
	}

	// OK
	return c.SendString("Success")
}

func main() {
	// flag
	parseFlag()

	// redis
	redisOpt, err := redis.ParseURL(FlagRedis)
	if err != nil {
		panic(err)
	}
	RedisClient = redis.NewClient(redisOpt)

	// redis context
	ctx := context.Background()

	// fiber template engine
	engine := html.New("./web", ".html")

	// fiber
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code if it's a *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			// Send custom error page
			err = ctx.Status(code).Render(fmt.Sprintf("error/%d", code), WebConfig, "layouts/default")
			if err != nil {
				// In case the SendFile fails
				return ctx.Status(code).SendString(fmt.Sprintf("%d error", code))
			}

			// Return from handler
			return nil
		},
	})
	// middleware
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())

	// static
	app.Static("/static", "./static")

	// index
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("web")
	})

	// web (dashboard)
	app.Get("/web", homePage)
	app.Get("/web/add", addPage)
	app.Get("/web/del", delPage)

	// key (url short api)
	app.Get("/:key", getShortURL)
	app.Post("/:key", postShortURL)
	app.Delete("/:key", deleteShortURL)

	// admin
	admin := app.Group("/admin")
	admin.Get("/metrics", monitor.New())

	// get addr
	var addr string
	addr, err = RedisClient.Get(ctx, "conf:addr").Result()
	if err == redis.Nil {
		addr = "localhost:8000"
		if err := RedisClient.Set(ctx, "conf:addr", addr, 0).Err(); err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}

	// start server
	log.Printf("addr: %s\n", addr)
	log.Fatal(app.Listen(addr))
}
