package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/parnurzeal/gorequest"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/eywa/utils"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ulimit -n 1048576; go run tasks/benchmark.go  -host=<host> -ports=8080:8081 -user=root -passwd=waterISwide -fields=temperature:float -c=20000 -p=5 -m=5 -r=300s -w=10s -i=20000 -I=3 > bench.log 2>&1 &

type Dialer struct {
	counter uint64
	dialers []*websocket.Dialer
}

func (d *Dialer) Dial(urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error) {
	c := atomic.AddUint64(&d.counter, 1)
	return d.dialers[c%uint64(len(d.dialers))].Dial(urlStr, requestHeader)
}

func main() {
	host := flag.String("host", "localhost", "the target server host")
	ports := flag.String("ports", "8080:8081", "the http port and device port")
	fields := flag.String("fields", "temperature:float", "fields that are used for bench test. Format: 'field1:type1,field2:type2'")
	user := flag.String("user", "root", "username for authenticating eywa")
	passwd := flag.String("passwd", "waterISwide", "passwd for authenticating eywa")

	c := flag.Int("c", 1000, "number of concurrent clients")
	p := flag.Int("p", 100, "number of ping messages to send")
	m := flag.Int("m", 50, "number of payload messages to send")
	r := flag.Duration("r", 15*time.Second, "wait time for reading messages")
	w := flag.Duration("w", 2*time.Second, "wait time for writing messages")
	i := flag.Int("i", 5000, "wait milliseconds interval between each sends in client, randomized")
	I := flag.Int("I", 1000, "wait milliseconds interval between each connection, randomized")
	b := flag.String("b", "", "ip addresses used to bind clients, defaults to localhost")
	s := flag.Duration("s", 10*time.Second, "the sleep time after messages are all set for each client")

	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	dialers := make([]*websocket.Dialer, 0)
	if len(*b) == 0 {
		dialers = append(dialers, websocket.DefaultDialer)
	} else {
		_ips := strings.Split(*b, ",")
		for _, _ip := range _ips {
			ip, err := net.ResolveIPAddr("ip4", strings.Trim(_ip, " "))
			if err != nil {
				log.Fatalf("%s is not a valid IPv4 address.\n", _ip)
			}

			localTCPAddr := &net.TCPAddr{
				IP: ip.IP,
			}

			dialers = append(dialers, &websocket.Dialer{
				Proxy: http.ProxyFromEnvironment,
				NetDial: (&net.Dialer{
					LocalAddr: localTCPAddr,
				}).Dial,
			})
		}
	}

	if len(dialers) == 0 {
		log.Fatalln("none of the localAddr's are valid")
	}

	dialer := &Dialer{dialers: dialers}

	_ports := strings.Split(*ports, ":")
	if len(_ports) != 2 {
		log.Fatalln("Invalid ports format, expecting <http port>:<device port>.")
	}
	httpPort := _ports[0]
	devicePort := _ports[1]

	log.Println("Login the eywa and get the auth token...")
	url := fmt.Sprintf("http://%s:%s/login", *host, httpPort)
	req := gorequest.New()
	response, bodyBytes, errs := req.Get(url).SetBasicAuth(*user, *passwd).EndBytes()
	if len(errs) > 0 {
		log.Fatalln(errs[0].Error())
	}
	if response.StatusCode != 200 {
		log.Fatalln("Unable to authenticate to Eywa. Please check the user/passwd pair.")
	}
	var loggedIn map[string]string
	err := json.Unmarshal(bodyBytes, &loggedIn)
	if err != nil {
		log.Fatalln("Unable to get auth response")
	}
	auth := loggedIn["auth_token"]
	if len(auth) > 0 {
		log.Println("Successfully logged in.")
	} else {
		log.Fatalln("Unable to get auth token, please check the server log.")
	}

	log.Println("Creating a channel for testing...")
	chanName := fmt.Sprintf("bench_channel_%d", time.Now().UnixNano())
	token := "123456789"
	url = fmt.Sprintf("http://%s:%s/channels", *host, httpPort)
	fieldDefs := strings.Split(*fields, ",")
	fieldMap := make(map[string]string)
	for _, def := range fieldDefs {
		pair := strings.Split(def, ":")
		fieldMap[pair[0]] = pair[1]
	}
	reqbody := map[string]interface{}{
		"name":          chanName,
		"description":   "bench test channel",
		"fields":        fieldMap,
		"access_tokens": []string{token},
	}
	asBytes, err := json.Marshal(reqbody)
	if err != nil {
		log.Fatalln(err.Error())
	}

	req = gorequest.New()
	response, bodyBytes, errs = req.Post(url).Set("AuthToken", auth).
		Send(string(asBytes)).EndBytes()
	if len(errs) > 0 {
		log.Fatalln(errs[0].Error())
	}
	if response.StatusCode != 201 {
		log.Fatalln("Unable to create test channel. Please check server log.")
	}

	var created map[string]string
	err = json.Unmarshal(bodyBytes, &created)
	if err != nil {
		log.Fatalln("Unable to get channel creation response")
	}
	chId := created["id"]
	if len(chId) > 0 {
		log.Println("Successfully created channel.")

		defer func() {
			log.Println("Deleting test channel...")
			req = gorequest.New()
			url = fmt.Sprintf("http://%s:%s/channels/%s", *host, httpPort, chId)
			_, _, errs = req.Delete(url).Set("AuthToken", auth).End()
			if len(errs) > 0 {
				log.Fatalln(errs[0].Error())
			}
			log.Println("Successfully deleted test channel.")
		}()

	} else {
		log.Fatalln("Unable to get created channel Id. Please check server log.")
	}

	log.Println("Starting clients...")
	clients := make([]*WsClient, *c)
	var wg sync.WaitGroup
	wg.Add(*c)

	for _i := 0; _i < *c; _i++ {
		time.Sleep(time.Duration(rand.Intn(*I)) * time.Millisecond)
		go func(idx int) {
			defer wg.Done()
			c := &WsClient{
				Dialer:      dialer,
				Server:      fmt.Sprintf("%s:%s", *host, devicePort),
				ChannelId:   chId,
				DeviceId:    fmt.Sprintf("device-%d-%d", idx, time.Now().UnixNano()),
				AccessToken: token,
				NPing:       *p,
				NMessage:    *m,
				RWait:       *r,
				WWait:       *w,
				Itv:         *i,
				ch:          make(chan struct{}),
				fields:      fieldMap,
				Sleep:       *s,
			}

			clients[idx] = c
			c.StartTest()
		}(_i)
	}

	log.Println("Waiting for clients to complete...")
	wg.Wait()

	log.Println("collecting test results...")
	report := make(map[string]interface{})
	report["total_clients"] = *c

	var connErrs int
	var pingErrs int
	var pings int
	var msgs int
	var pongs int
	var closeErrs int
	var msgErrs int
	var msgSent int
	var pingSent int

	for _, cli := range clients {
		pings += cli.NPing
		msgs += cli.NMessage
		pongs += cli.Pongs
		msgErrs += cli.MessageErr
		pingErrs += cli.PingErr
		msgSent += cli.MessageSent
		pingSent += cli.PingSent

		if cli.ConnErr != nil {
			connErrs += 1
		}

		if cli.MessageCloseErr != nil {
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

	fmt.Println("******************************************************************")
	js, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(js))
	fmt.Println("******************************************************************")
}

type WsClient struct {
	Dialer      *Dialer
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
	Sleep           time.Duration
}

func (c *WsClient) StartTest() {
	defer func() {
		log.Printf("devices %s completed.\n", c.DeviceId)
	}()

	p := fmt.Sprintf("/ws/channels/%s/devices/%s", c.ChannelId, c.DeviceId)
	u := url.URL{Scheme: "ws", Host: c.Server, Path: p}
	h := map[string][]string{"AccessToken": []string{c.AccessToken}}

	cli, resp, err := c.Dialer.Dial(u.String(), h)
	c.ConnErr = err
	c.ConnResp = resp
	c.Cli = cli

	if err != nil {
		log.Println(err.Error())
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
			FatalIfErr(err)
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

	time.Sleep(c.Sleep)

	cli.SetWriteDeadline(time.Now().Add(c.WWait))
	err = cli.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		c.MessageCloseErr = err
	}

}
