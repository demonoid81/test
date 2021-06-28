package resolvers

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sphera-erp/sphera/internal/models"
	"moul.io/http2curl"
)

func (r *mutationResolver) PersonMutation(ctx context.Context, person *models.Person) (*models.Person, error) {
	return r.persons.PersonaMutation(ctx, person)
}

func (r *mutationResolver) PersonCourseMutation(ctx context.Context, course *models.PersonCourse) (*models.PersonCourse, error) {
	return r.persons.PersonCourseMutation(ctx, course)
}

func (r *mutationResolver) PersonRatingMutation(ctx context.Context, personRating *models.PersonRating) (*models.PersonRating, error) {
	return r.persons.PersonRatingMutation(ctx, personRating)
}

func (r *mutationResolver) Agreement(ctx context.Context, incomeRegistration bool, taxPayment bool) (bool, error) {
	return r.persons.Agreement(ctx, incomeRegistration, taxPayment)
}

func (r *mutationResolver) RemoveContact(ctx context.Context, person *models.Person, contact *models.Contact) (bool, error) {
	return r.persons.RemoveContact(ctx, person, contact)
}

func (r *mutationResolver) ReqToPartner(ctx context.Context) (bool, error) {
	return r.persons.ReqToPartner(ctx)
}

func (r *queryResolver) Person(ctx context.Context, person models.Person) (*models.Person, error) {
	return r.persons.Person(ctx, person)
}

func (r *queryResolver) Persons(ctx context.Context, person *models.Person, filter *models.PersonFilter, sort []models.PersonSort, offset *int, limit *int) ([]*models.Person, error) {
	return r.persons.Persons(ctx, person, filter, sort, offset, limit)
}

func (r *queryResolver) ParsePerson(ctx context.Context, photo *graphql.Upload, passport *graphql.Upload) (*models.PersonValidateStatus, error) {
	return r.persons.ParsePerson(ctx, photo, passport)
}

func (r *queryResolver) ValidateInn(ctx context.Context, inn string) (*bool, error) {
	req, _ := validateINNPostReq(inn)

	command, _ := http2curl.GetCurlCommand(req)
	fmt.Println(command)

	rClient := &http.Client{}

	response, err := rClient.Do(req)

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("err req")
	}
	defer response.Body.Close()

	fmt.Println(response.Body)
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("err req")
	}
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(data, &jsonMap)
	if err != nil {
		return nil, fmt.Errorf("err req")
	}

	res := jsonMap["status"].(bool)

	return &res, nil
}

func (r *queryResolver) PersonCourses(ctx context.Context, course *models.PersonCourse) ([]*models.PersonCourse, error) {
	return r.persons.PersonCourses(ctx, course)
}

func (r *queryResolver) GetPersonRating(ctx context.Context, person models.Person) (*float64, error) {
	return r.persons.GetPersonRating(ctx, person)
}

func (r *queryResolver) GetMyRating(ctx context.Context) (*float64, error) {
	return r.persons.GetMyRating(ctx)
}

func (r *queryResolver) PersonRating(ctx context.Context, personRating *models.PersonRating) (*models.PersonRating, error) {
	return r.persons.PersonRating(ctx, personRating)
}

func (r *queryResolver) PersonRatings(ctx context.Context, personRating *models.PersonRating, offset *int, limit *int) ([]*models.PersonRating, error) {
	return r.persons.PersonRatings(ctx, personRating, offset, limit)
}

func (r *queryResolver) GetSelfEmployerStatus(ctx context.Context) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) PersonTax(ctx context.Context) (*models.Taxes, error) {
	return r.persons.PersonTax(ctx)
}

func (r *subscriptionResolver) ParsePersonSub(ctx context.Context) (<-chan *models.RecognizedFields, error) {
	return r.persons.ParsePersonSub(ctx)
}

func (r *subscriptionResolver) PersonSub(ctx context.Context) (<-chan *models.Person, error) {
	return r.persons.PersonSub(ctx)
}

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
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
