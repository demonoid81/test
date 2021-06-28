package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func main() {

	url := "http://192.168.10.244/gql"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("operations", "{\"operationName\":\"ParsePerson\",\"variables\":{\"passport\":null,\"photo\":null},\"query\":\"query ParsePerson($passport: Upload, $photo: Upload) { response: parsePerson(passport: $passport, photo: $photo) { avatar passport}}\"}")
	_ = writer.WriteField("map", "{\"0\":[\"variables.passport\"], \"1\":[\"variables.photo\"]}")
	passportFile, err := os.Open("/home/demonoid/test/photo_2021-04-23_04-17-40.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer passportFile.Close()
	passport, err := writer.CreateFormFile("0", filepath.Base("/home/demonoid/test/photo_2021-04-23_04-17-40.jpg"))
	_, err = io.Copy(passport, passportFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	photoFile, err := os.Open("/home/demonoid/test/photo_2021-04-23_04-18-04.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer photoFile.Close()
	photo, err := writer.CreateFormFile("1", filepath.Base("/home/demonoid/test/photo_2021-04-23_04-18-04.jpg"))
	_, err = io.Copy(photo, photoFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MjQ0NDEyNjQsInR5cGUiOiJTZWxmRW1wbG95ZWQiLCJ1c2VyIjoiZjM3Zjc4YmUtMTRiOS00ZmU4LWJmZDUtNmU1YTYwMjViZTYxIn0.XyYZBLaqEERBF5pgPQRyagK0imXSYZRApQTU15jtYD4")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
