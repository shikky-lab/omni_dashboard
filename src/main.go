package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-yaml/yaml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"gobot.io/x/gobot/platforms/ble"
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
	ID    int `gorm:"cloumn:id"` //タグつけるとcloumn名と紐づけ．なけらばField名が使われる．
	Value float32
	Time  time.Time
}
type temperature struct {
	ID    int `gorm:"cloumn:id"` //タグつけるとcloumn名と紐づけ．なけらばField名が使われる．
	Value float32
	Time  time.Time
}
type illuminance struct {
	ID    int `gorm:"cloumn:id"` //タグつけるとcloumn名と紐づけ．なけらばField名が使われる．
	Value float32
	Time  time.Time
}
type moved struct {
	ID    int `gorm:"cloumn:id"` //タグつけるとcloumn名と紐づけ．なけらばField名が使われる．
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
	token string
	db    *gorm.DB
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

	humidityOrm := humidity{}
	humidityOrm.Value = devices[0].NewestEvents.Hu.Val
	humidityOrm.Time = devices[0].NewestEvents.Hu.CreatedAt

	illuminanceOrm := illuminance{}
	illuminanceOrm.Value = devices[0].NewestEvents.Il.Val
	illuminanceOrm.Time = devices[0].NewestEvents.Il.CreatedAt

	temperatureOrm := temperature{}
	temperatureOrm.Value = devices[0].NewestEvents.Te.Val
	temperatureOrm.Time = devices[0].NewestEvents.Te.CreatedAt

	movedOrm := moved{}
	movedOrm.Value = devices[0].NewestEvents.Mo.Val
	movedOrm.Time = devices[0].NewestEvents.Mo.CreatedAt

	// db.Table("Humidity").Create(&humidityOrm)
	r.db.Save(&humidityOrm)
	r.db.Save(&illuminanceOrm)
	r.db.Save(&temperatureOrm)
	r.db.Save(&movedOrm)

	return nil
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

	bleAdaptor := ble.NewClientAdaptor("E1:EC:E9:82:8F:60")
	err = bleAdaptor.Connect()
	if err != nil {
		log.Print("Failed to connect to bluetooth device")
		panic(err)
	}

	fmt.Println("Connected to device")
	// byteCode, readErr := bleAdaptor.ReadCharacteristic("cba20002-224d-11e6-9fb8-0002a5d5c51b")
	byteCode, readErr := bleAdaptor.ReadCharacteristic("cba20d00-224d-11e6-9fb8-0002a5d5c51b")

	if readErr != nil {
		log.Print("Failed to read characteristic")
		panic(err)
	}
	fmt.Println("Has read characteristic")
	for _, s := range byteCode {
		fmt.Printf("%02x", s)
		fmt.Println()
	}
	fmt.Println()

	// for {
	// remoOperator.retrieveSensorValue()
	// time.Sleep(5 * time.Minute)
	// }
}
