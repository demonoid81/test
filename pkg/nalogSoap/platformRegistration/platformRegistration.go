package platformRegistration

import (
	"fmt"
	"time"

	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/authRequest"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/models/smz"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/pkg/soap"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/smz"
)

const img = "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"

func PlatformRegistration(app *app.App) error {

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
			return err
		}
	}

	params := smz_models.PostPlatformRegistrationRequest{
		PartnerName:        "ВФМ технолоджи",
		PartnerType:        "PARTNER",
		PartnerDescription: "Приложение, которое меняет представление о прозрачной подработке для самозанятых рядом с домом. Выбирайте смены и зарабатывайте. Просто, легально и безопасно.",
		PartnerConnectable: "true",
		TransitionLink:     "https://wfmt.ru",
		PartnersText:       "Платформа подработок для самозанятых",
		PartnerImage:       img,
		Inn:                "9710087299",
		Phone:              "8 (964) 482-51-56",
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
		return err
	}

	MessageID := reply.MessageId

	for {
		select {
		case <-time.After(30 * time.Second):
			return nil
		case <-time.After(5 * time.Second):
			reply, err := service.GetMessage(&smz.GetMessageRequest{
				MessageId: MessageID,
			})
			if err != nil {
				return err
			}
			MessageID := reply

			fmt.Println(MessageID.Message.PostPlatformRegistrationResponse.PartnerID, MessageID.Message.PostPlatformRegistrationResponse.RegistrationDate)
			return nil
		}
	}
}
