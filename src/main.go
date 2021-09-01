package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	// "strconv"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-yaml/yaml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/pkg/errors"
)

const natureRemoTokenPath = "./nature_remo_token.txt"
const dbInfoPath = "./dbinfo.yml"

type user struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	Superuser bool   `json:"superuser"`
}

type illuminanceValue struct {
	Val       float32   `json:"val"`
	CreatedAt time.Time `json:"created_at"`
}

type humidityValue struct {
	Val       float32   `json:"val"`
	CreatedAt time.Time `json:"created_at"`
}

type movedValue struct {
	Val       float32   `json:"val"`
	CreatedAt time.Time `json:"created_at"`
}

type temperatureValue struct {
	Val       float32   `json:"val"`
	CreatedAt time.Time `json:"created_at"`
}

type events struct {
	Hu humidityValue    `json:"hu"`
	Il illuminanceValue `json:"il"`
	Mo movedValue       `json:"mo"`
	Te temperatureValue `json:"te"`
}

type remoDevice struct {
	Name              string    `json:"name"`
	ID                string    `json:"id"`
	CratedAt          time.Time `json:"crated_at"`
	MacAddress        string    `json:"mac_address"`
	SerialNumber      string    `json:"serial_number"`
	FirmwareVersion   string    `json:"firmware_version"`
	TemperatureOffset int       `json:"temperature_offset"`
	HumidityOffset    int       `json:"humidity_offset"`
	Users             []user    `json:"users"`
	NewestEvents      events    `json:"newest_events"`
}

type humidity struct {
	Value float32
	Time  time.Time
}
type temperature struct {
	Value float32
	Time  time.Time
}
type illuminance struct {
	Value float32
	Time  time.Time
}
type moved struct {
	Value float32
	Time  time.Time
}

type sbothumidity struct {
	Value int
	Time  time.Time
}
type sbottemperature struct {
	Value float32
	Time  time.Time
}

