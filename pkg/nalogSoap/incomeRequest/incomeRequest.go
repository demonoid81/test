package incomeRequest

import (
	"fmt"
	"time"

	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/authRequest"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/models/smz"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/pkg/soap"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/smz"
)

func IncomeRequest(app *app.App, jobCost float64, jobName, personInn, orgInn string) (string, string, error) {

	var err error
	app.Cfg.Api.FnsTempToken.Lock()
	expire := app.Cfg.Api.FnsTempToken.Expire
	token := app.Cfg.Api.FnsTempToken.Token
	app.Cfg.Api.FnsTempToken.Unlock()
	fmt.Println("************************************************************:    ", expire.Before(time.Now()))
	if expire.Before(time.Now()) {
		token, err = authRequest.AuthRequest(app)
		if err != nil {
			fmt.Println("Error update token")
			return "", "", err
		}
	}

	loc, _ := time.LoadLocation("Europe/Moscow")

	params := smz_models.PostIncomeRequest{
		Inn:           personInn,
		RequestTime:   time.Now().In(loc).Format(time.RFC3339),
		OperationTime: time.Now().In(loc).Format(time.RFC3339),
		IncomeType:    "FROM_LEGAL_ENTITY",
		CustomerInn:   orgInn,
		Services: smz_models.IncomeService{
			Amount:   jobCost,
			Name:     jobName,
			Quantity: 1,
		},
		TotalAmount: jobCost,
	}

	headers := map[string]string{
		"FNS-OpenApi-Token":     token,
		"FNS-OpenApi-UserToken": app.Cfg.Api.MasterToken,
	}

	client := soap.NewClient("https://himself-ktr-api.nalog.ru:8090/ais3/smz/SmzIntegrationService?wsdl", soap.WithHTTPHeaders(headers))
	service := smz.NewOpenApiAsyncSMZ(client)
	reply, err := service.SendMessage(&smz.SendMessageRequest{
		Message: smz.SendMessage{
			Message: params,
		},
	})
	if err != nil {
		return "", "", err
	}

	MessageID := reply.MessageId

	for {
		select {
		case <-time.After(30 * time.Second):
			return "", "", err
		case <-time.After(5 * time.Second):
			fmt.Println("GetMessage")
			reply, err := service.GetMessage(&smz.GetMessageRequest{
				MessageId: MessageID,
			})
			if err != nil {
				return "", "", err
			}

			MessageID := reply

			fmt.Println(*MessageID.ProcessingStatus)

			if *MessageID.ProcessingStatus != "PROCESSING" {
				fmt.Println(MessageID.Message.PostIncomeResponse.Link)
				return MessageID.Message.PostIncomeResponse.Link, MessageID.Message.PostIncomeResponse.ReceiptId, nil
			}
		}
	}
}
