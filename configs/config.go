package configs

import (
	"bytes"
	"encoding/xml"
	. "github.com/eywa/utils"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io"
	"sync/atomic"
	"text/template"
	"unsafe"
)

var cfgPtr unsafe.Pointer
var filename string
var params map[string]string

// function to update config pointer
func SetConfig(cfg *Conf) {
	atomic.StorePointer(&cfgPtr, unsafe.Pointer(cfg))
}

// return pointer to the cached config
func Config() *Conf {
	return (*Conf)(cfgPtr)
}

// read configuration from a .yml file and cache it in cfgPtr.
func ReadConfig(buf io.Reader) (*Conf, error) {
	v := viper.New()
	v.SetConfigType("yml")
	err := v.ReadConfig(buf)
	if err != nil {
		return nil, err
	}

	serviceConfig := &ServiceConf{
		Host:       v.GetString("service.host"),
		ApiPort:    v.GetInt("service.api_port"),
		DevicePort: v.GetInt("service.device_port"),
		PidFile:    v.GetString("service.pid_file"),
		Assets:     v.GetString("service.assets"),
		Templates:  v.GetString("service.templates"),
	}

	securityConfig := &SecurityConf{
		Dashboard: &DashboardSecurityConf{
			Username:    v.GetString("security.dashboard.username"),
			Password:    v.GetString("security.dashboard.password"),
			TokenExpiry: &JSONDuration{v.GetDuration("security.dashboard.token_expiry")},
			AES: &AESConf{
				KEY: v.GetString("security.dashboard.aes.key"),
				IV:  v.GetString("security.dashboard.aes.iv"),
			},
		},
		SSL: &SSLConf{
			CertFile: v.GetString("security.ssl.cert_file"),
			KeyFile:  v.GetString("security.ssl.key_file"),
		},
		ApiKey: v.GetString("security.api_key"),
	}

	dbConfig := &DbConf{
		DbType: v.GetString("database.db_type"),
		DbFile: v.GetString("database.db_file"),
	}

	indexConfig := &IndexConf{
		Disable:          v.GetBool("indices.disable"),
		Host:             v.GetString("indices.host"),
		Port:             v.GetInt("indices.port"),
		NumberOfShards:   v.GetInt("indices.number_of_shards"),
		NumberOfReplicas: v.GetInt("indices.number_of_replicas"),
		TTLEnabled:       v.GetBool("indices.ttl_enabled"),
		TTL:              &JSONDuration{v.GetDuration("indices.ttl")},
	}

	connConfig := &ConnectionsConf{
		Http: &HttpConnectionConf{
			Timeouts: &HttpConnectionTimeoutConf{
				LongPolling: &JSONDuration{v.GetDuration("connections.http.timeouts.long_polling")},
			},
		},
		Websocket: &WsConnectionConf{
			RequestQueueSize: v.GetInt("connections.websocket.request_queue_size"),
			Timeouts: &WsConnectionTimeoutConf{
				Write:    &JSONDuration{v.GetDuration("connections.websocket.timeouts.write")},
				Read:     &JSONDuration{v.GetDuration("connections.websocket.timeouts.read")},
				Request:  &JSONDuration{v.GetDuration("connections.websocket.timeouts.request")},
				Response: &JSONDuration{v.GetDuration("connections.websocket.timeouts.response")},
			},
			BufferSizes: &WsConnectionBufferSizeConf{
				Write: v.GetInt("connections.websocket.buffer_sizes.write"),
				Read:  v.GetInt("connections.websocket.buffer_sizes.read"),
			},
		},
	}

	logEywa := &LogConf{
		Filename:   v.GetString("logging.eywa.filename"),
		MaxSize:    v.GetInt("logging.eywa.maxsize"),
		MaxAge:     v.GetInt("logging.eywa.maxage"),
		MaxBackups: v.GetInt("logging.eywa.maxbackups"),
		Level:      v.GetString("logging.eywa.level"),
		BufferSize: v.GetInt("logging.eywa.buffer_size"),
	}

	logIndices := &LogConf{
		Filename:   v.GetString("logging.indices.filename"),
		MaxSize:    v.GetInt("logging.indices.maxsize"),
		MaxAge:     v.GetInt("logging.indices.maxage"),
		MaxBackups: v.GetInt("logging.indices.maxbackups"),
		Level:      v.GetString("logging.indices.level"),
		BufferSize: v.GetInt("logging.indices.buffer_size"),
	}

	logDatabase := &LogConf{
		Filename:   v.GetString("logging.database.filename"),
		MaxSize:    v.GetInt("logging.database.maxsize"),
		MaxAge:     v.GetInt("logging.database.maxage"),
		MaxBackups: v.GetInt("logging.database.maxbackups"),
		Level:      v.GetString("logging.database.level"),
		BufferSize: v.GetInt("logging.database.buffer_size"),
	}

	cfg := &Conf{
		Service:     serviceConfig,
		Security:    securityConfig,
		Connections: connConfig,
		Indices:     indexConfig,
		Database:    dbConfig,
		Logging: &LogsConf{
			Eywa:     logEywa,
			Indices:  logIndices,
			Database: logDatabase,
		},
	}

	return cfg, nil
}

