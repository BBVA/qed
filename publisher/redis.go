package main

// type publisher interface {
// 	Publish(key, value string)
// }

// type client struct {
// 	rcli *redis.Client
// }

// func NewClient() *client {
// 	c := redis.NewClient(&redis.Options{
// 		Addr:     "localhost:6379",
// 		Password: "", // no password set
// 		DB:       0,  // use default DB
// 	})

// 	pong, err := c.Ping().Result()
// 	fmt.Println(pong, err)
// 	// Output: PONG <nil>
// 	return &client{rcli: c}
// }

// func (c *client) Publish(key, value string) {
// 	err := c.rcli.Set(key, value, 0).Err()
// 	if err != nil {
// 		panic(err)
// 	}
// }
