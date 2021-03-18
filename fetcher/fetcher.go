package fetcher

import (
	"fmt"
	"github.com/go-redis/redis"
	"os"
	"strings"
)

type LoggedRequest struct {
	ReqID string
	Body []byte
}

type Redis struct{}

func (r Redis) Get() ([]LoggedRequest, error) {
	host := os.Getenv("REDIS_ADDR")
	if len(host) == 0 {
		return nil, fmt.Errorf("required environment variable named REDIS_ADDR missing")
	}

	list := os.Getenv("REDIS_LISTNAME")
	if len(list) == 0 {
		return nil, fmt.Errorf("required environment variable named REDIS_LISTNAME missing")
	}

	c := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: os.Getenv("REDIS_KEY"),
		DB:       0, // use default DB
	})
	defer c.Close()

	count, err := c.LLen(list).Result()
	if err != nil {
		return nil, err
	}

	var results []LoggedRequest
	for i := 0; i < int(count); i++ {
		reqID, b, err := r.next(c, list)
		if err != nil {
			return nil, err
		}

		results = append(results, LoggedRequest{
			ReqID: reqID,
			Body:  b,
		})
	}

	return results, nil
}

func (r Redis) next(c *redis.Client, list string) (reqID string, b []byte, err error) {
	s, err := c.LPop(list).Result()
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}

	buf := strings.Split(s, "\n|\n")
	if len(buf) != 2 {
		return reqID, b, fmt.Errorf("unable to split request result")
	}

	b = []byte(buf[0])
	reqID = buf[1]
	return
}

