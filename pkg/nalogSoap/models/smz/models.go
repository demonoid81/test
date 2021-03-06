// Code generated by https://github.com/gocomply/xsd2go; DO NOT EDIT.
// Models for urn://x-artefacts-gnivc-ru/ais3/SMZ/SmzPartnersIntegrationService/types/1.0
package smz_models

import (
	"encoding/xml"
)

// Element
type SmzPlatformError struct {
	XMLName xml.Name `xml:"SmzPlatformError"`

	Code string `xml:"Code"`

	Message string `xml:"Message"`

	Args []SmzPlatformErrorArgs `xml:"Args"`
}

// Element
type GetTaxpayerRestrictionsRequest struct {
	XMLName xml.Name `xml:"GetTaxpayerRestrictionsRequest"`

	Inn string `xml:",any"`
}

// Element
type GetTaxpayerRestrictionsResponse struct {
	XMLName xml.Name `xml:"GetTaxpayerRestrictionsResponse"`

	RequestResult string `xml:"RequestResult"`

	RejectionCode string `xml:"RejectionCode"`
}

// Element
type GetTaxpayerStatusRequest struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/ais3/SMZ/SmzPartnersIntegrationService/types/1.0 GetTaxpayerStatusRequest"`

	Inn string `xml:",any"`
}

// Element
type GetTaxpayerStatusResponse struct {
	XMLName xml.Name `xml:"GetTaxpayerStatusResponse"`

	FirstName string `xml:"FirstName"`

	SecondName string `xml:"SecondName"`

	Patronymic string `xml:"Patronymic"`

	RegistrationTime string `xml:"RegistrationTime"`

	UnregistrationTime string `xml:"UnregistrationTime"`

	UnregistrationReason string `xml:"UnregistrationReason"`

	Activities []string `xml:"Activities"`

	Region string `xml:"Region"`

	Phone string `xml:"Phone"`

	Email string `xml:"Email"`

	AccountNumber string `xml:"AccountNumber"`

	UpdateTime string `xml:"UpdateTime"`

	RegistrationCertificateNumber string `xml:"RegistrationCertificateNumber"`
}

// Element
type PostRegistrationRequest struct {
	XMLName xml.Name `xml:"PostRegistrationRequest"`

	Inn string `xml:"Inn"`

	FirstName string `xml:"FirstName"`

	SecondName string `xml:"SecondName"`

	Patronymic string `xml:"Patronymic"`

	Birthday string `xml:"Birthday"`

	PassportSeries string `xml:"PassportSeries"`

	PassportNumber string `xml:"PassportNumber"`

	Activities []string `xml:"Activities"`

	Phone string `xml:"Phone"`

	Email string `xml:"Email"`

	BankcardNumber string `xml:"BankcardNumber"`

	BankcardAccountNumber string `xml:"BankcardAccountNumber"`

	RequestTime string `xml:"RequestTime"`

	Oktmo string `xml:"Oktmo"`
}

// Element
type PostRegistrationResponse struct {
	XMLName xml.Name `xml:"PostRegistrationResponse"`

	Id string `xml:",any"`
}

// Element
type GetRegistrationStatusRequest struct {
	XMLName xml.Name `xml:"GetRegistrationStatusRequest"`

	Id string `xml:",any"`
}

// Element
type GetRegistrationStatusResponse struct {
	XMLName xml.Name `xml:"GetRegistrationStatusResponse"`

	RequestResult string `xml:"RequestResult"`

	RejectionReason string `xml:"RejectionReason"`

	RegistrationTime string `xml:"RegistrationTime"`

	LastRegistrationTime string `xml:"LastRegistrationTime"`

	UpdateTime string `xml:"UpdateTime"`

	UnregistrationTime string `xml:"UnregistrationTime"`

	BindRequestId string `xml:"BindRequestId"`

	RegistrationCertificateNumber string `xml:"RegistrationCertificateNumber"`

	Inn string `xml:"Inn"`
}

// Element
type PostUnregistrationRequest struct {
	XMLName xml.Name `xml:"PostUnregistrationRequest"`

	Inn string `xml:"Inn"`

	Code string `xml:"Code"`
}

// Element
type PostUnregistrationResponse struct {
	XMLName xml.Name `xml:"PostUnregistrationResponse"`

	Id string `xml:",any"`
}

// Element
type PostUnregistrationRequestV2 struct {
	XMLName xml.Name `xml:"PostUnregistrationRequestV2"`

	Inn string `xml:"Inn"`

	ReasonCode string `xml:"ReasonCode"`
}

// Element
type PostUnregistrationResponseV2 struct {
	XMLName xml.Name `xml:"PostUnregistrationResponseV2"`

	Id string `xml:",any"`
}

// Element
type GetUnregistrationStatusRequest struct {
	XMLName xml.Name `xml:"GetUnregistrationStatusRequest"`

	Id string `xml:",any"`
}

// Element
type GetUnregistrationStatusResponse struct {
	XMLName xml.Name `xml:"GetUnregistrationStatusResponse"`

	RequestResult string `xml:"RequestResult"`

	RejectionReason string `xml:"RejectionReason"`

	UnregistrationTime string `xml:"UnregistrationTime"`
}

