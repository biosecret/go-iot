package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/biosecret/go-iot/database"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type session struct {
	val          float64
	stateChannel chan string
}

type sessionsLock struct {
	MU       sync.Mutex
	sessions []*session
}

func (sl *sessionsLock) addSession(s *session) {
	sl.MU.Lock()
	sl.sessions = append(sl.sessions, s)
	sl.MU.Unlock()
}

func (sl *sessionsLock) removeSession(s *session) {
	sl.MU.Lock()
	idx := slices.Index(sl.sessions, s)
	if idx != -1 {
		sl.sessions[idx] = nil
		sl.sessions = slices.Delete(sl.sessions, idx, idx+1)
	}
	sl.MU.Unlock()
}

func Filter[T any](filter func(n T) bool) func(T []T) []T {
	return func(list []T) []T {
		r := make([]T, 0, len(list))
		for _, n := range list {
			if filter(n) {
				r = append(r, n)
			}
		}
		return r
	}
}

var currentSessions sessionsLock

func formatSSEMessage(eventType string, data any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	m := map[string]any{
		"data": data,
	}

	err := enc.Encode(m)
	if err != nil {
		return "", nil
	}
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf("event: %s\n", eventType))
	sb.WriteString(fmt.Sprintf("retry: %d\n", 15000))
	sb.WriteString(fmt.Sprintf("data: %v\n\n", buf.String()))

	return sb.String(), nil
}

func HandleSendMsg(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	query := c.Query("query")

	log.Printf("New Request\n")

	stateChan := make(chan string)

	val, err := strconv.ParseFloat(query, 64)
	if err != nil {
		val = 0
	}

	s := session{
		val:          val,
		stateChannel: stateChan,
	}

	currentSessions.addSession(&s)

	notify := c.Context().Done()

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		keepAliveTickler := time.NewTicker(15 * time.Second)
		keepAliveMsg := ":keepalive\n"

		// listen to signal to close and unregister (doesn't seem to be called)
		go func() {
			<-notify
			log.Printf("Stopped Request\n")
			currentSessions.removeSession(&s)
			keepAliveTickler.Stop()
		}()

		for loop := true; loop; {
			select {

			case ev := <-stateChan:
				sseMessage, err := formatSSEMessage("current-value", ev)
				if err != nil {
					log.Printf("Error formatting sse message: %v\n", err)
					continue
				}

				// send sse formatted message
				_, err = fmt.Fprintf(w, sseMessage)

				if err != nil {
					log.Printf("Error while writing Data: %v\n", err)
					continue
				}

				err = w.Flush()
				if err != nil {
					log.Printf("Error while flushing Data: %v\n", err)
					currentSessions.removeSession(&s)
					keepAliveTickler.Stop()
					loop = false
					break
				}
			case <-keepAliveTickler.C:
				fmt.Fprintf(w, keepAliveMsg)
				err := w.Flush()
				if err != nil {
					log.Printf("Error while flushing: %v.\n", err)
					currentSessions.removeSession(&s)
					keepAliveTickler.Stop()
					loop = false
					break
				}
			}
		}

		log.Println("Exiting stream")
	}))

	return nil
}

func HandleSSE(c *fiber.Ctx) error {

	HandleSendMsg(c)

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			/*
				case <-ctxTimeout.Done():
					fmt.Println("Ticker stopped")
					return
			*/
			case <-ticker.C:
				// fmt.Println("Tick at", t)
				wg := &sync.WaitGroup{}

				// send a broadcast event, so all clients connected
				// will receive it, by filtering based on some info
				// stored in the session it is possible to address
				// only specific clients
				for _, s := range currentSessions.sessions {
					wg.Add(1)
					go func(cs *session) {
						defer wg.Done()
						rows, err := database.GetDB().Query("SELECT topic, message FROM sensors ORDER BY id DESC LIMIT 1")
						if err != nil {
							log.Printf("Error fetching data: %v", err)
							return // Nếu có lỗi thì thoát
						}
						defer rows.Close()

						var topic, message string
						if rows.Next() {
							err := rows.Scan(&topic, &message)
							if err != nil {
								log.Printf("Error scanning data: %v", err)
								return // Nếu có lỗi quét dữ liệu thì thoát
							}
						} else {
							log.Println("No data found in sensors table.")
							return // Nếu không có dữ liệu, không gửi gì
						}

						// Gửi dữ liệu vào kênh stateChannel của session
						cs.stateChannel <- message
					}(s)
				}
				wg.Wait()
			}
		}
	}()

	return nil
}

func handleMQTTMessage(msg mqtt.Message) {
	fmt.Printf("* [%s] %s\n", msg.Topic(), string(msg.Payload()))

	query := "INSERT INTO sensors (topic, message) VALUES ($1, $2)"
	_, err := database.GetDB().Exec(query, msg.Topic(), string(msg.Payload()))
	if err != nil {
		log.Printf("Lỗi khi lưu tin nhắn vào database: %v", err)
	}
}

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	// opts.SetUsername(uri.User.Username())
	// password, _ := uri.User.Password()
	// opts.SetPassword(password)
	opts.SetClientID(clientId)
	return opts
}

func listen(uri *url.URL, topic string) {
	client := connect("sub", uri)
	client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		handleMQTTMessage(msg)
	})
}

// Khởi động MQTT client và subscribe tới topic
func InitMQTTSubscriber() {
	uri, err := url.Parse(os.Getenv("MQTT_URL"))
	if err != nil {
		log.Fatal(err)
	}

	topic := uri.Path[1:len(uri.Path)]
	if topic == "" {
		topic = "test"
	}

	go listen(uri, topic)

	// client := connect("pub", uri)
	// timer := time.NewTicker(1 * time.Second)
	// for t := range timer.C {
	// 	client.Publish(topic, 0, false, t.String())
	// }

}
