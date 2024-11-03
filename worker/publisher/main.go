package publisher

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/go-redis/redis"
)

func main() {
	log.Println("Publisher started")

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", "redis", "6379"),
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		log.Fatal("Unbale to connect to Redis", err)
	}

	log.Println("Connected to Redis server")

	for i := 0; i < 3000; i++ {
		err = publishTicketReceivedEvent(rdb)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func publishTicketReceivedEvent(client *redis.Client) error {
	log.Println("Publishing event to Redis")

	err := client.XAdd(&redis.XAddArgs{
		Stream:       "tickets",
		MaxLen:       0,
		MaxLenApprox: 0,
		ID:           "",
		Values: map[string]interface{}{
			"whatHappened": string("ticket received"),
			"ticketID":     int(rand.Intn(100000000)),
			"ticketData":   string("some ticket data"),
		},
	}).Err()

	return err
}