// Element
type PutTaxpayerDataRequest struct {
	XMLName xml.Name `xml:"PutTaxpayerDataRequest"`

	Inn string `xml:"Inn"`

	Phone string `xml:"Phone"`

	Email string `xml:"Email"`

	Activities []string `xml:"Activities"`

	Region string `xml:"Region"`
}

// Element
type PutTaxpayerDataResponse struct {
	XMLName xml.Name `xml:"PutTaxpayerDataResponse"`

	UpdateTime string `xml:",any"`
}

// Element
type GetTaxpayerAccountStatusRequest struct {
	XMLName xml.Name `xml:"GetTaxpayerAccountStatusRequest"`

	Inn string `xml:",any"`
}

// Element
type GetTaxpayerAccountStatusResponse struct {
	XMLName xml.Name `xml:"GetTaxpayerAccountStatusResponse"`

	BonusAmount float64 `xml:"BonusAmount"`

	UnpaidAmount float64 `xml:"UnpaidAmount"`

	DebtAmount float64 `xml:"DebtAmount"`
}

// Element
type PostBindPartnerWithInnRequest struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/ais3/SMZ/SmzPartnersIntegrationService/types/1.0 PostBindPartnerWithInnRequest"`

	Inn string `xml:"Inn"`

	Permissions []string `xml:"Permissions"`
}

// Element
type PostBindPartnerWithInnResponse struct {
	XMLName xml.Name `xml:"PostBindPartnerWithInnResponse"`

	Id string `xml:",any"`
}

// Element
type PostBindPartnerWithPhoneRequest struct {
	XMLName xml.Name `xml:"PostBindPartnerWithPhoneRequest"`

	Phone string `xml:"Phone"`

	Permissions []string `xml:"Permissions"`
}

// Element
type PostBindPartnerWithPhoneResponse struct {
	XMLName xml.Name `xml:"PostBindPartnerWithPhoneResponse"`

	Id string `xml:",any"`
}

// Element
type GetBindPartnerStatusRequest struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/ais3/SMZ/SmzPartnersIntegrationService/types/1.0 GetBindPartnerStatusRequest"`

	Id string `xml:",any"`
}

// Element
type GetBindPartnerStatusResponse struct {
	XMLName xml.Name `xml:"GetBindPartnerStatusResponse"`

	Result string `xml:"Result"`

	Inn string `xml:"Inn"`

	Permissions []string `xml:"Permissions"`

	ProcessingTime string `xml:"ProcessingTime"`
}

// Element
type PostGrantedPermissionsRequest struct {
	XMLName xml.Name `xml:"PostGrantedPermissionsRequest"`

	Inn string `xml:"Inn"`

	Permissions []string `xml:"Permissions"`
}

// Element
type PostGrantedPermissionsResponse struct {
	XMLName xml.Name `xml:"PostGrantedPermissionsResponse"`

	Id string `xml:",any"`
}

// Element
type PostUnbindPartnerRequest struct {
	XMLName xml.Name `xml:"PostUnbindPartnerRequest"`

	Inn string `xml:",any"`
}

// Element
type PostUnbindPartnerResponse struct {
	XMLName xml.Name `xml:"PostUnbindPartnerResponse"`

	UnregistrationTime string `xml:",any"`
}

// Element
type GetGrantedPermissionsRequest struct {
	XMLName xml.Name `xml:"GetGrantedPermissionsRequest"`

	Inn string `xml:",any"`
}

// Element
type GetGrantedPermissionsResponse struct {
	XMLName xml.Name `xml:"GetGrantedPermissionsResponse"`

	GrantedPermissionsList []string `xml:",any"`
}

// Element
type PostIncomeRequest struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/ais3/SMZ/SmzPartnersIntegrationService/types/1.0 PostIncomeRequest"`

	Inn string `xml:"Inn"`

	ReceiptId string `xml:"ReceiptId"`

	RequestTime string `xml:"RequestTime"`

	OperationTime string `xml:"OperationTime"`

	IncomeType string `xml:"IncomeType"`

	CustomerInn string `xml:"CustomerInn"`

	CustomerOrganization string `xml:"CustomerOrganization"`

	Services IncomeService `xml:"Services"`

	TotalAmount float64 `xml:"TotalAmount"`

	IncomeHashCode string `xml:"IncomeHashCode"`

	Link string `xml:"Link"`

	GeoInfo *GeoInfo `xml:"GeoInfo"`

	OperationUniqueId string `xml:"OperationUniqueId"`
}

// Element
type PostIncomeResponse struct {
	XMLName xml.Name `xml:"PostIncomeResponse"`

	ReceiptId string `xml:"ReceiptId"`

	Link string `xml:"Link"`
}

// Element
type PostIncomeRequestV2 struct {
	XMLName xml.Name `xml:"PostIncomeRequestV2"`

	Inn string `xml:"Inn"`

	ReceiptId string `xml:"ReceiptId"`

	RequestTime string `xml:"RequestTime"`

	OperationTime string `xml:"OperationTime"`

	IncomeType string `xml:"IncomeType"`

	CustomerInn string `xml:"CustomerInn"`

	CustomerOrganization string `xml:"CustomerOrganization"`

	SupplierInn string `xml:"SupplierInn"`

	Services []IncomeService `xml:"Services"`

	TotalAmount float64 `xml:"TotalAmount"`

	IncomeHashCode string `xml:"IncomeHashCode"`

	Link string `xml:"Link"`

	GeoInfo *GeoInfo `xml:"GeoInfo"`

	OperationUniqueId string `xml:"OperationUniqueId"`
}

