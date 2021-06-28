package taxpayerStatus

import (
	"errors"
	"fmt"
	"time"

	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/authRequest"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/models/smz"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/pkg/soap"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/smz"
)

func TaxpayerStatus(app *app.App) (*smz_models.GetTaxpayerStatusResponse, error) {
	var err error
	app.Cfg.Api.FnsTempToken.Lock()
	expire := app.Cfg.Api.FnsTempToken.Expire
	token := app.Cfg.Api.FnsTempToken.Token
	app.Cfg.Api.FnsTempToken.Unlock()

	if expire.Before(time.Now()) {
		token, err = authRequest.AuthRequest(app)
		if err != nil {
			fmt.Println("Error update token")
			return nil, err
		}
	}

	params := smz_models.GetTaxpayerStatusRequest{
		Inn: "246215107274",
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
		return nil, err
	}

	MessageID := reply.MessageId

	for {
		select {
		case <-time.After(30 * time.Second):
			return nil, errors.New("Timed out expire FNS req")
		case <-time.After(5 * time.Second):
			reply, err := service.GetMessage(&smz.GetMessageRequest{
				MessageId: MessageID,
			})
			if err != nil {
				return nil, err
			}

			MessageID := reply

			if *MessageID.ProcessingStatus != "PROCESSING" {
				fmt.Println(MessageID.Message.GetTaxpayerStatusResponse)
				return &MessageID.Message.GetTaxpayerStatusResponse, nil
			}
		}
	}
}
