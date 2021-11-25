package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/zserge/lorca"

	"github.com/hugolgst/rich-go/client"
)

//go:embed www
var fs embed.FS

type Song struct {
	Song SongStruct `json:"song"`
}

type SongStruct struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

func request() string {
	resp, err := http.Get("https://upbeatradio.net/api/v1/stats")
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	s := Song{}
	json.Unmarshal(body, &s)

	if err != nil {
		log.Fatal(err)
	}
	return s.Song.Title
}

func doEvery(d time.Duration, f func(time.Time)) {
	for x := range time.Tick(d) {
		f(x)
	}
}

func createActivity(t time.Time) {
	m1 := request()
	err := client.Login("781004595546816522")

	if err != nil {
		panic(err)
	}
	now := time.Now()
	println(m1)
	err = client.SetActivity(client.Activity{
		State:      "UpBeatRadio",
		Details:    `Listening to ` + m1,
		LargeImage: "upb",
		SmallImage: "ups",
		LargeText:  "UpBeatRadio Desktop",
		SmallText:  "UpBeatRadio Desktop",
		Timestamps: &client.Timestamps{
			Start: &now,
		},
		Buttons: []*client.Button{
			{
				Label: "Listen",
				Url:   "https://live.upbeat.pw",
			},
			{
				Label: "Website",
				Url:   "https://upbeat.pw",
			},
		},
	})

	if err != nil {
		panic(err)
	}
}
func main() {
	args := []string{}
	if runtime.GOOS == "linux" {
		args = append(args, "--class=Lorca")
	}
	ui, err := lorca.New("", "", 318, 500, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()
	ui.Bind("start", func() {
		log.Println("UI is ready")
	})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(http.FS(fs)))
	ui.Load(fmt.Sprintf("http://%s/www", ln.Addr()))
	ui.Eval(`
		console.log("Hello, world!");
		console.log('Multiple values:', [1, false, {"x":5}]);
	`)

	if err != nil {
		panic(err)
	}
	doEvery(120*time.Second, createActivity)

	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	log.Println("exiting...")
}
