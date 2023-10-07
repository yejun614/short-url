package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type PostShortURLPayload struct {
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

func getShortUrl(c *fiber.Ctx) error {
	key := c.Params("key")
	ctx := context.Background()

	val, err := RedisClient.Get(ctx, fmt.Sprintf("short:%s", key)).Result()
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	} else {
		return c.Redirect(val)
	}
}

func postShortUrl(c *fiber.Ctx) error {
	key := c.Params("key")
	p := new(PostShortURLPayload)
	if err := c.BodyParser(p); err != nil {
		return err
	}

	ctx := context.Background()

	// duplicate check
	err = RedisClient.Get(ctx, fmt.Sprintf("short:%s", key)).Err()
	if err != redis.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Duplicated url",
		})
	}

	// add url
	err = RedisClient.Set(ctx, fmt.Sprintf("short:%s", key), p.URL, 0).Err()
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"success": true,
	})
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

	// fiber
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/s/:key", getShortUrl)
	app.Post("/s/:key", postShortUrl)

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
