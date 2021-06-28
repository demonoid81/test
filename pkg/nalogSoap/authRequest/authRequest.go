package authRequest

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/models/tns"
	"github.com/sphera-erp/sphera/pkg/nalogSoap/pkg/soap"
)

type ReqMessage struct {
	AuthRequest tns.AuthRequest `xml:"urn://x-artefacts-gnivc-ru/ais3/kkt/AuthService/types/1.0 AuthRequest,omitempty"`
}

type GetMessageRequest struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/inplat/servin/OpenApiMessageConsumerService/types/1.0 GetMessageRequest"`

	Message ReqMessage `xml:"Message,omitempty" json:"Message,omitempty"`
}

type ResMessage struct {
	AuthResponse tns.AuthResponse
	// PostPlatformRegistrationResponse smz.PostPlatformRegistrationResponse
}

type GetMessageResponse struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/inplat/servin/OpenApiMessageConsumerService/types/1.0 GetMessageResponse"`

	Message ResMessage `xml:"Message,omitempty" json:"Message,omitempty"`
}

type OpenApiMessageConsumerServicePortType interface {

	// Error can be either of the following types:
	//
	//   - AuthenticationException

	GetMessage(request *GetMessageRequest) (*GetMessageResponse, error)

	GetMessageContext(ctx context.Context, request *GetMessageRequest) (*GetMessageResponse, error)
}

type openApiMessageConsumerServicePortType struct {
	client *soap.Client
}

func NewOpenApiMessageConsumerServicePortType(client *soap.Client) OpenApiMessageConsumerServicePortType {
	return &openApiMessageConsumerServicePortType{
		client: client,
	}
}

func (service *openApiMessageConsumerServicePortType) GetMessageContext(ctx context.Context, request *GetMessageRequest) (*GetMessageResponse, error) {
	response := new(GetMessageResponse)
	err := service.client.CallContext(ctx, "urn:GetMessageRequest", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (service *openApiMessageConsumerServicePortType) GetMessage(request *GetMessageRequest) (*GetMessageResponse, error) {
	return service.GetMessageContext(
		context.Background(),
		request,
	)
}

func AuthRequest(app *app.App) (string, error) {

	app.Cfg.Api.FnsTempToken.Lock()
	defer app.Cfg.Api.FnsTempToken.Unlock()
	params := tns.AuthRequest{
		AuthAppInfo: tns.AuthAppInfo{
			MasterToken: app.Cfg.Api.MasterToken,
		},
	}

	client := soap.NewClient("https://himself-ktr-api.nalog.ru:8090/open-api/AuthService?wsdl")
	service := NewOpenApiMessageConsumerServicePortType(client)
	reply, err := service.GetMessage(&GetMessageRequest{
		Message: ReqMessage{
			AuthRequest: params,
		},
	})
	if err != nil {
		return "", err
	}

	app.Cfg.Api.FnsTempToken.Token = reply.Message.AuthResponse.Result.Token

	expireTime, err := time.Parse(time.RFC3339, reply.Message.AuthResponse.Result.ExpireTime)
	if err != nil {
		fmt.Println("Error parse date")
		return "", err
	}

	fmt.Println("******************* NEW TOKEN ******************")

	app.Cfg.Api.FnsTempToken.Expire = expireTime

	return reply.Message.AuthResponse.Result.Token, nil

}