// Element
type PostIncomeResponseV2 struct {
	XMLName xml.Name `xml:"PostIncomeResponseV2"`

	ReceiptId string `xml:"ReceiptId"`

	Link string `xml:"Link"`
}

// Element
type PostIncomeFromIndividualRequest struct {
	XMLName xml.Name `xml:"PostIncomeFromIndividualRequest"`

	Inn string `xml:"Inn"`

	ReceiptId string `xml:"ReceiptId"`

	RequestTime string `xml:"RequestTime"`

	OperationTime string `xml:"OperationTime"`

	SupplierInn string `xml:"SupplierInn"`

	Services []IncomeService `xml:"Services"`

	TotalAmount float64 `xml:"TotalAmount"`

	IncomeHashCode string `xml:"IncomeHashCode"`

	Link string `xml:"Link"`

	GeoInfo *GeoInfo `xml:"GeoInfo"`

	OperationUniqueId string `xml:"OperationUniqueId"`
}

// Element
type PostIncomeFromIndividualResponse struct {
	XMLName xml.Name `xml:"PostIncomeFromIndividualResponse"`

	ReceiptId string `xml:"ReceiptId"`

	Link string `xml:"Link"`
}

// Element
type PostCancelReceiptRequest struct {
	XMLName xml.Name `xml:"PostCancelReceiptRequest"`

	Inn string `xml:"Inn"`

	ReceiptId string `xml:"ReceiptId"`

	Message string `xml:"Message"`
}

// Element
type PostCancelReceiptResponse struct {
	XMLName xml.Name `xml:"PostCancelReceiptResponse"`

	RequestResult string `xml:",any"`
}

// Element
type PostCancelReceiptRequestV2 struct {
	XMLName xml.Name `xml:"PostCancelReceiptRequestV2"`

	Inn string `xml:"Inn"`

	ReceiptId string `xml:"ReceiptId"`

	ReasonCode string `xml:"ReasonCode"`
}

// Element
type PostCancelReceiptResponseV2 struct {
	XMLName xml.Name `xml:"PostCancelReceiptResponseV2"`

	RequestResult string `xml:",any"`
}

// Element
type GetIncomeRequest struct {
	XMLName xml.Name `xml:"GetIncomeRequest"`

	Inn string `xml:"Inn"`

	From string `xml:"From"`

	To string `xml:"To"`

	Limit *int `xml:"Limit"`

	Offset *int `xml:"Offset"`
}

// Element
type GetIncomeResponse struct {
	XMLName xml.Name `xml:"GetIncomeResponse"`

	HasMore bool `xml:"HasMore"`

	Receipts []Receipt `xml:"Receipts"`
}

// Element
type GetIncomeRequestV2 struct {
	XMLName xml.Name `xml:"GetIncomeRequestV2"`

	Inn string `xml:"Inn"`

	From string `xml:"From"`

	To string `xml:"To"`

	Limit *int `xml:"Limit"`

	Offset *int `xml:"Offset"`
}

// Element
type GetIncomeResponseV2 struct {
	XMLName xml.Name `xml:"GetIncomeResponseV2"`

	HasMore bool `xml:"HasMore"`

	Receipts []ReceiptV2 `xml:"Receipts"`
}

// Element
type GetIncomeForPeriodRequest struct {
	XMLName xml.Name `xml:"GetIncomeForPeriodRequest"`

	Inn string `xml:"Inn"`

	TaxPeriodId string `xml:"TaxPeriodId"`
}

// Element
type GetIncomeForPeriodResponse struct {
	XMLName xml.Name `xml:"GetIncomeForPeriodResponse"`

	TotalAmount float64 `xml:"TotalAmount"`

	CanceledTotalAmount float64 `xml:"CanceledTotalAmount"`

	Tax float64 `xml:"Tax"`
}

// Element
type GetTaxpayerRatingRequest struct {
	XMLName xml.Name `xml:"GetTaxpayerRatingRequest"`

	Inn string `xml:",any"`
}

// Element
type GetTaxpayerRatingResponse struct {
	XMLName xml.Name `xml:"GetTaxpayerRatingResponse"`

	Rating string `xml:",any"`
}

// Element
type PostRestrictionsRequest struct {
	XMLName xml.Name `xml:"PostRestrictionsRequest"`

	Inn string `xml:"Inn"`

	Type string `xml:"Type"`

	Message string `xml:"Message"`
}

// Element
type PostRestrictionsResponse struct {
	XMLName xml.Name `xml:"PostRestrictionsResponse"`

	Id string `xml:",any"`
}

// Element
type GetRestrictionsStatusRequest struct {
	XMLName xml.Name `xml:"GetRestrictionsStatusRequest"`

	Id string `xml:",any"`
}

