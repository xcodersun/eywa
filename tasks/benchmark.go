package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/parnurzeal/gorequest"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/octopus/utils"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"time"
)

//go run tasks/benchmark.go  -h=107.170.239.173 -p=8080 -w=8081 -n=1 -ch=bench_test -tk=1234567 -fs=temperature:float -np=1 -nm=1 -auth=M_MPDzpPSUuYt_CLdeeMKRBYRpVc-IhTUNXxvKqSi5Xa8zRNyz6_rEaQKWnlk-vDQePYSVcOQiaA6dw283C1yHdJ71NzEAePND3OkA-3p8FcMwSKF8UUmY_y3gLO_dJPvznWlL0LI70EA4lh8xtyKjTMhK1qF96YYRJCG_Q7BLi9r3kJRbN0Hb4OBBy-gVyNrGAKkjphMOBpfpKXkMgZkU6L41rGodZfNwX6lt1AKO0ZWiGPKuujLdQIuPlFR3axyWHDIUF6k56pKy-NFUMoQ6Kxwj1hunMEi68YpkVTxCDqHYZ7xkrq-IBKsbrgEX0Nv9VvSkVDMOIREOCDkoKSkPOBsEDTUe1OdL-lUYDtegVgN3jDW1Qmvjts8LzvuLprmpuIToxeBHbH9KZebxGD2dcyG9hN9sWVszv0JdPCaZiGRQjGZk4Adwbi2tqNx06jHNIOokM7Mbbyk0L_LTC9O8YdzqoLnLnp-MuWOeKVTuyZB3LyoA_Vpxv--y88jWw7ySEQihVyoTb9F9zyAlBa-OTxcTjSzU0C0fvsUeM2Z525re0q9Ek6MswNKjSiow==

func main() {
	h := flag.String("h", "localhost", "the target server host")
	p := flag.String("p", "8080", "the http port")
	w := flag.String("w", "8081", "the ws port")
	n := flag.Int("n", 1000, "number of concurrent clients")
	ch := flag.String("ch", "test", "channel name for testing")
	tk := flag.String("tk", "1234567", "access token used for testing")
	fs := flag.String("fs", "temperature:float", "fields that are used for bench test. Format: 'field1:type1,field2:type2'")
	np := flag.Int("np", 100, "number of ping messages to send")
	nm := flag.Int("nm", 50, "number of payload messages to send")
	rw := flag.Duration("rw", 15*time.Second, "wait time for reading messages")
	ww := flag.Duration("ww", 2*time.Second, "wait time for writing messages")
	itv := flag.Int("i", 5000, "wait milliseconds interval between each sends in client, randomized")
	citv := flag.Int("I", 1000, "wait milliseconds interval between each connection, randomized")
	auth := flag.String("auth", "", "auth_token for creating channel")

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	//create a channel for testing
	url := fmt.Sprintf("http://%s:%s/channels", *h, *p)
	fieldDefs := strings.Split(*fs, ",")
	fields := make(map[string]string)
	for _, def := range fieldDefs {
		pair := strings.Split(def, ":")
		fields[pair[0]] = pair[1]
	}
	body := map[string]interface{}{
		"name":          *ch,
		"description":   "bench test channel",
		"fields":        fields,
		"access_tokens": []string{*tk},
	}
	asBytes, err := json.Marshal(body)
	PanicIfErr(err)
	req := gorequest.New()

	_, bodyBytes, errs := req.Post(url).Set("AuthToken", *auth).
		Send(string(asBytes)).EndBytes()
	if len(errs) > 0 {
		PanicIfErr(errs[0])
	}

	var created map[string]interface{}
	json.Unmarshal(bodyBytes, &created)
	chId := created["id"].(string)

	//start clients
	clients := make([]*WsClient, *n)
	var wg sync.WaitGroup
	wg.Add(*n)

	for i := 0; i < *n; i++ {
		time.Sleep(time.Duration(rand.Intn(*citv)) * time.Millisecond)
		go func(idx int) {
			defer wg.Done()
			c := &WsClient{
				Server:      fmt.Sprintf("%s:%s", *h, *w),
				ChannelId:   chId,
				DeviceId:    fmt.Sprintf("device-%d", idx),
				AccessToken: *tk,
				NPing:       *np,
				NMessage:    *nm,
				RWait:       *rw,
				WWait:       *ww,
				Itv:         *itv,
				ch:          make(chan struct{}),
				fields:      fields,
			}

			clients[idx] = c
			c.StartTest()
		}(i)
	}

	wg.Wait()

	//collecting test results
	report := make(map[string]interface{})
	report["total_clients"] = *n

	var connErrs int
	var pingErrs int
	var pings int
	var msgs int
	var pongs int
	var closeErrs int
	var msgErrs int
	var msgSent int
	var pingSent int

	for _, c := range clients {
		pings += c.NPing
		msgs += c.NMessage
		pongs += c.Pongs
		msgErrs += c.MessageErr
		pingErrs += c.PingErr
		msgSent += c.MessageSent
		pingSent += c.PingSent

		if c.ConnErr != nil {
			connErrs += 1
		}

		if c.MessageCloseErr != nil {
			closeErrs += 1
		}
	}

	report["total_conn_errs"] = connErrs
	report["total_ping_errs"] = pingErrs
	report["total_close_errs"] = closeErrs
	report["total_pings"] = pings
	report["total_pongs"] = pongs
	report["total_msgs"] = msgs
	report["total_msg_errs"] = msgErrs
	report["total_msg_sent"] = msgSent
	report["total_ping_sent"] = pingSent

	js, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(js))
}