type dbinfo struct {
	Dbms     string `yaml:"dbms"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	Protocol string `yaml:"protocol"`
	Dbname   string `yaml:"dbname"`
	Param    string `yaml:"param"`
}

type myCo2SensorRest struct {
	Val       int   `json:"co2"`
}

type co2ppm struct {
	Value       int 
	Time time.Time
}

func gormConnect() *gorm.DB {
	f, err := os.Open(dbInfoPath)
	if err != nil {
		log.Fatal("Failed to open dbinfo from file.", err)
		panic(err.Error())
	}
	defer f.Close()

	var dbinfos dbinfo
	err = yaml.NewDecoder(f).Decode(&dbinfos)
	println("dbinfo.user=" + dbinfos.User)

	CONNECT := dbinfos.User + ":" + dbinfos.Pass + "@" + dbinfos.Protocol + "/" + dbinfos.Dbname + "?" + dbinfos.Param
	db, err := gorm.Open(dbinfos.Dbms, CONNECT)

	if err != nil {
		log.Fatal("Failed to connect to database by gorm.", err)
		panic(err.Error())
	}
	return db
}

/*RemoOperator is contorller class of nature remo*/
type RemoOperator struct {
	token          string
	db             *gorm.DB
	humidityOrm    humidity
	illuminanceOrm illuminance
	temperatureOrm temperature
	movedOrm       moved
}

func (r *RemoOperator) readToken(tokenPath string) error {
	//remoのトークン取得/接続準備
	fp, err := os.Open(tokenPath)
	if err != nil {
		log.Fatal("Failed to open nature remo token text file")
		return err
	}
	defer fp.Close()

	reader := bufio.NewReader(fp)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Failed to read nature remo token from text")
			return err
		}

		r.token = string(line)
		fmt.Println("Succeeded to read remo token!")
	}
	return nil
}

func (r *RemoOperator) retrieveSensorValue() error {
	client := http.Client{}
	url := "https://api.nature.global/1/devices"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+string(r.token))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed to access to remo")
		return err
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)

	var devices []remoDevice
	if err := json.Unmarshal(byteArray, &devices); err != nil {
		log.Fatal("Failed to parse json from remo")
		return err
	}
	fmt.Println("Retrieved sensor value at" + time.Now().Format("2006/01/02 15:04:05"))

	// for _, d := range devices {
	// 	fmt.Print("il:createdAt = ")
	// 	fmt.Println(d.NewestEvents.Il.CreatedAt.Local())
	// }

	r.humidityOrm.Value = devices[0].NewestEvents.Hu.Val
	r.humidityOrm.Time = devices[0].NewestEvents.Hu.CreatedAt

	r.illuminanceOrm.Value = devices[0].NewestEvents.Il.Val
	r.illuminanceOrm.Time = devices[0].NewestEvents.Il.CreatedAt

	r.temperatureOrm.Value = devices[0].NewestEvents.Te.Val
	r.temperatureOrm.Time = devices[0].NewestEvents.Te.CreatedAt

	r.movedOrm.Value = devices[0].NewestEvents.Mo.Val
	r.movedOrm.Time = devices[0].NewestEvents.Mo.CreatedAt

	return nil
}

func (r *RemoOperator) writeSensoreValue() error {
	r.db.Save(&r.humidityOrm)
	r.db.Save(&r.illuminanceOrm)
	r.db.Save(&r.temperatureOrm)
	r.db.Save(&r.movedOrm)

	return nil
}

// SwitchBotReceiver is parser of switchbot sensor.
type SwitchBotReceiver struct {
	humidity    int
	temperature float32
	retrievedAt time.Time
	mu          sync.Mutex
	db          *gorm.DB
}

func (s *SwitchBotReceiver) setValues(temperature float32, humidity int) {
	s.mu.Lock()
	s.humidity = humidity
	s.temperature = temperature
	s.retrievedAt = time.Now()
	s.mu.Unlock()
}

func (s *SwitchBotReceiver) getValues() (temperature float32, humidity int, retrievedAt time.Time) {
	s.mu.Lock()
	humidity = s.humidity
	temperature = s.temperature
	retrievedAt = s.retrievedAt
	s.mu.Unlock()
	return
}

// Co2 Operator class
type co2Operator struct {
	db             *gorm.DB
	co2Orm    co2ppm
}

func (c *co2Operator) retrieveSensorValue() error {
	client := http.Client{}
	url := "http://192.168.0.116/co2"
	req, err := http.NewRequest("GET", url, nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed to access to remo")
		return err
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)

	// fmt.Println(string(byteArray))

	var myCo2Sensor myCo2SensorRest
	if err := json.Unmarshal(byteArray, &myCo2Sensor); err != nil {
		log.Fatal("Failed to parse json from co2 sensor")
		return err
	}
	// fmt.Println("Retrieved co2 value : " + strconv.Itoa(myCo2Sensor.Val))

	c.co2Orm.Value = myCo2Sensor.Val
	c.co2Orm.Time =time.Now() 

	return nil
}

func (c *co2Operator) writeSensoreValue() error {
	c.db.Save(&c.co2Orm)

	return nil
}

// intrrupting funcs
func retrieveFromNatureRemoTicker(remoOperator *RemoOperator, interval int) {
	t := time.NewTicker(time.Duration(interval) * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := remoOperator.retrieveSensorValue()
			if err != nil {
				continue
			}
			remoOperator.writeSensoreValue()
		}
	}

}

var switchBotReceiver SwitchBotReceiver

func retrieveFromSBotSensorTicker(interval int) {
	t := time.NewTicker(time.Duration(interval) * time.Minute)
	defer t.Stop()

	d, err := dev.NewDevice("default")
	if err != nil {
		log.Print("Failed to connect to bluetooth device")
		panic(err)
	}
	ble.SetDefaultDevice(d)

	sbotHumidityOrm := sbothumidity{}
	sbotTemperatureOrm := sbottemperature{}
	for {
		select {
		case <-t.C:
			//10秒ほど待てば1回は取得できる．
			ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 10*time.Second))

			//ブロッキング処理になる．
			chkErr(ble.Scan(ctx, true, advHandler, advFilter))

			// remoOperator.retrieveSensorValue()
			temperatureValue, humidityValue, retrievedAt := switchBotReceiver.getValues()

			if humidityValue == 0 || temperatureValue == 0 {
				fmt.Println("zero!!")
				continue
			}

			fmt.Printf("temperature=%2f\n", temperatureValue)
			fmt.Printf("humidity=%d\n", humidityValue)

			//Update values if it is different from last value.
			if sbotHumidityOrm.Value != humidityValue {
				sbotHumidityOrm.Value = humidityValue
				sbotHumidityOrm.Time = retrievedAt
			}
			if sbotTemperatureOrm.Value != temperatureValue {
				sbotTemperatureOrm.Value = temperatureValue
				sbotTemperatureOrm.Time = retrievedAt
			}

			//Update values if it is different from lastvalue(If it is same date, fails to update because of unique key error).
			switchBotReceiver.db.Save(&sbotHumidityOrm)
			switchBotReceiver.db.Save(&sbotTemperatureOrm)
		}
	}

}

func advFilter(a ble.Advertisement) bool {
	if strings.EqualFold(a.Addr().String(), "E1:EC:E9:82:8F:60") {
		return true
	}
	return false
}

func advHandler(a ble.Advertisement) {
	if a.Connectable() {
		fmt.Printf("[%s] C %3d:", a.Addr(), a.RSSI())
	} else {
		fmt.Printf("[%s] N %3d:", a.Addr(), a.RSSI())
	}
	comma := ""
	if len(a.LocalName()) > 0 {
		fmt.Printf(" Name: %s", a.LocalName())
		comma = ","
	}
	if len(a.Services()) > 0 {
		fmt.Printf("%s Svcs: %v", comma, a.Services())
		comma = ","
	}
	if len(a.ManufacturerData()) > 0 {
		fmt.Printf("%s MD: %X", comma, a.ManufacturerData())
	}
	fmt.Println()
	var temperature float32
	var humidity int
	for _, s := range a.ServiceData() {
		// fmt.Println(s.UUID.String())
		for i, d := range s.Data {
			// fmt.Printf("%2x", d)
			// fmt.Println()
			switch i {
			case 0:
			case 1:
			case 2:
			case 3:
				temperature = float32(int(d&(0x0f))) / 10
			case 4:
				temperature += float32(int(d &^ (1 << 7)))
				if d&(1<<7) == 0 {
					temperature *= -1
				}
			case 5:
				humidity = int(d &^ (1 << 7))
			}
		}
	}
	if temperature != 0 && humidity != 0 {
		fmt.Printf("temp=%f\n", temperature)
		fmt.Printf("hum=%d\n", humidity)
		switchBotReceiver.setValues(temperature, humidity)
	}
}

func retrieveFromMyCo2SensorTicker(_co2Operator *co2Operator, interval int) {
	t := time.NewTicker(time.Duration(interval) * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := _co2Operator.retrieveSensorValue()
			if err != nil {
				continue
			}
			_co2Operator.writeSensoreValue()
		}
	}
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		// fmt.Printf("done\n")
	case context.Canceled:
		fmt.Printf("canceled\n")
	default:
		log.Fatalf(err.Error())
	}
}

func main() {
	//DBアクセス準備
	db := gormConnect()
	defer db.Close()
	db.SingularTable(true) //テーブル名とtype名を一致させる(複数形をやめる)

	remoOperator := RemoOperator{db: db}
	err := remoOperator.readToken(natureRemoTokenPath)
	if err != nil {
		panic(err)
	}
	go retrieveFromNatureRemoTicker(&remoOperator, 5)

	switchBotReceiver.db = db
	go retrieveFromSBotSensorTicker(5)

	_co2Operator := co2Operator{db: db}
	go retrieveFromMyCo2SensorTicker(&_co2Operator, 5)

	for {
	}
}