// Element
type GetRestrictionsStatusResponse struct {
	XMLName xml.Name `xml:"GetRestrictionsStatusResponse"`

	RequestResult string `xml:"RequestResult"`

	Message string `xml:"Message"`

	ProcessingTime string `xml:"ProcessingTime"`
}

// Element
type GetKeysRequest struct {
	XMLName xml.Name `xml:"GetKeysRequest"`

	Inn []string `xml:",any"`
}

// Element
type GetKeysResponse struct {
	XMLName xml.Name `xml:"GetKeysResponse"`

	Keys []KeyInfo `xml:",any"`
}

// Element
type GetLegalEntityInfoRequest struct {
	XMLName xml.Name `xml:"GetLegalEntityInfoRequest"`

	Inn string `xml:"Inn"`

	Ogrn string `xml:"Ogrn"`

	Name string `xml:"Name"`

	Oktmo string `xml:"Oktmo"`
}

// Element
type GetLegalEntityInfoResponse struct {
	XMLName xml.Name `xml:"GetLegalEntityInfoResponse"`

	Inn string `xml:"Inn"`

	Ogrn string `xml:"Ogrn"`

	Name string `xml:"Name"`

	Address string `xml:"Address"`

	TerminationDate string `xml:"TerminationDate"`

	InvalidationDate string `xml:"InvalidationDate"`
}

// Element
type GetNewlyUnboundTaxpayersRequest struct {
	XMLName xml.Name `xml:"GetNewlyUnboundTaxpayersRequest"`

	From string `xml:"From"`

	To string `xml:"To"`

	Limit *int `xml:"Limit"`

	Offset *int `xml:"Offset"`
}

// Element
type GetNewlyUnboundTaxpayersResponse struct {
	XMLName xml.Name `xml:"GetNewlyUnboundTaxpayersResponse"`

	Taxpayers []NewlyUnboundTaxpayersInfo `xml:"Taxpayers"`

	HasMore bool `xml:"HasMore"`
}

// Element
type GetRegionsListRequest struct {
	XMLName xml.Name `xml:"GetRegionsListRequest"`

	RequestTime string `xml:",any"`
}

// Element
type GetRegionsListResponse struct {
	XMLName xml.Name `xml:"GetRegionsListResponse"`

	Regions []GetRegionsListResponseRegions `xml:",any"`
}

// Element
type GetActivitiesListRequest struct {
	XMLName xml.Name `xml:"GetActivitiesListRequest"`

	RequestTime string `xml:",any"`
}

// Element
type GetActivitiesListResponse struct {
	XMLName xml.Name `xml:"GetActivitiesListResponse"`

	Activities []GetActivitiesListResponseActivities `xml:",any"`
}

// Element
type GetActivitiesListRequestV2 struct {
	XMLName xml.Name `xml:"GetActivitiesListRequestV2"`

	RequestTime string `xml:",any"`
}

// Element
type GetActivitiesListResponseV2 struct {
	XMLName xml.Name `xml:"GetActivitiesListResponseV2"`

	Activities []GetActivitiesListResponseV2Activities `xml:",any"`
}

// Element
type PostNewActivityRequest struct {
	XMLName xml.Name `xml:"PostNewActivityRequest"`

	Activity string `xml:",any"`
}

// Element
type PostNewActivityResponse struct {
	XMLName xml.Name `xml:"PostNewActivityResponse"`

	Id string `xml:",any"`
}

// Element
type GetRejectionReasonsListRequest struct {
	XMLName xml.Name `xml:"GetRejectionReasonsListRequest"`

	RequestTime string `xml:",any"`
}

// Element
type GetRejectionReasonsListResponse struct {
	XMLName xml.Name `xml:"GetRejectionReasonsListResponse"`

	Codes []GetRejectionReasonsListResponseCodes `xml:",any"`
}

// Element
type GetUnregistrationReasonsListRequest struct {
	XMLName xml.Name `xml:"GetUnregistrationReasonsListRequest"`

	RequestTime string `xml:",any"`
}

// Element
type GetUnregistrationReasonsListResponse struct {
	XMLName xml.Name `xml:"GetUnregistrationReasonsListResponse"`

	Codes []GetUnregistrationReasonsListResponseCodes `xml:",any"`
}

// Element
type GetInnByPersonalInfoRequest struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/ais3/SMZ/SmzPartnersIntegrationService/types/1.0 GetInnByPersonalInfoRequest"`

	FirstName string `xml:"FirstName"`

	SecondName string `xml:"SecondName"`

	Patronymic string `xml:"Patronymic"`

	Birthday string `xml:"Birthday"`

	PassportSeries string `xml:"PassportSeries"`

	PassportNumber string `xml:"PassportNumber"`
}

// Element
type GetInnByPersonalInfoResponse struct {
	XMLName xml.Name `xml:"GetInnByPersonalInfoResponse"`

	Inn []string `xml:"Inn"`

	Status string `xml:"Status"`
}

// Element
type GetInnByPersonalInfoRequestV2 struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/ais3/SMZ/SmzPartnersIntegrationService/types/1.0 GetInnByPersonalInfoRequestV2"`

	PersonalInfoList []PersonalInfo `xml:",any"`
}

// Element
type GetInnByPersonalInfoResponseV2 struct {
	XMLName xml.Name `xml:"GetInnByPersonalInfoResponseV2"`

	InnList []InnByPersonalInfo `xml:",any"`
}

