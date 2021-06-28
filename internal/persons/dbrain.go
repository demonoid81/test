package persons

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/objectStorage"
	"github.com/sphera-erp/sphera/internal/passports"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/bindPartnerStatus"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/bindPartnerWithInn"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"moul.io/http2curl"
)

func ParsePerson(ctx context.Context, app *app.App, photo *graphql.Upload, passport *graphql.Upload) (*models.PersonValidateStatus, error) {
	results := models.PersonValidateStatus{
		Passport: false,
		Avatar:   false,
	}
	userUUID, err := middleware.ExtractUserInTokenMetadata(ctx, app)
	if err != nil {
		return nil, gqlerror.Errorf("Error get user uuid from context")
	}
	tx, err := app.Cockroach.BeginX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	if photo == nil {
		bucket, uuidObject, err := GetPersonaPhotoByUserUUID(ctx, userUUID, tx, app.Logger)
		if err != nil {
			return nil, gqlerror.Errorf("Error run transaction")
		}
		fmt.Println(bucket, uuidObject)
		if uuidObject != nil {
			reader, err := app.S3.GetObject(context.Background(), *bucket, uuidObject.String(), minio.GetObjectOptions{})
			if err != nil {
				return nil, gqlerror.Errorf("No get object")
			}
			stat, err := reader.Stat()
			if err != nil {
				return nil, gqlerror.Errorf("No get stat of object")
			}
			photo = new(graphql.Upload)
			photo.File = reader
			photo.Size = stat.Size
			photo.Filename = uuidObject.String()
			results.Avatar = true
		}
	} else {
		uuidPhoto, err := upload(ctx, *photo, "users", app, tx)
		if err != nil {
			return nil, gqlerror.Errorf("Error run transaction")
		}
		_, err = pglxqb.Update("persons").
			Set("uuid_photo", uuidPhoto).
			Where(pglxqb.Eq{"uuid_user": userUUID}).
			RunWith(tx).Exec(ctx)
		if err != nil {
			return nil, gqlerror.Errorf("Error run transaction")
		}
		results.Avatar = true
	}
	if passport == nil {
		bucket, uuidObject, err := GetPassportPhotoByUserUUID(ctx, userUUID, tx, app.Logger)
		if err != nil {
			return nil, gqlerror.Errorf("Error run transaction")
		}
		fmt.Println(bucket, uuidObject)
		if uuidObject != nil {
			reader, err := app.S3.GetObject(context.Background(), *bucket, uuidObject.String(), minio.GetObjectOptions{})
			if err != nil {
				return nil, gqlerror.Errorf("No get object")
			}
			stat, err := reader.Stat()
			if err != nil {
				return nil, gqlerror.Errorf("No get stat of object")
			}
			passport = new(graphql.Upload)
			passport.File = reader
			passport.Size = stat.Size
			passport.Filename = uuidObject.String()
			results.Passport = true
		}
	} else {
		uuidPassportScan, err := upload(ctx, *passport, "users", app, tx)
		if err != nil {
			return nil, gqlerror.Errorf("Error run transaction")
		}
		var uuidPassport *uuid.UUID
		err = pglxqb.
			Select("p.uuid").
			From("persons").
			LeftJoin("passports p on persons.uuid_passport = p.uuid").
			Where(pglxqb.Eq{"uuid_user": userUUID}).
			RunWith(tx).
			QueryRow(ctx).Scan(&uuidPassport)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		if uuidPassport != nil {
			_, err = pglxqb.Update("passports").
				Set("uuid_scan", uuidPassportScan).
				Where(pglxqb.Eq{"uuid": uuidPassport}).
				RunWith(tx).Exec(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error update passport scan")
				return nil, gqlerror.Errorf("Error update passport scan")
			}
		} else {
			uuidPassport := uuid.New()
			if _, err = pglxqb.Insert("passports").
				Columns("uuid", "uuid_scan").
				Values(uuidPassport, uuidPassportScan).
				RunWith(tx).Exec(ctx); err != nil {
				app.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error update passport scan")
				return nil, gqlerror.Errorf("Error update passport scan")
			}
			if _, err = pglxqb.Update("persons").
				Set("uuid_passport", uuidPassport).
				Where(pglxqb.Eq{"uuid_user": userUUID}).
				RunWith(tx).Exec(ctx); err != nil {
				app.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error update passport scan")
				return nil, gqlerror.Errorf("Error update passport scan")
			}
		}

		results.Passport = true
	}
	err = tx.Commit(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "PersonaMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	fmt.Println("------------------", results.Passport, "--------", results.Avatar, "------------------")
	if results.Passport && results.Avatar {
		fmt.Println("go to dbrain")
		go Exec(context.Background(), app, photo, passport, &userUUID)
	}
	return &results, nil
}

func Exec(ctx context.Context, app *app.App, photo *graphql.Upload, passport *graphql.Upload, userUUID *uuid.UUID) {
	recognizeResult := new(models.RecognizedFields)
	req, _ := createDistancePostReq(photo, passport)

	var token *string
	err := pglxqb.Select("notification_token").
		From("users").
		Where(pglxqb.Eq{"uuid": userUUID}).RunWith(app.Cockroach).QueryRow(ctx).Scan(&token)
	if err != nil {
		app.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
		return
	}

	result, err := executeDistancePostReq(req, userUUID, app, ctx)
	if err != nil {
		if err.Error() != "notEq" {
			rollback(ctx, app, userUUID, token)
			fmt.Println("executeDistancePostReq err")
			sendErrPush(token, app)
			return
		}
	}
	// выполним сохранение объектов
	if result {
		if err = passports.PassportRecognize(ctx, app, passport, recognizeResult, userUUID); err != nil {
			rollback(ctx, app, userUUID, token)
			fmt.Println("PassportRecognize err")
			if token != nil {
				text := "Произошла ошибка в проверке"
				app.SendPush("192.168.10.244:9999", []string{*token}, text)
			}
			return
		}
	} else {
		fmt.Println("not eq")
		rollback(ctx, app, userUUID, token)
		if token != nil {
			text := "Ваше селфи и фотография в паспорте не идентичны, попробуйте заново"
			app.SendPush("192.168.10.244:9999", []string{*token}, text)
		}
		return
	}

	// проверим что у нас нет такой же записи

	var count int
	err = pglxqb.Select("COUNT(*)").
		From("passports").
		Where("serial = ?", recognizeResult.Serial.Result).
		Where("number = ?", recognizeResult.Number.Result).
		RunWith(app.Cockroach).
		Scan(ctx, &count)
	if err != nil {
		sendErrPush(token, app)
	}

	if count > 0 {
		rollback(ctx, app, userUUID, token)
		fmt.Println("Пользователь с таким паспортом уже зарегистрирован в системе")
		if token != nil {
			text := "Пользователь с таким паспортом уже зарегистрирован в системе"
			app.SendPush("192.168.10.244:9999", []string{*token}, text)
		}
		return
	}

	fmt.Println(recognizeResult.BirthDate.Result)
	s := strings.Split(recognizeResult.BirthDate.Result, ".")

	birthDate, err := time.Parse("2006-01-02", s[2]+"-"+s[1]+"-"+s[0])
	if err != nil {
		fmt.Println("err parse birthDate")
		sendErrPush(token, app)
		return
	}

	var uuidPassport uuid.UUID
	if err = pglxqb.Update("persons").
		Set("surname", recognizeResult.Surname.Result).
		Set("name", recognizeResult.Name.Result).
		Set("patronymic", recognizeResult.Patronymic.Result).
		Set("birth_date", birthDate).
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		Suffix("RETURNING uuid_passport").
		RunWith(app.Cockroach).QueryRow(ctx).Scan(&uuidPassport); err != nil {
		errString := "Error update person info"
		fmt.Println("Error update person info")
		sendErrPush(token, app)
		recognizeResult.Error = &errString
		return
	}

	s = strings.Split(recognizeResult.DateIssue.Result, ".")
	fmt.Println(s[2] + "-" + s[1] + "-" + s[0])
	dateIssue, err := time.Parse("2006-01-02", s[2]+"-"+s[1]+"-"+s[0])
	if err != nil {
		errString := "Error parse date"
		fmt.Println("Error parse dateIssue")
		recognizeResult.Error = &errString
		sendErrPush(token, app)
		return
	}

	if _, err = pglxqb.Update("passports").
		Set("serial", recognizeResult.Serial.Result).
		Set("number", recognizeResult.Number.Result).
		Set("department", recognizeResult.Department.Result).
		Set("date_issue", dateIssue).
		Set("department_code", recognizeResult.DepartmentCode.Result).
		Where(pglxqb.Eq{"uuid": uuidPassport}).
		RunWith(app.Cockroach).Exec(ctx); err != nil {
		errString := "Error update user passport"
		fmt.Println("Error update user passport")
		sendErrPush(token, app)
		recognizeResult.Error = &errString
		return
	}

	req, _ = createGetINNPostReq(recognizeResult)

	rClient := &http.Client{}

	response, err := rClient.Do(req)
	if err != nil {
		fmt.Println("Error send get INN")
		fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error send get INN stat")
		fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
		return
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error send get INN get body")
		fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
		return
	}
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		fmt.Println("Error send get INN marshal")
		fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
		return
	}

	inn := ""
	for _, item := range jsonMap["items"].([]interface{}) {
		if item.(map[string]interface{})["ИНН"] != nil {
			inn = item.(map[string]interface{})["ИНН"].(string)
			if token != nil {
				fmt.Println("По распознанным  данным определен ИНН " + inn)
				text := "По распознанным  данным определен ИНН " + inn
				app.SendPush("192.168.10.244:9999", []string{*token}, text)
			}
		} else {
			fmt.Println("По распознанным данным не удалось определить ИНН")
			if token != nil {
				text := "По распознанным данным не удалось определить ИНН"
				app.SendPush("192.168.10.244:9999", []string{*token}, text)
			}
			return
		}
	}

	if inn != "" {

		var count int
		err = pglxqb.Select("COUNT(*)").
			From("persons").
			Where("inn = ?", inn).
			RunWith(app.Cockroach).
			Scan(ctx, &count)
		if err != nil {
			sendErrPush(token, app)
		}

		if count > 0 {
			rollback(ctx, app, userUUID, token)
			fmt.Println("Пользователь с таким ИНН уже зарегистрирован в системе")
			if token != nil {
				text := "Пользователь с таким ИНН уже зарегистрирован в системе"
				app.SendPush("192.168.10.244:9999", []string{*token}, text)
			}
			return
		}

		if _, err = pglxqb.Update("persons").
			Set("inn", inn).
			Where(pglxqb.Eq{"uuid_user": userUUID}).
			RunWith(app.Cockroach).Exec(ctx); err != nil {
			fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
			errString := err.Error()
			fmt.Println("err save inn")
			recognizeResult.Error = &errString
		}

		// req, _ = validateINNPostReq(inn)

		// response, err = rClient.Do(req)

		// if response.StatusCode != http.StatusOK {
		// 	fmt.Println("err validate inn status ok")
		// 	fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
		// 	return
		// }
		// defer response.Body.Close()

		// data, err = ioutil.ReadAll(response.Body)
		// if err != nil {
		// 	fmt.Println("err validate inn read body")
		// 	fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
		// 	return
		// }

		// jsonMap = make(map[string]interface{})
		// err = json.Unmarshal(data, &jsonMap)
		// if err != nil {
		// 	fmt.Println("err validate inn unmarshall")
		// 	fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
		// 	return
		// }

		// res := jsonMap["status"].(bool)

		// if res == true {
		// 	fmt.Println("Проверка ИНН пройдена, Вы являетесь самозанятым")
		// 	if token != nil {
		// 		text := "Проверка ИНН пройдена, Вы являетесь самозанятым"
		// 		app.SendPush("192.168.10.244:9999", []string{*token}, text)
		// 	}
		// if _, err = pglxqb.Update("persons").
		// 	Set("validated", true).
		// 	Where(pglxqb.Eq{"uuid_user": userUUID}).
		// 	RunWith(app.Cockroach).Exec(ctx); err != nil {
		// 	errString := err.Error()
		// 	fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
		// 	fmt.Println("err save")
		// 	recognizeResult.Error = &errString
		// }
		// } else {
		// 	fmt.Println("Проверка ИНН не пройдена, Вы не являетесь самозанятым")
		// 	if token != nil {
		// 		text := "Проверка ИНН не пройдена, Вы не являетесь самозанятым"
		// 		app.SendPush("192.168.10.244:9999", []string{*token}, text)
		// 	}
		// }
	}

	// Отправимзаявку на работу с налогами

	reqId, err := bindPartnerWithInn.BindPartnerWithInn(app, inn)

	fmt.Println(reqId)

	for {
		select {
		case <-time.After(10 * time.Minute):
			fmt.Println("Time out FNS Request")
			if token != nil {
				text := "в течении 30 мин вы не подтвердили партнерство с платформой"
				app.SendPush("192.168.10.244:9999", []string{*token}, text)
			}
			return
		case <-time.After(1 * time.Minute):
			fmt.Println("*************************************************** Validate **************************************************************")
			result, err := bindPartnerStatus.BindPartnerStatus(app, reqId)
			if err != nil {
				if err.Error() != "Timeout" {
					fmt.Println("Error FNS Request")
					if token != nil {
						text := "Произошла ошибка при проверке партнерства, попробуйте поптыку позже"
						app.SendPush("192.168.10.244:9999", []string{*token}, text)
					}
					return
				}
			}
			if result {
				if _, err = pglxqb.Update("persons").
					Set("validated", true).
					Where(pglxqb.Eq{"uuid_user": userUUID}).
					RunWith(app.Cockroach).Exec(ctx); err != nil {
					errString := err.Error()
					fixCommitAndSendErrPush(ctx, token, app, recognizeResult)
					fmt.Println("err save")
					recognizeResult.Error = &errString
					return
				}
				return
			}
		}
	}

	// for _, c := range subscriptionsParsePerson.parsePersonResult[*userUUID] {
	// 	fmt.Println("send response")
	// 	c <- recognizeResult
	// }
}

