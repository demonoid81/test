package passports

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type (
	// Option applies some option to clientOptions.
	Option func(opts *clientOptions)
)

type (
	// Client is a transport for API calls.
	Client struct {
		options clientOptions
	}

	clientOptions struct {
		httpClient  *http.Client
		credential  string
		endpointURL *url.URL
	}
)

var URL string = "https://latest.dbrain.io"

func NewClient(endpointURL *url.URL, opts ...Option) *Client {
	options := clientOptions{
		httpClient:  http.DefaultClient,
		credential:  "dbrain_api_key",
		endpointURL: endpointURL,
	}

	applyOptions(&options, opts...)

	return &Client{
		options: options,
	}
}

func applyOptions(options *clientOptions, opts ...Option) {
	for _, o := range opts {
		o(options)
	}
}

func (c *Client) RecognizeInS3(ctx context.Context, method string) {

}

func PassportRecognize(ctx context.Context, app *app.App, passport *graphql.Upload, recognizeResult *models.RecognizedFields, userUUID *uuid.UUID) (err error) {
	// if db == nil {
	// 	db, err = app.Cockroach.BeginX(ctx)
	// 	if err != nil {
	// 		errString := err.Error()
	// 		recognizeResult.Error = &errString
	// 	}
	// }
	req, err := createRecognizePostReq(passport)
	return executeRecognizePostReq(req, recognizeResult, userUUID, app, ctx)
}

func createRecognizePostReq(passport *graphql.Upload) (*http.Request, error) {
	//extraFields := map[string]string{}
	extraFields := map[string]string{
		"external_check_is_valid": "true",
	}

	URL := fmt.Sprintf("https://latest.dbrain.io/recognize")
	//b, w, err := createMultipartFormS3("image","users", "862c8003-e864-4751-9129-8b003cb45f29", extraFields, app )
	b, w, err := createMultipartFormRequest(passport, extraFields)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", URL, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", "fYwwuFtKN7Wu3YgDTjxoo4dwFZ2bvbh9aU4pDIaffj0x6f5xgNJOLOaKoysAF0Wh"))
	return req, nil
}

func executeRecognizePostReq(req *http.Request, recognizeResult *models.RecognizedFields, userUUID *uuid.UUID, app *app.App, ctx context.Context) (err error) {
	rClient := &http.Client{}
	response, err := rClient.Do(req)
	if err != nil {
		errString := err.Error()
		recognizeResult.Error = &errString
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		errString := "assport_recognition_service_error"
		recognizeResult.Error = &errString
		return
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		errString := err.Error()
		recognizeResult.Error = &errString
		return
	}
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		errString := err.Error()
		recognizeResult.Error = &errString
		return
	}

	var token *string
	if err = pglxqb.Select("notification_token").
		From("users").
		Where(pglxqb.Eq{"uuid": userUUID}).
		RunWith(app.Cockroach).QueryRow(ctx).Scan(&token); err != nil {
		app.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
	}
	err = parse(jsonMap, recognizeResult)
	fmt.Println(recognizeResult)
	if err != nil {
		errString := err.Error()
		recognizeResult.Error = &errString
		if token != nil {
			text := "Ошибка в распознавании или валидации"
			app.SendPush("192.168.10.244:9999", []string{*token}, text)
		}
		return
	}

	if token != nil {
		text := "Ваш паспорт распознан и валидирован"
		app.SendPush("192.168.10.244:9999", []string{*token}, text)
	}

	// запишем результат
	fmt.Println(recognizeResult)
	recognizeResultJson, err := json.Marshal(recognizeResult)
	_, err = pglxqb.Update("persons").
		Set("recognize_result", data).
		Set("recognized_fields", recognizeResultJson).
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		RunWith(app.Cockroach).Exec(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

func createMultipartFormRequest(passport *graphql.Upload, extraFormFields map[string]string) (b bytes.Buffer, w *multipart.Writer, err error) {
	w = multipart.NewWriter(&b)
	var fw io.Writer
	if fw, err = w.CreateFormFile("image", passport.Filename); err != nil {
		return
	}
	if _, err = io.Copy(fw, passport.File); err != nil {
		return
	}

	for k, v := range extraFormFields {
		w.WriteField(k, v)
	}

	w.Close()
	return
}

func createMultipartFormS3(FieldName, bucket, object string, extraFormFields map[string]string, app *app.App) (b bytes.Buffer, w *multipart.Writer, err error) {
	w = multipart.NewWriter(&b)

	reader, err := app.S3.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	var fw io.Writer

	if fw, err = w.CreateFormFile(FieldName, object); err != nil {
		return
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = io.CopyN(fw, reader, stat.Size); err != nil {
		return
	}

	for k, v := range extraFormFields {
		w.WriteField(k, v)
	}

	w.Close()

	return
}

func createField(fields interface{}) *models.RecognizedField {
	var confidence float64
	var text string
	for fKey, fField := range fields.(map[string]interface{}) {
		if fKey == "confidence" {
			confidence = fField.(float64)
		}
		if fKey == "text" {
			text = fField.(string)
		}
	}
	field := new(models.RecognizedField)
	field.Confidence = confidence
	field.Result = text
	return field
}

func parse(jsonMap map[string]interface{}, result *models.RecognizedFields) error {
	for _, item := range jsonMap["items"].([]interface{}) {
		if item.(map[string]interface{})["error"] != nil {
			errorString := item.(map[string]interface{})["error"].(string)
			result.Error = &errorString
			return nil
		} else {
			switch item.(map[string]interface{})["doc_type"].(string) {
			case "passport_main":
				fields := item.(map[string]interface{})["fields"].(map[string]interface{})
				for key, field := range fields {
					if key != "" {
						switch key {
						case "first_name":
							result.Name = createField(field)
						case "surname":
							result.Surname = createField(field)
						case "other_names":
							result.Patronymic = createField(field)
						case "sex":
							result.Gender = createField(field)
						case "date_of_birth":
							result.BirthDate = createField(field)
						case "series_and_number":
							serial := createField(field)
							s := strings.Split(serial.Result, " ")
							if len(s) == 2 {
								number := new(models.RecognizedField)
								number.Confidence = serial.Confidence
								number.Result = s[1]
								serial.Result = s[0]
								result.Number = number
								result.Serial = serial
							}

						case "date_of_issue":
							result.DateIssue = createField(field)
						case "subdivision_code":
							result.DepartmentCode = createField(field)
						case "issuing_authority":
							result.Department = createField(field)
						}
					}
				}
				return nil
			default:
				return nil
			}
		}
	}
	return nil
}