// Element
type GetInnByPersonalInfoRequestV3 struct {
	XMLName xml.Name `xml:"GetInnByPersonalInfoRequestV3"`

	PersonalInfoList []PersonalInfoV3 `xml:",any"`
}

// Element
type GetInnByPersonalInfoResponseV3 struct {
	XMLName xml.Name `xml:"GetInnByPersonalInfoResponseV3"`

	InnList []InnByPersonalInfo `xml:",any"`
}

// Element
type PostPlatformRegistrationRequest struct {
	XMLName xml.Name `xml:"urn://x-artefacts-gnivc-ru/ais3/SMZ/SmzPartnersIntegrationService/types/1.0 PostPlatformRegistrationRequest"`

	PartnerName string `xml:"PartnerName"`

	PartnerType string `xml:"PartnerType"`

	PartnerDescription string `xml:"PartnerDescription"`

	PartnerConnectable string `xml:"PartnerConnectable"`

	PartnerAvailableForBind *bool `xml:"PartnerAvailableForBind"`

	TransitionLink string `xml:"TransitionLink"`

	PartnersText string `xml:"PartnersText"`

	PartnerImage string `xml:"PartnerImage"`

	Inn string `xml:"Inn"`

	Phone string `xml:"Phone"`
}

// Element
type PostPlatformRegistrationResponse struct {
	XMLName xml.Name `xml:"PostPlatformRegistrationResponse"`

	PartnerID string `xml:"PartnerID"`

	RegistrationDate string `xml:"RegistrationDate"`
}

// Element
type GetRegistrationReferenceRequestV2 struct {
	XMLName xml.Name `xml:"GetRegistrationReferenceRequestV2"`

	Inn string `xml:"Inn"`

	RequestTime string `xml:"RequestTime"`

	RequestYear string `xml:"RequestYear"`
}

// Element
type GetRegistrationReferenceResponseV2 struct {
	XMLName xml.Name `xml:"GetRegistrationReferenceResponseV2"`

	RegistrationReferencePdf AttachedFile `xml:",any"`
}

// Element
type GetIncomeReferenceRequestV2 struct {
	XMLName xml.Name `xml:"GetIncomeReferenceRequestV2"`

	Inn string `xml:"Inn"`

	RequestTime string `xml:"RequestTime"`

	RequestYear string `xml:"RequestYear"`
}

// Element
type GetIncomeReferenceResponseV2 struct {
	XMLName xml.Name `xml:"GetIncomeReferenceResponseV2"`

	IncomeReferencePdf AttachedFile `xml:",any"`
}

// Element
type GetChangeInnHistoryRequest struct {
	XMLName xml.Name `xml:"GetChangeInnHistoryRequest"`

	Offset int64 `xml:"Offset"`

	Limit int `xml:"Limit"`
}

// Element
type GetChangeInnHistoryResponse struct {
	XMLName xml.Name `xml:"GetChangeInnHistoryResponse"`

	Items []GetChangeInnHistoryResponseItems `xml:",any"`
}

// Element
type GetGrantedPermissionsStatusRequest struct {
	XMLName xml.Name `xml:"GetGrantedPermissionsStatusRequest"`

	Id string `xml:",any"`
}

// Element
type GetGrantedPermissionsStatusResponse struct {
	XMLName xml.Name `xml:"GetGrantedPermissionsStatusResponse"`

	Inn string `xml:"Inn"`

	Result string `xml:"Result"`

	ProcessingTime string `xml:"ProcessingTime"`
}

// Element
type GetNotificationsRequest struct {
	XMLName xml.Name `xml:"GetNotificationsRequest"`

	NotificationsRequest []NotificationsRequest `xml:",any"`
}

// Element
type GetNotificationsResponse struct {
	XMLName xml.Name `xml:"GetNotificationsResponse"`

	NotificationsResponse []NotificationsResponse `xml:",any"`
}

// Element
type PostNotificationsAckRequest struct {
	XMLName xml.Name `xml:"PostNotificationsAckRequest"`

	NotificationList []NotificationsActionRequest `xml:",any"`
}

// Element
type PostNotificationsAckResponse struct {
	XMLName xml.Name `xml:"PostNotificationsAckResponse"`

	Status string `xml:",any"`
}

// Element
type PostNotificationsArchRequest struct {
	XMLName xml.Name `xml:"PostNotificationsArchRequest"`

	NotificationList []NotificationsActionRequest `xml:",any"`
}

// Element
type PostNotificationsArchResponse struct {
	XMLName xml.Name `xml:"PostNotificationsArchResponse"`

	Status string `xml:",any"`
}

// Element
type PostNotificationsAckAllRequest struct {
	XMLName xml.Name `xml:"PostNotificationsAckAllRequest"`

	Inn []string `xml:",any"`
}

// Element
type PostNotificationsAckAllResponse struct {
	XMLName xml.Name `xml:"PostNotificationsAckAllResponse"`

	Status string `xml:",any"`
}

// Element
type PostNotificationsArchAllRequest struct {
	XMLName xml.Name `xml:"PostNotificationsArchAllRequest"`

	Inn []string `xml:",any"`
}