type WsClient struct {
	Server      string
	ChannelId   string
	DeviceId    string
	AccessToken string
	NPing       int
	NMessage    int
	RWait       time.Duration
	WWait       time.Duration
	Itv         int
	wg          sync.WaitGroup
	ch          chan struct{}
	fields      map[string]string

	Cli             *websocket.Conn
	ConnErr         error
	ConnResp        *http.Response
	PingErr         int
	Pongs           int
	MessageErr      int
	MessageCloseErr error
	MessageSent     int
	PingSent        int
}

func (c *WsClient) StartTest() {
	p := fmt.Sprintf("/ws/channels/%s/devices/%s", c.ChannelId, c.DeviceId)
	u := url.URL{Scheme: "ws", Host: c.Server, Path: p}
	h := map[string][]string{"AccessToken": []string{c.AccessToken}}

	cli, resp, err := websocket.DefaultDialer.Dial(u.String(), h)
	c.ConnErr = err
	c.ConnResp = resp
	c.Cli = cli

	if err != nil {
		return
	}

	cli.SetPongHandler(func(string) error {
		c.Pongs += 1
		return nil
	})
	c.wg.Add(2)

	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-c.ch:
				return
			default:
				cli.SetReadDeadline(time.Now().Add(c.RWait))
				cli.ReadMessage()
			}
		}

	}()

	go func() {
		defer c.wg.Done()

		n := 0
		m := 0

		for n < c.NPing || m < c.NMessage {
			cli.SetWriteDeadline(time.Now().Add(c.WWait))
			msgBody := map[string]interface{}{}
			for f, t := range c.fields {
				switch t {
				case "float":
					msgBody[f] = rand.Float32()
				case "int":
					msgBody[f] = rand.Int31()
				case "boolean":
					msgBody[f] = true
				case "string":
					msgBody[f] = uuid.NewV1().String()
				default:
					msgBody[f] = uuid.NewV1().String()
				}
			}
			asBytes, err := json.Marshal(msgBody)
			PanicIfErr(err)
			if n >= c.NPing {
				err := cli.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("1|%d|%s", rand.Int31(), string(asBytes))))
				m += 1
				if err != nil {
					c.MessageErr += 1
				}
			} else if m >= c.NMessage {
				err := cli.WriteMessage(websocket.PingMessage, []byte{})
				n += 1
				if err != nil {
					c.PingErr += 1
				}
			} else {
				r := rand.Intn(2)
				if r == 0 {
					err := cli.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("1|%d|%s", rand.Int31(), string(asBytes))))
					m += 1
					if err != nil {
						c.MessageErr += 1
					}
				} else {
					err := cli.WriteMessage(websocket.PingMessage, []byte{})
					n += 1
					if err != nil {
						c.PingErr += 1
					}
				}
			}
			time.Sleep(time.Duration(c.Itv) * time.Millisecond)
		}

		c.MessageSent = m
		c.PingSent = n

		time.Sleep(3 * time.Second)
		close(c.ch)

	}()

	c.wg.Wait()

	cli.SetWriteDeadline(time.Now().Add(c.WWait))
	err = cli.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		c.MessageCloseErr = err
	}

}