func executeDistancePostReq(req *http.Request, userUUID *uuid.UUID, app *app.App, ctx context.Context) (bool, error) {
	var err error
	rClient := &http.Client{}

	response, err := rClient.Do(req)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return false, errors.New("service dbrain error")
	}
	fmt.Println(response.Body)
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		return false, err
	}
	// паспорт и лицо совпали
	distance, err := parseDistanceResponse(jsonMap)
	if err != nil {
		return false, err
	}
	if distance {
		// запишем результат

		_, err := pglxqb.Update("persons").Set("distance_result", data).
			Where(pglxqb.Eq{"uuid_user": userUUID}).
			RunWith(app.Cockroach).Exec(ctx)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, errors.New("notEq")
}

func rollback(ctx context.Context, app *app.App, userUUID *uuid.UUID, token *string) {
	db, err := app.Cockroach.BeginX(ctx)
	if err != nil {
		sendErrPush(token, app)
		return
	}
	defer db.Rollback(ctx)
	var uuidPassport *uuid.UUID
	if err = pglxqb.Update("persons").
		Set("uuid_photo", nil).
		Where(pglxqb.Eq{"uuid_user": userUUID}).
		Suffix("RETURNING uuid_passport").
		RunWith(db).QueryRow(ctx).Scan(&uuidPassport); err != nil {
		fmt.Println("Error delete user photo")
		sendErrPush(token, app)
		return
	}
	if _, err = pglxqb.Update("passports").
		Set("uuid_scan", nil).
		Where(pglxqb.Eq{"uuid": uuidPassport}).
		RunWith(db).Exec(ctx); err != nil {
		fmt.Println("Error delete user passports")
		sendErrPush(token, app)
		return
	}

	err = db.Commit(ctx)
	if err != nil {
		fmt.Println("err commit")
		sendErrPush(token, app)
		return
	}
}