// Element
type PostNotificationsArchAllResponse struct {
	XMLName xml.Name `xml:"PostNotificationsArchAllResponse"`

	Status string `xml:",any"`
}

// Element
type GetNotificationsCountRequest struct {
	XMLName xml.Name `xml:"GetNotificationsCountRequest"`

	Inn []string `xml:",any"`
}

// Element
type GetNotificationsCountResponse struct {
	XMLName xml.Name `xml:"GetNotificationsCountResponse"`

	Status []NotificationsCount `xml:",any"`
}

// Element
type PostNotificationsDeliveredRequest struct {
	XMLName xml.Name `xml:"PostNotificationsDeliveredRequest"`

	NotificationList []NotificationsActionRequest `xml:",any"`
}

// Element
type PostNotificationsDeliveredResponse struct {
	XMLName xml.Name `xml:"PostNotificationsDeliveredResponse"`

	Status string `xml:",any"`
}

// Element
type GetNewPermissionsChangeRequest struct {
	XMLName xml.Name `xml:"GetNewPermissionsChangeRequest"`

	Inn []string `xml:",any"`
}

// Element
type GetNewPermissionsChangeResponse struct {
	XMLName xml.Name `xml:"GetNewPermissionsChangeResponse"`

	Taxpayers []ChangePermissoinsInfo `xml:",any"`
}

// Element
type PostDecisionPermissionsChangeRequest struct {
	XMLName xml.Name `xml:"PostDecisionPermissionsChangeRequest"`

	RequestId string `xml:"requestId"`

	Inn string `xml:"inn"`

	Status string `xml:"status"`
}

// Element
type PostDecisionPermissionsChangeResponse struct {
	XMLName xml.Name `xml:"PostDecisionPermissionsChangeResponse"`

	Status string `xml:",any"`
}

// Element
type GetPartnersPermissionsRequest struct {
	XMLName xml.Name `xml:"GetPartnersPermissionsRequest"`

	Inn string `xml:",any"`
}

// Element
type GetPartnersPermissionsResponse struct {
	XMLName xml.Name `xml:"GetPartnersPermissionsResponse"`

	PartnersPermissionsList []PartnersAndPermissions `xml:",any"`
}

// Element
type GetAccrualsAndDebtsRequest struct {
	XMLName xml.Name `xml:"GetAccrualsAndDebtsRequest"`

	InnList []string `xml:",any"`
}

// Element
type GetAccrualsAndDebtsResponse struct {
	XMLName xml.Name `xml:"GetAccrualsAndDebtsResponse"`

	AccrualsAndDebtsList []AccrualsAndDebts `xml:",any"`
}

// Element
type GetPaymentDocumentsRequest struct {
	XMLName xml.Name `xml:"GetPaymentDocumentsRequest"`

	InnList []string `xml:",any"`
}

// Element
type GetPaymentDocumentsResponse struct {
	XMLName xml.Name `xml:"GetPaymentDocumentsResponse"`

	PaymentDocumentsList []PaymentDocumentList `xml:",any"`
}

// Element
type GetCancelIncomeReasonsListRequest struct {
	XMLName xml.Name `xml:"GetCancelIncomeReasonsListRequest"`

	RequestTime string `xml:",any"`
}

// Element
type GetCancelIncomeReasonsListResponse struct {
	XMLName xml.Name `xml:"GetCancelIncomeReasonsListResponse"`

	Codes []GetCancelIncomeReasonsListResponseCodes `xml:",any"`
}

// Element
type GetTaxpayerUnregistrationReasonsListRequest struct {
	XMLName xml.Name `xml:"GetTaxpayerUnregistrationReasonsListRequest"`

	RequestTime string `xml:",any"`
}

// Element
type GetTaxpayerUnregistrationReasonsListResponse struct {
	XMLName xml.Name `xml:"GetTaxpayerUnregistrationReasonsListResponse"`

	Codes []GetTaxpayerUnregistrationReasonsListResponseCodes `xml:",any"`
}

// Element
type SmzPlatformErrorArgs struct {
	XMLName xml.Name `xml:"Args"`

	Key string `xml:"Key"`

	Value string `xml:"Value"`
}

// Element
type GetRegionsListResponseRegions struct {
	XMLName xml.Name `xml:"Regions"`

	Oktmo string `xml:"Oktmo"`

	Name string `xml:"Name"`
}

// Element
type GetActivitiesListResponseActivities struct {
	XMLName xml.Name `xml:"Activities"`

	Id string `xml:"Id"`

	Name string `xml:"Name"`
}

// Element
type GetActivitiesListResponseV2Activities struct {
	XMLName xml.Name `xml:"Activities"`

	Id int `xml:"Id"`

	ParentId *int `xml:"ParentId"`

	Name string `xml:"Name"`

	IsActive bool `xml:"IsActive"`
}

// Element
type GetRejectionReasonsListResponseCodes struct {
	XMLName xml.Name `xml:"Codes"`

	Code string `xml:"Code"`

	Description string `xml:"Description"`
}

// Element
type GetUnregistrationReasonsListResponseCodes struct {
	XMLName xml.Name `xml:"Codes"`

	Code string `xml:"Code"`

	Description string `xml:"Description"`
}