// update the config cache
func Update(settings map[string]interface{}) error {
	_cfg, err := Config().DeepCopy()
	if err != nil {
		return err
	}

	err = Assign(_cfg, settings, map[string]AssignReader{"jsonduration": JSONDurationAssignReader})
	if err != nil {
		return err
	}

	SetConfig(_cfg)
	return nil
}

func InitializeConfig(f string, p map[string]string) error {
	filename = f
	params = p

	// get default config
	buf := bytes.NewBuffer([]byte{})
	t, err := template.New("defaults").Parse(DefaultConfigs)
	if err != nil {
		return err
	}
	err = t.Execute(buf, params)
	if err != nil {
		return err
	}

	_cfg, err := ReadConfig(buf)
	if err != nil {
		return err
	}

	// get custom config
	t, err = template.ParseFiles(filename)
	if err != nil {
		return err
	}

	buf = bytes.NewBuffer([]byte{})
	err = t.Execute(buf, params)
	if err != nil {
		return err
	}

	s := map[interface{}]interface{}{}
	err = yaml.Unmarshal(buf.Bytes(), &s)
	if err != nil {
		return err
	}
	strMap, err := ToStringMap(s)
	if err != nil {
		return err
	}

	err = ForceAssign(_cfg, strMap, map[string]AssignReader{"jsonduration": JSONDurationAssignReader})
	if err != nil {
		return err
	}

	SetConfig(_cfg)
	return nil
}

// copy every field in the config cache and returns a new cfg pointer to
// the new copy. The old figure cache will be garbage collected. 
func (cfg *Conf) DeepCopy() (*Conf, error) {
	// The reason of not using json decoder is that json decoder ignores tags
	// with `json:"-"`. So it will miss some fields which is not what DeepCopy
	// wants. Instead, xml decoder marshals every fields.
	asBytes, err := xml.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	_cfg := &Conf{}
	err = xml.Unmarshal(asBytes, _cfg)
	if err != nil {
		return nil, err
	}

	return _cfg, nil
}

type Conf struct {
	Service     *ServiceConf     `json:"service" assign:"service;;-"`
	Security    *SecurityConf    `json:"security" assign:"security;;"`
	Connections *ConnectionsConf `json:"connections" assign:"connections;;"`
	Indices     *IndexConf       `json:"indices" assign:"indices;;"`
	Database    *DbConf          `json:"database" assign:"database;;-"`
	Logging     *LogsConf        `json:"logging" assign:"logging;;-"`
}

type DbConf struct {
	DbType string `json:"db_type" assign:"db_type;;-"`
	DbFile string `json:"db_file" assign:"db_file;;-"`
}

