package innByPersonalInfo

import (
	"fmt"
	"time"

	"github.com/sphera-erp/sphera/pkg/nalogSoap/models/smz"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/pkg/soap"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/smz"
)

func InnByPersonalInfo(masterToken, token string) error {

	params := smz_models.GetInnByPersonalInfoRequest{
		FirstName:      "Михаил",
		SecondName:     "Новиков",
		Patronymic:     "Владимирович",
		Birthday:       "1984-04-25",
		PassportSeries: "0405",
		PassportNumber: "095677",
	}

	headers := map[string]string{
		"FNS-OpenApi-Token":     token,
		"FNS-OpenApi-UserToken": masterToken,
	}

	client := soap.NewClient("https://himself-ktr-api.nalog.ru:8090/ais3/smz/SmzIntegrationService?wsdl", soap.WithHTTPHeaders(headers))
	service := smz.NewOpenApiAsyncSMZ(client)
	reply, err := service.SendMessage(&smz.SendMessageRequest{
		Message: smz.SendMessage{
			Message: params,
		},
	})
	if err != nil {
		return err
	}

	MessageID := reply.MessageId

	for {
		select {
		case <-time.After(30 * time.Second):
			return nil
		case <-time.After(5 * time.Second):
			fmt.Println("GetMessage")
			reply, err := service.GetMessage(&smz.GetMessageRequest{
				MessageId: MessageID,
			})
			if err != nil {
				return err
			}

			MessageID := reply

			fmt.Println(*MessageID.ProcessingStatus)

			if *MessageID.ProcessingStatus != "PROCESSING" {
				fmt.Println(MessageID.Message.GetInnByPersonalInfoResponse.Inn)
				return nil
			}
		}
	}
}