// Element
type GetChangeInnHistoryResponseItems struct {
	XMLName xml.Name `xml:"Items"`

	Offset int64 `xml:"Offset"`

	PreviousInn string `xml:"PreviousInn"`

	Inn string `xml:"Inn"`

	From string `xml:"From"`

	To string `xml:"To"`
}

// Element
type GetCancelIncomeReasonsListResponseCodes struct {
	XMLName xml.Name `xml:"Codes"`

	Code string `xml:"Code"`

	Description string `xml:"Description"`
}

// Element
type GetTaxpayerUnregistrationReasonsListResponseCodes struct {
	XMLName xml.Name `xml:"Codes"`

	Code string `xml:"Code"`

	Description string `xml:"Description"`
}

// Element
type KeyRecord struct {
	XMLName xml.Name `xml:"KeyRecord"`

	SequenceNumber int `xml:"SequenceNumber"`

	ExpireTime string `xml:"ExpireTime"`

	Base64Key string `xml:"Base64Key"`
}

// XSD ComplexType declarations

type IncomeService struct {
	XMLName xml.Name

	Amount float64 `xml:"Amount"`

	Name string `xml:"Name"`

	Quantity int64 `xml:"Quantity"`

	InnerXml string `xml:",innerxml"`
}

type Receipt struct {
	XMLName xml.Name

	Link string `xml:"Link"`

	TotalAmount float64 `xml:"TotalAmount"`

	ReceiptId string `xml:"ReceiptId"`

	RequestTime string `xml:"RequestTime"`

	OperationTime string `xml:"OperationTime"`

	PartnerCode string `xml:"PartnerCode"`

	CancelationTime string `xml:"CancelationTime"`

	Services IncomeService `xml:"Services"`

	InnerXml string `xml:",innerxml"`
}

type ReceiptV2 struct {
	XMLName xml.Name

	Link string `xml:"Link"`

	TotalAmount float64 `xml:"TotalAmount"`

	ReceiptId string `xml:"ReceiptId"`

	IncomeType string `xml:"IncomeType"`

	RequestTime string `xml:"RequestTime"`

	OperationTime string `xml:"OperationTime"`

	TaxPeriodId string `xml:"TaxPeriodId"`

	TaxToPay float64 `xml:"TaxToPay"`

	PartnerCode string `xml:"PartnerCode"`

	SupplierInn string `xml:"SupplierInn"`

	CancelationTime string `xml:"CancelationTime"`

	Services []IncomeService `xml:"Services"`

	InnerXml string `xml:",innerxml"`
}

type GeoInfo struct {
	XMLName xml.Name

	Latitude float64 `xml:"Latitude"`

	Longitude float64 `xml:"Longitude"`

	InnerXml string `xml:",innerxml"`
}

type NewlyUnboundTaxpayersInfo struct {
	XMLName xml.Name

	Inn string `xml:"Inn"`

	FirstName string `xml:"FirstName"`

	SecondName string `xml:"SecondName"`

	Patronymic string `xml:"Patronymic"`

	UnboundTime string `xml:"UnboundTime"`

	RegistrationTime string `xml:"RegistrationTime"`

	Phone string `xml:"Phone"`

	InnerXml string `xml:",innerxml"`
}

type KeyInfo struct {
	XMLName xml.Name

	Inn string `xml:"Inn"`

	KeyRecord []KeyRecord `xml:"KeyRecord"`

	InnerXml string `xml:",innerxml"`
}

type GetTaxpayerAccountStatusInfo struct {
	XMLName xml.Name

	Id string `xml:"Id"`

	PaymentTime string `xml:"PaymentTime"`

	PaymentReceivedTime string `xml:"PaymentReceivedTime"`

	RejectionReason string `xml:"RejectionReason"`

	Amount float64 `xml:"Amount"`

	TaxPaymentTime string `xml:"TaxPaymentTime"`

	PenaltyAmount float64 `xml:"PenaltyAmount"`

	TaxBonus float64 `xml:"TaxBonus"`

	Surplus float64 `xml:"Surplus"`

	ReportTime string `xml:"ReportTime"`

	InnerXml string `xml:",innerxml"`
}

type AttachedFile struct {
	XMLName xml.Name

	Mimetype string `xml:"mimetype"`

	Filename string `xml:"filename"`

	Content string `xml:"content"`

	InnerXml string `xml:",innerxml"`
}

type PersonalInfo struct {
	XMLName xml.Name

	FirstName string `xml:"FirstName"`

	SecondName string `xml:"SecondName"`

	Patronymic string `xml:"Patronymic"`

	Birthday string `xml:"Birthday"`

	PassportSeries string `xml:"PassportSeries"`

	PassportNumber string `xml:"PassportNumber"`

	InnerXml string `xml:",innerxml"`
}

type PersonalInfoV3 struct {
	XMLName xml.Name

	FirstName string `xml:"FirstName"`

	SecondName string `xml:"SecondName"`

	Patronymic string `xml:"Patronymic"`

	Birthday string `xml:"Birthday"`

	DocumentSpdul string `xml:"DocumentSpdul"`

	DocumentSeries string `xml:"DocumentSeries"`

	DocumentNumber string `xml:"DocumentNumber"`

	InnerXml string `xml:",innerxml"`
}