func createDistancePostReq(photo *graphql.Upload, passport *graphql.Upload) (*http.Request, error) {

	URL := fmt.Sprintf("https://latest.dbrain.io/face/distance")
	b, w, err := createMultipartFormRequest(photo, passport)
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

func createGetINNPostReq(recognizeResult *models.RecognizedFields) (*http.Request, error) {

	URL := fmt.Sprintf("https://api-fns.ru/api/innfl?fam=%s&nam=%s&otch=%s&bdate=%s&doctype=21&docno=%s&key=4e6ceeac34afd510dbd8c47e0d1db6a4374ccd47",
		recognizeResult.Surname.Result,
		recognizeResult.Name.Result,
		recognizeResult.Patronymic.Result,
		recognizeResult.BirthDate.Result,
		recognizeResult.Serial.Result+recognizeResult.Number.Result,
	)

	req, err := http.NewRequest("GET", URL, nil)

	command, _ := http2curl.GetCurlCommand(req)
	fmt.Println(command)

	if err != nil {
		return nil, err
	}

	return req, nil
}

func validateINNPostReq(inn string) (*http.Request, error) {

	URL := fmt.Sprintf("https://statusnpd.nalog.ru/api/v1/tracker/taxpayer_status")

	data := map[string]interface{}{
		"inn":         inn,
		"requestDate": time.Now().Local().Format("2006-01-02"),
	}

	b, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func createMultipartFormRequest(photo *graphql.Upload, passport *graphql.Upload) (b bytes.Buffer, w *multipart.Writer, err error) {
	w = multipart.NewWriter(&b)
	files := []*graphql.Upload{photo, passport}
	for i, file := range files {
		var fw io.Writer
		if fw, err = w.CreateFormFile(fmt.Sprintf("image%d", i+1), file.Filename); err != nil {
			return
		}
		var f bytes.Buffer
		if _, err = io.Copy(fw, io.TeeReader(file.File, &f)); err != nil {
			return
		}
		file.File = &f
	}
	w.Close()
	return
}

func parseDistanceResponse(data map[string]interface{}) (bool, error) {
	var distance float64
	for _, item := range data["items"].([]interface{}) {
		if item.(map[string]interface{})["distance"] != nil {
			distance = item.(map[string]interface{})["distance"].(float64)
			return distance < 0.50, nil
		}
		return false, errors.New(item.(map[string]interface{})["reason"].(string))
	}
	return false, nil
}

func upload(ctx context.Context, file graphql.Upload, bucket string, app *app.App, db pglxqb.BaseRunner) (uuid.UUID, error) {
	uuidObject := uuid.New()
	//encryption := encrypt.DefaultPBKDF([]byte(r.env.Cfg.Api.S3Key), []byte(bucket + uuidObject.String()))
	//minio.PutObjectOptions{ServerSideEncryption: encryption}
	n, err := app.S3.PutObject(ctx, bucket, uuidObject.String(), file.File, file.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		app.Logger.Error().Str("module", "objectStorage").Str("function", "SingleUpload").Err(err).Msg("Error uploading file")
		return uuid.Nil, gqlerror.Errorf("Error uploading file: %s", err)
	}
	app.Logger.Debug().Str("module", "objectStorage").Str("function", "SingleUpload").Msgf("Uploaded %s of size: %d Successfully.", file.Filename, n.Size)
	err = objectStorage.PutObjectInDB(ctx, app, db, n, uuidObject)
	if err != nil {
		// удалим мертвый объект так как нанего нет ссылки
		opts := minio.RemoveObjectOptions{
			GovernanceBypass: true,
		}

		err = app.S3.RemoveObject(ctx, bucket, file.Filename, opts)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return uuidObject, nil
}

func fixCommitAndSendErrPush(ctx context.Context, token *string, app *app.App, recognizeResult *models.RecognizedFields) {
	sendErrPush(token, app)
}

func sendErrPush(token *string, app *app.App) {
	if token != nil {
		text := "Произошла ошибка в проверке"
		app.SendPush("192.168.10.244:9999", []string{*token}, text)
	}
}
