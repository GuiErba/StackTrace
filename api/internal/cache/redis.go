package cache

import (
    "context"
    "os"

    "github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

func Init() {
    opt, err := redis.ParseURL(os.Getenv("UPSTASH_REDIS_URL"))
    if err != nil {
        panic("Redis URL invalid: " + err.Error())
    }
    Client = redis.NewClient(opt)
}
