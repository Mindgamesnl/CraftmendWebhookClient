package CraftmendWebhookClient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type WebhookClient struct {
	usedKeys *TTLMap
	handlers []func(event string, data string)
}

type responseFormat []struct {
	Data  string `json:"data"`
	Event string `json:"event"`
	ID    string `json:"id"`
}

func CreateWebhookClient(password string) *WebhookClient {
	client := WebhookClient{
		New(100, 60 * 2),
		[]func(event string, data string){},
	}

	go func() {
		for _ = range time.Tick(time.Second * 5) {
			data, _ := doGetOrFail("https://webhooks.craftmend.com/" + password)
			var responses responseFormat
			json.Unmarshal([]byte(data), &responses)

			for i := range responses {
				response := responses[i]
				if client.usedKeys.Get(response.ID) != response.ID {
					for i2 := range client.handlers {
						client.handlers[i2](response.Event, response.Data)
					}
					client.usedKeys.Put(response.ID, response.ID)
				}
			}
		}
	}()

	return &client
}

func (i *WebhookClient) On(handler func(event string, data string)) {
	i.handlers = append(i.handlers, handler)
}

func doGetOrFail(endpoint string) (string, error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", endpoint, nil)

	if err != nil {
		return "", err
	}

	urlValues := url.Values{}

	request.PostForm = urlValues

	resp, e := client.Do(request)
	if e != nil {
		return "", e
	}

	rb, er := ioutil.ReadAll(resp.Body)
	if er != nil {
		return "", er
	}
	responseBody := string(rb)

	return responseBody, nil
}