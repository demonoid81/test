package addresses

import (
	"context"
	"fmt"
	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/ekomobile/dadata/v2/client"
	"github.com/sphera-erp/sphera/app"
	"net/url"
)

func parser (raw string, app *app.App) ([]*string, error) {

	fmt.Println("raw: ", raw )
	var err error
	endpointUrl, err := url.Parse("https://suggestions.dadata.ru/suggestions/api/4_1/rs/")
	if err != nil {
		return nil, err
	}

	creds := client.Credentials{
		ApiKeyValue:    app.Cfg.Api.DadataApi,
		SecretKeyValue: app.Cfg.Api.DadataSecret,
	}

	api := suggest.Api{
		Client: client.NewClient(endpointUrl, client.WithCredentialProvider(&creds)),
	}

	params := suggest.RequestParams{
		Query: raw,
	}

	suggestions, err := api.Address(context.Background(), &params)
	if err != nil {
		return nil, err
	}

	var result []*string
	for _, s := range suggestions {
		result = append(result, &s.Value)
		fmt.Printf("Result: %s - %s - %s\n", s.Data.Source, s.Data.Result, s.Value)
		fmt.Printf("RegionFiasID: %s - %s\n", s.Data.RegionFiasID, s.Data.Region)
		fmt.Printf("AreaFiasID: %s - %s\n", s.Data.AreaFiasID, s.Data.Area)
		fmt.Printf("CityFiasID: %s - %s\n", s.Data.CityFiasID, s.Data.City)
		fmt.Printf("CityDistrictFiasID: %s - %s\n", s.Data.CityDistrictFiasID, s.Data.CityDistrict)
		fmt.Printf("SettlementFiasID: %s - %s\n", s.Data.SettlementFiasID, s.Data.Settlement)
		fmt.Printf("StreetFiasID: %s - %s\n", s.Data.StreetFiasID, s.Data.Street)
	}
	return result, nil
}

