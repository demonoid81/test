package smsCommunicator

import (
	"fmt"
	"github.com/tiaguinho/gosoap"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"moul.io/http2curl"
)

var wsdlUrl = "https://www.mcommunicator.ru/m2m/m2m_api.asmx"

func client(token string) *gosoap.Client {
	httpClient := &http.Client{
		Timeout: 1500 * time.Millisecond,
	}
	soap, err := gosoap.SoapClient(wsdlUrl, httpClient)
	if err != nil {
		log.Fatalf("SoapClient error: %s", err)
	}
	soap.HeaderParams = gosoap.HeaderParams{
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
	return soap
}

func preparePhone(phone string) string {
	var re = regexp.MustCompile(`\D`)
	phone = re.ReplaceAllString(phone, "")

	if len(phone) == 10 {
		return fmt.Sprintf("7%s", phone)
	}
	if phone[:1] == "8" {
		return fmt.Sprintf("7%s", phone[1:])
	}
	return phone
}

//5aa37174-bb4e-481d-b02f-12b682db6da7

func SendMessage(phone, message string) error {
	apiUrl := "https://a2p-sms-https.beeline.ru"
	resource := "/proto/http"

	data := url.Values{}
	data.Add("user", "1680291")
	data.Add("pass", "16802911")
	data.Add("action", "post_sms")
	data.Add("message", message)
	data.Add("target", phone)
	data.Add("sender",  "jamrabota")
	data.Add("period", "600")

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	r, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	command, _ := http2curl.GetCurlCommand(r)
	fmt.Println(command)
	fmt.Println(r)
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return nil
}
