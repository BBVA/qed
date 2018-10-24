package backends

import (
	"fmt"

	"github.com/go-redis/redis"
)

type Client struct {
	rcli *redis.Client
}

func NewRedisClient() *Client {
	c := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := c.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
	return &Client{rcli: c}
}

func (c *Client) Publish(key, value string) {
	err := c.rcli.Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}
