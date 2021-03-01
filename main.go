package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var (
	HTTPProxy         *url.URL
	Payload           []Danbooru
	DiscordWebHookURL string
	FistRunning       *bool
	DisableMale       bool
	RemoveDuplicate   bool
	LewdsPic          []string
	Tags              []string
)

const (
	EndPoint = "https://147.135.4.93/posts.json?tags=order:change%20"
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
		log.Fatal("DISCORD WebHookURL not found")
	}

	tgs := os.Getenv("TAGS")
	if tgs != "" {
		Tags = strings.Split(tgs, ",")
	} else {
		log.Fatal("Tags not found")
	}

	Ml := os.Getenv("MALE")
	if Ml != "" {
		DisableMale = false
	} else {
		DisableMale = true
	}

	dpl := os.Getenv("DUPLICATE")
	if dpl != "" {
		if strings.ToLower(dpl) == "enable" || dpl == "1" {
			RemoveDuplicate = false
		} else {
			RemoveDuplicate = true
		}
	} else {
		RemoveDuplicate = true
	}

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}

func main() {
	log.Info("Starting RSS")
	StartCheck()
	c := cron.New()
	c.Start()
	c.AddFunc("@every 0h2m0s", StartCheck)

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
	Counter := 0
	for _, v := range Tags {
		log.Info("Start checking lewd ", v)
		Data, err := Curl(EndPoint + v + "&limit=20")
		if err != nil {
			log.Error(err)
			break
		}
		err = json.Unmarshal(Data, &Payload)
		if err != nil {
			log.Error(err)
		}
		if *FistRunning {
			Counter++
			log.Info("First Running")
			for _, v := range Payload {
				if v.CheckRSS() {
					v.AddNewLewd()
				}
			}

			if Counter == len(Tags) {
				tmp := false
				FistRunning = &tmp
			}

		} else {
			for _, v := range Payload {
				if v.CheckNew() && v.CheckRSS() {
					log.Info("New Pic ", v.ID)
					v.AddNewLewd()
					fixURL := strings.Replace(v.FileURL, "147.135.4.93", "danbooru.donmai.us", -1)
					Pic, err := json.Marshal(map[string]interface{}{
						"content": fixURL,
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
	}
}

func (Data Danbooru) CheckRSS() bool {
	safebutcrott, _ := regexp.MatchString("(swimsuits|lingerie|pantyshot)", Data.TagString)
	male, _ := regexp.MatchString("(male_focus|yaoi|mature_male|muscular_male)", Data.TagString)

	if DisableMale && male {
		return false
	}

	if RemoveDuplicate && Data.HasChildren {
		return false
	}

	if Data.Rating == "e" || Data.Rating == "q" || safebutcrott {
		return true
	}
	return false
}

func (Data Danbooru) CheckNew() bool {
	for _, v := range LewdsPic {
		if Data.FileURL == v {
			return false
		}
	}

	return true
}

func (Data Danbooru) AddNewLewd() {
	LewdsPic = append(LewdsPic, Data.FileURL)
}

func Curl(URL string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}
	if HTTPProxy != nil {
		client = http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(HTTPProxy),
				DialContext: (&net.Dialer{
					Timeout: 10 * time.Second,
				}).DialContext,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:62.0) Gecko/20100101 Firefox/62.0")
	req.Header.Set("Host", "danbooru.donmai.us")

	result, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if result.StatusCode != http.StatusOK {
		return nil, errors.New(result.Status)
	}

	data, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type Danbooru struct {
	ID                  int         `json:"id,omitempty"`
	CreatedAt           string      `json:"created_at"`
	UploaderID          int         `json:"uploader_id"`
	Score               int         `json:"score"`
	Source              string      `json:"source"`
	Md5                 string      `json:"md5,omitempty"`
	LastCommentBumpedAt interface{} `json:"last_comment_bumped_at"`
	Rating              string      `json:"rating"`
	ImageWidth          int         `json:"image_width"`
	ImageHeight         int         `json:"image_height"`
	TagString           string      `json:"tag_string"`
	IsNoteLocked        bool        `json:"is_note_locked"`
	FavCount            int         `json:"fav_count"`
	FileExt             string      `json:"file_ext,omitempty"`
	LastNotedAt         interface{} `json:"last_noted_at"`
	IsRatingLocked      bool        `json:"is_rating_locked"`
	ParentID            interface{} `json:"parent_id"`
	HasChildren         bool        `json:"has_children"`
	ApproverID          interface{} `json:"approver_id"`
	TagCountGeneral     int         `json:"tag_count_general"`
	TagCountArtist      int         `json:"tag_count_artist"`
	TagCountCharacter   int         `json:"tag_count_character"`
	TagCountCopyright   int         `json:"tag_count_copyright"`
	FileSize            int         `json:"file_size"`
	IsStatusLocked      bool        `json:"is_status_locked"`
	PoolString          string      `json:"pool_string"`
	UpScore             int         `json:"up_score"`
	DownScore           int         `json:"down_score"`
	IsPending           bool        `json:"is_pending"`
	IsFlagged           bool        `json:"is_flagged"`
	IsDeleted           bool        `json:"is_deleted"`
	TagCount            int         `json:"tag_count"`
	UpdatedAt           string      `json:"updated_at"`
	IsBanned            bool        `json:"is_banned"`
	PixivID             int         `json:"pixiv_id"`
	LastCommentedAt     interface{} `json:"last_commented_at"`
	HasActiveChildren   bool        `json:"has_active_children"`
	BitFlags            int         `json:"bit_flags"`
	TagCountMeta        int         `json:"tag_count_meta"`
	HasLarge            bool        `json:"has_large"`
	HasVisibleChildren  bool        `json:"has_visible_children"`
	TagStringGeneral    string      `json:"tag_string_general"`
	TagStringCharacter  string      `json:"tag_string_character"`
	TagStringCopyright  string      `json:"tag_string_copyright"`
	TagStringArtist     string      `json:"tag_string_artist"`
	TagStringMeta       string      `json:"tag_string_meta"`
	FileURL             string      `json:"file_url,omitempty"`
	LargeFileURL        string      `json:"large_file_url,omitempty"`
	PreviewFileURL      string      `json:"preview_file_url,omitempty"`
}
