package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var (
	HTTPProxy         *url.URL
	Payload           LewdPic
	OldPayload        LewdPic
	DiscordWebHookURL string
	FistRunning       *bool
)

const (
	EndPoint = "https://hr.hanime.tv/api/v8/community_uploads?channel_name__in[]=nsfw-general&query_method=offset&__offset=0"
)

func init() {
	tmp := true
	FistRunning = &tmp
	Prxy := os.Getenv("PROXY")
	if Prxy != "" {
		if strings.HasPrefix(Prxy, "http") {
			urlproxy, err := url.Parse(Prxy)
			if err != nil {
				log.Error(err)
			}
			HTTPProxy = urlproxy
		} else {
			log.Warning("Invalid http proxy,format http://<addr>:<port>")
		}
	}

	Web := os.Getenv("DISCORD")
	if Web != "" {
		DiscordWebHookURL = Web
	} else {
		log.Panic("DISCORD WebHookURL not found")
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	log.Info("Starting RSS")
	StartCheck()
	c := cron.New()
	c.Start()
	c.AddFunc("@every 0h5m0s", StartCheck)

	shutdown := make(chan int)
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info("Shutting down...")
		shutdown <- 1
	}()

	<-shutdown

}

func StartCheck() {
	log.Info("Start checking sins")
	Data := Curl(EndPoint)
	err := json.Unmarshal(Data, &Payload)
	if err != nil {
		log.Error(err)
	}

	if *FistRunning {
		log.Info("First Running")
		OldPayload = Payload
		tmp := false
		FistRunning = &tmp
		return
	} else {
		for _, v := range Payload.Data {
			for _, v2 := range OldPayload.Data {
				if v.ID != v2.ID {
					log.Info("New Pic", v.URL)
					Pic, err := json.Marshal(map[string]interface{}{
						"username":   v.Username,
						"avatar_url": v.UserAvatarURL,
						"content":    v.URL,
					})
					if err != nil {
						log.Error(err)
					}

					req, err := http.NewRequest("POST", DiscordWebHookURL, bytes.NewReader(Pic))
					if err != nil {
						log.Error(err)
					}
					req.Header.Set("Content-Type", "application/json")

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						log.Error(err)
					}
					defer resp.Body.Close()
				}
			}
		}
		OldPayload = Payload
	}
}

func Curl(URL string) []byte {
	client := http.DefaultClient
	if HTTPProxy != nil {
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(HTTPProxy),
				DialContext: (&net.Dialer{
					Timeout: 10 * time.Second,
				}).DialContext,
			},
		}
	}

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		log.Error(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:62.0) Gecko/20100101 Firefox/62.0")

	result, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}

	if result.StatusCode != http.StatusOK {
		log.Error(result.Status)
	}

	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.Error((err))
	}
	return data
}

type LewdPic struct {
	Meta struct {
		Total  int         `json:"total"`
		Offset int         `json:"offset"`
		Count  int         `json:"count"`
		Error  interface{} `json:"error"`
	} `json:"meta"`
	Data []LewdPayload `json:"data"`
}

type LewdPayload struct {
	ID            int    `json:"id"`
	ChannelName   string `json:"channel_name"`
	Username      string `json:"username"`
	URL           string `json:"url"`
	ProxyURL      string `json:"proxy_url"`
	Extension     string `json:"extension"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Filesize      int    `json:"filesize"`
	CreatedAtUnix int    `json:"created_at_unix"`
	UpdatedAtUnix int    `json:"updated_at_unix"`
	DiscordUserID string `json:"discord_user_id"`
	UserAvatarURL string `json:"user_avatar_url"`
	CanonicalURL  string `json:"canonical_url"`
}
