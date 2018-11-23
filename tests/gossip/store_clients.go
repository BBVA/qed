/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"fmt"

	"github.com/go-redis/redis"
)

type StoreClient interface {
	Put(key, value string)
}

type RedisCli struct {
	rcli *redis.Client
}

func NewRedisClient() *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := c.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
	return &RedisCli{rcli: c}
}

func (c *RedisCli) Put(key, value string) {
	err := c.rcli.Set(key, value, 0).Err()
	if err != nil {
		panic(err)
	}
}

// TODO: SeMaaS