type InnByPersonalInfo struct {
	XMLName xml.Name

	Inn []string `xml:"Inn"`

	Status string `xml:"Status"`

	InnerXml string `xml:",innerxml"`
}

type NotificationsRequest struct {
	XMLName xml.Name

	Inn string `xml:"inn"`

	GetAcknowleged *bool `xml:"GetAcknowleged"`

	GetArchived *bool `xml:"GetArchived"`

	InnerXml string `xml:",innerxml"`
}

type NotificationsResponse struct {
	XMLName xml.Name

	Inn string `xml:"inn"`

	Notif []Notifications `xml:"notif"`

	InnerXml string `xml:",innerxml"`
}

type Notifications struct {
	XMLName xml.Name

	Id string `xml:"id"`

	Title string `xml:"title"`

	Message string `xml:"message"`

	Status string `xml:"status"`

	CreatedAt string `xml:"createdAt"`

	UpdatedAt string `xml:"updatedAt"`

	PartnerId string `xml:"partnerId"`

	ApplicationId string `xml:"applicationId"`

	InnerXml string `xml:",innerxml"`
}

type NotificationsActionRequest struct {
	XMLName xml.Name

	Inn string `xml:"inn"`

	MessageId []string `xml:"messageId"`

	InnerXml string `xml:",innerxml"`
}

type NotificationsCount struct {
	XMLName xml.Name

	Inn string `xml:"inn"`

	Count *int `xml:"count"`

	InnerXml string `xml:",innerxml"`
}

type PartnersAndPermissions struct {
	XMLName xml.Name

	PartnerId string `xml:"PartnerId"`

	PartnerName string `xml:"PartnerName"`

	PartnerBindStatus string `xml:"PartnerBindStatus"`

	BindTime string `xml:"BindTime"`

	PermissionsChangeTime string `xml:"PermissionsChangeTime"`

	PermissionsList []string `xml:"PermissionsList"`

	InnerXml string `xml:",innerxml"`
}

type ChangePermissoinsInfo struct {
	XMLName xml.Name

	Inn string `xml:"Inn"`

	RequestId string `xml:"requestId"`

	RequestPartnerId string `xml:"requestPartnerId"`

	PartnerName string `xml:"PartnerName"`

	PermissionsList []string `xml:"PermissionsList"`

	RequestTime string `xml:"RequestTime"`

	InnerXml string `xml:",innerxml"`
}

type TaxCharge struct {
	XMLName xml.Name

	Amount float64 `xml:"Amount"`

	DueDate string `xml:"DueDate"`

	TaxPeriodId int `xml:"TaxPeriodId"`

	Oktmo string `xml:"Oktmo"`

	Kbk string `xml:"Kbk"`

	PaidAmount float64 `xml:"PaidAmount"`

	CreateTime string `xml:"CreateTime"`

	Id int64 `xml:"Id"`

	InnerXml string `xml:",innerxml"`
}

type Krsb struct {
	XMLName xml.Name

	Debt float64 `xml:"Debt"`

	Penalty float64 `xml:"Penalty"`

	Overpayment float64 `xml:"Overpayment"`

	Oktmo string `xml:"Oktmo"`

	Kbk string `xml:"Kbk"`

	TaxOrganCode string `xml:"TaxOrganCode"`

	UpdateTime string `xml:"UpdateTime"`

	Id int64 `xml:"Id"`

	InnerXml string `xml:",innerxml"`
}

type AccrualsAndDebts struct {
	XMLName xml.Name

	Inn string `xml:"Inn"`

	TaxChargeList []TaxCharge `xml:"TaxChargeList"`

	KrsbList []Krsb `xml:"KrsbList"`

	InnerXml string `xml:",innerxml"`
}

type PaymentDocument struct {
	XMLName xml.Name

	Type string `xml:"Type"`

	DocumentIndex string `xml:"DocumentIndex"`

	FullName string `xml:"FullName"`

	Address string `xml:"Address"`

	Inn string `xml:"Inn"`

	Amount float64 `xml:"Amount"`

	RecipientBankName string `xml:"RecipientBankName"`

	RecipientBankBik string `xml:"RecipientBankBik"`

	RecipientBankAccountNumber string `xml:"RecipientBankAccountNumber"`

	Recipient string `xml:"Recipient"`

	RecipientAccountNumber string `xml:"RecipientAccountNumber"`

	RecipientInn string `xml:"RecipientInn"`

	RecipientKpp string `xml:"RecipientKpp"`

	Kbk string `xml:"Kbk"`

	Oktmo string `xml:"Oktmo"`

	Code101 string `xml:"Code101"`

	Code106 string `xml:"Code106"`

	Code107 string `xml:"Code107"`

	Code110 string `xml:"Code110"`

	DueDate string `xml:"DueDate"`

	CreateTime string `xml:"CreateTime"`

	SourceId *int64 `xml:"SourceId"`

	InnerXml string `xml:",innerxml"`
}

type PaymentDocumentList struct {
	XMLName xml.Name

	Inn string `xml:"Inn"`

	DocumentList []PaymentDocument `xml:"DocumentList"`

	InnerXml string `xml:",innerxml"`
}

// XSD SimpleType declarations