type IndexConf struct {
	Disable          bool          `json:"disable" assign:"disable;;"`
	Host             string        `json:"host" assign:"host;;-"`
	Port             int           `json:"port" assign:"port;;-"`
	NumberOfShards   int           `json:"number_of_shards" assign:"number_of_shards;;-"`
	NumberOfReplicas int           `json:"number_of_replicas" assign:"number_of_replicas;;-"`
	TTLEnabled       bool          `json:"ttl_enabled" assign:"ttl_enabled;;-"`
	TTL              *JSONDuration `json:"ttl" assign:"ttl;jsonduration;-"`
}

type ServiceConf struct {
	Host       string `json:"host" assign:"host;;-"`
	ApiPort    int    `json:"api_port" assign:"api_port;;-"`
	DevicePort int    `json:"device_port" assign:"device_port;;-"`
	PidFile    string `json:"-" assign:"pid_file;;-"`
	Assets     string `json:"-" assign:"assets;;-"`
	Templates  string `json:"-" assign:"templates;;-"`
}

type ConnectionsConf struct {
	Http      *HttpConnectionConf `json:"http" assign:"http;;"`
	Websocket *WsConnectionConf   `json:"websocket" assign:"websocket;;"`
}

type HttpConnectionConf struct {
	Timeouts *HttpConnectionTimeoutConf `json:"timeouts" assign:"timeouts;;"`
}

type HttpConnectionTimeoutConf struct {
	LongPolling *JSONDuration `json:"long_polling" assign:"long_polling;jsonduration;"`
}

type WsConnectionConf struct {
	RequestQueueSize int                         `json:"request_queue_size" assign:"request_queue_size;;"`
	Timeouts         *WsConnectionTimeoutConf    `json:"timeouts" assign:"timeouts;;"`
	BufferSizes      *WsConnectionBufferSizeConf `json:"buffer_sizes" assign:"buffer_sizes;;"`
}

type WsConnectionTimeoutConf struct {
	Write    *JSONDuration `json:"write" assign:"write;jsonduration;"`
	Read     *JSONDuration `json:"read" assign:"read;jsonduration;"`
	Request  *JSONDuration `json:"request" assign:"request;jsonduration;"`
	Response *JSONDuration `json:"response" assign:"response;jsonduration;"`
}

type WsConnectionBufferSizeConf struct {
	Write int `json:"write" assign:"write;;"`
	Read  int `json:"read" assign:"read;;"`
}

type LogsConf struct {
	Eywa     *LogConf `json:"eywa" assign:"eywa;;-"`
	Indices  *LogConf `json:"indices" assign:"indices;;-"`
	Database *LogConf `json:"database" assign:"database;;-"`
}

type LogConf struct {
	Filename   string `json:"filename" assign:"filename;;-"`
	MaxSize    int    `json:"maxsize" assign:"maxsize;;-"`
	MaxAge     int    `json:"maxage" assign:"maxage;;-"`
	MaxBackups int    `json:"maxbackups" assign:"maxbackups;;-"`
	Level      string `json:"level" assign:"level;;-"`
	BufferSize int    `json:"buffer_size" assign:"buffer_size;;-"`
}

type SecurityConf struct {
	Dashboard *DashboardSecurityConf `json:"dashboard" assign:"dashboard;;"`
	SSL       *SSLConf               `json:"-" assign:"ssl;;-"`
	ApiKey    string                 `json:"-" assign:"api_key;;"`
}

type DashboardSecurityConf struct {
	Username    string        `json:"username" assign:"username;;"`
	Password    string        `json:"-" assign:"password;;"`
	TokenExpiry *JSONDuration `json:"token_expiry" assign:"token_expiry;jsonduration;"`
	AES         *AESConf      `json:"-" assign:"aes;;-"`
}

type AESConf struct {
	KEY string `json:"key" assign:"key;;-"`
	IV  string `json:"iv" assign:"iv;;-"`
}

type SSLConf struct {
	CertFile string `json:"cert_file" assign:"cert_file;;-"`
	KeyFile  string `json:"key_file" assign:"key_file;;-"`
}
