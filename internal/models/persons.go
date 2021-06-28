package models

import (
	"context"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"io"
	"reflect"
	"strconv"
	"time"
)

type Gender string

type Person struct {
	UUID                 *uuid.UUID             `json:"uuid" db:"uuid"`
	Created              *time.Time             `json:"created" db:"created"`
	Updated              *time.Time             `json:"updated" db:"updated"`
	UUIDUser             *uuid.UUID             `db:"uuid_user"`
	User                 *User                  `json:"user" relay:"uuid_user" link:"UUIDUser"`
	UUIDActualContact    *uuid.UUID             `db:"uuid_actual_contact"`
	ActualContact        *Contact               `json:"actualContact" relay:"uuid_actual_contact" link:"UUIDActualContact"`
	UUIDContacts         []*uuid.UUID           `db:"uuid_contacts"`
	Contacts             []*Contact             `json:"contacts" relay:"uuid_contacts" link:"UUIDContacts"`
	UUIDPassport         *uuid.UUID             `db:"uuid_passport"`
	Passport             *Passport              `json:"passport" relay:"uuid_passport" link:"UUIDPassport"`
	Surname              *string                `json:"surname" db:"surname"`
	Name                 *string                `json:"name" db:"name"`
	Patronymic           *string                `json:"patronymic" db:"patronymic"`
	BirthDate            *time.Time             `json:"birthDate" db:"birth_date"`
	Gender               *Gender                `json:"gender" db:"gender"`
	UUIDCountry          *uuid.UUID             `db:"uuid_country"`
	Country              *Country               `json:"country" relay:"uuid_country" link:"UUIDCountry"`
	INN                  *string                `json:"inn" db:"inn"`
	UUIDMedicalBook      *uuid.UUID             `db:"uuid_medical_book"`
	MedicalBook          *MedicalBook           `json:"medicalBook" relay:"uuid_medical_book" link:"UUIDMedicalBook" `
	UUIDPhoto            *uuid.UUID             `db:"uuid_photo"`
	Photo                *Content               `json:"photo" relay:"uuid_photo" link:"UUIDPhoto"`
	UUIDPosition         *uuid.UUID             `db:"uuid_position"`
	Position             *OrganizationPosition  `json:"position" relay:"uuid_position" link:"UUIDPosition"`
	IsDeleted            *bool                  `json:"isDeleted" db:"is_deleted"`
	IsContact            *bool                  `json:"isContact" db:"is_contact"`
	RecognizeResult      map[string]interface{} `json:"recognizeResult" db:"recognize_result"`
	DistanceResult       map[string]interface{} `json:"distanceResult" db:"distance_result"`
	RecognizedFieldsJSON map[string]interface{} `db:"recognized_fields"`
	RecognizedFields     *RecognizedFields      `json:"recognizedFields" link:"RecognizedFieldsJSON"`
	Reward               *float64               `db:"reward"`
	Secondary            *bool                  `db:"secondary"`
	Validated            *bool                  `json:"validated" db:"validated"`
	Rating               *float64               `json:"rating"`
	TaxPayment           *bool                  `json:"taxPayment" db:"tax_payment"`
	IncomeRegistration   *bool                  `json:"incomeRegistration" db:"income_registration"`
}

type PersonSort struct {
	Field *UserSortableField `json:"field"`
	Order *SortOrder         `json:"order"`
}

type PersonValidateStatus struct {
	Passport bool `json:"passport"`
	Avatar   bool `json:"avatar"`
}

type PersonFilter struct {
	Created    *DateTimeFilter `json:"created" db:"created"`
	Updated    *DateTimeFilter `json:"updated" db:"updated"`
	Surname    *StringFilter   `json:"surname" db:"surname"`
	Name       *StringFilter   `json:"name" db:"name"`
	Patronymic *StringFilter   `json:"patronymic" db:"patronymic"`
	BirthDate  *DateFilter     `json:"birthDate" db:"birth_date"`
	IsContact  *bool           `json:"isContact" db:"is_contact"`
	INN        *StringFilter   `json:"inn" db:"inn"`
	And        []*PersonFilter `json:"and"`
	Or         []*PersonFilter `json:"or"`
	Not        *PersonFilter   `json:"not"`
}

type RecognizedField struct {
	Result     string  `json:"result"`
	Confidence float64 `json:"confidence"`
	Valid      bool    `json:"valid"`
}

type RecognizedFields struct {
	Error          *string          `json:"error"`
	Surname        *RecognizedField `json:"surname"`
	Name           *RecognizedField `json:"name"`
	Patronymic     *RecognizedField `json:"patronymic"`
	BirthDate      *RecognizedField `json:"birthDate"`
	Gender         *RecognizedField `json:"gender"`
	Serial         *RecognizedField `json:"serial"`
	Number         *RecognizedField `json:"number"`
	Department     *RecognizedField `json:"department"`
	DateIssue      *RecognizedField `json:"dateIssue"`
	DepartmentCode *RecognizedField `json:"departmentCode"`
}

const (
	Male   Gender = "Male"
	Female Gender = "Female"
)

func (e Gender) IsValid() bool {
	switch e {
	case Male, Female:
		return true
	}
	return false
}

func (e Gender) String() string {
	return string(e)
}

func (e *Gender) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Gender(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid gender", str)
	}
	return nil
}

func (e Gender) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type PersonSortableField string

const (
	PersonSortableFieldUUID       PersonSortableField = "uuid"
	PersonSortableFieldCreated    PersonSortableField = "created"
	PersonSortableFieldUpdated    PersonSortableField = "updated"
	PersonSortableFieldSurname    PersonSortableField = "surname"
	PersonSortableFieldName       PersonSortableField = "name"
	PersonSortableFieldPatronymic PersonSortableField = "patronymic"
	PersonSortableFieldBirthDate  PersonSortableField = "birthDate"
	PersonSortableFieldInn        PersonSortableField = "inn"
)

func (e PersonSortableField) IsValid() bool {
	switch e {
	case PersonSortableFieldUUID, PersonSortableFieldCreated, PersonSortableFieldUpdated, PersonSortableFieldSurname, PersonSortableFieldName, PersonSortableFieldPatronymic, PersonSortableFieldBirthDate, PersonSortableFieldInn:
		return true
	}
	return false
}

func (e PersonSortableField) String() string {
	return string(e)
}

func (e *PersonSortableField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PersonSortableField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PersonSortableField", str)
	}
	return nil
}

func (e PersonSortableField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (p *Person) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	// Уберем лишние колонки
	if len(columns) > 0 {
		columns = utils.ClearSQLFields(p, columns)
	}
	// если есть uuid значит манипулируем обектом
	if p.UUID != nil {
		if utils.CountFillFields(p) == 1 && len(columns) == 0 {
			return nil, p.UUID, nil
		}
		// востановим объект
		person, err := p.GetByUUID(ctx, app, db, p.UUID)

		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// Если не меняются родители то вернем uuid
		if Compare(person, columns) && utils.CountFillFields(p) == 1 {
			return nil, p.UUID, nil
		}
		// востановим все ссылки
		utils.RestoreUUID(p, person)
		// востановим подчиненные структуры
		if err = p.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		p.UUID = &newUUID
		_, err := pglxqb.Insert("persons").
			Columns("uuid").
			Values(newUUID).
			RunWith(db).Exec(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error insert user")
			return nil, nil, gqlerror.Errorf("Error insert user")
		}
	}
	parentColumns := map[string]interface{}{"uuid_person": p.UUID}
	// дополним пропущенные поля, если они есть
	setColumns, err := SqlGenKeys(ctx, app, db, p, columns, parentColumns)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(p, setColumns)
	if len(setColumns) > 0 {
		// todo Логика Обновления
		if value, ok := setColumns["is_contact"]; ok && p.UUIDUser != nil {
			fmt.Println("Update Logic")
			var UUIDPerson []*uuid.UUID
			var UUIDOrganization uuid.UUID
			if err := pglxqb.Select("organizations.uuid_persons, organizations.uuid").
				From("organizations").
				LeftJoin("users u on organizations.uuid = u.uuid_organization").
				Where(pglxqb.Eq{"u.uuid": p.UUIDUser}).
				RunWith(db).
				QueryRow(ctx).
				Scan(&UUIDPerson, &UUIDOrganization); err != nil {
				if err.Error() != "no rows in result set" {
					app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
					return nil, nil, err
				}
			} else {
				if UUIDOrganization != uuid.Nil {
					if value.(bool) {
						UUIDPerson = append(UUIDPerson, p.UUID)
					} else {
						for i, vUUID := range UUIDPerson {
							if vUUID == p.UUID {
								UUIDPerson = utils.RemoveUUIDIndex(UUIDPerson, i)
							}
						}
					}
					if _, err := pglxqb.Update("organizations").
						Set("uuid_persons", UUIDPerson).
						Where("uuid = ?", UUIDOrganization).
						Suffix(utils.PrepareSuffix(rColumns)).
						RunWith(db).Exec(ctx); err != nil {
						app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
						return nil, nil, err
					}
				}
			}
		}
		if rows, err := pglxqb.Update("persons").
			SetMap(setColumns).
			Where("uuid = ?", p.UUID).
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error update person")
			return nil, nil, gqlerror.Errorf("Error update person")
		} else {
			return rows, p.UUID, nil
		}
	}
	return nil, p.UUID, nil
}

func (p *Person) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Person, error) {
	var person Person
	for rows.Next() {
		if err := rows.StructScan(&person); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
	}
	if err := person.ParseRequestedFields(ctx, fields, app, db); err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
		return nil, gqlerror.Errorf("Error scan response to struct person")
	}
	if person.UUID != nil {

		var rating *float64
		if err := pglxqb.Select("sum(rating)/(count(uuid)::float)::float as rating").From("person_ratings").
			Where(pglxqb.Eq{"uuid_person": person.UUID}).
			RunWith(db).QueryRow(ctx).Scan(&rating); err != nil {
			app.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error Select person from user ")
			return nil, gqlerror.Errorf("Error Select person from user")
		}
		person.Rating = rating
	}
	return &person, nil
}

func (p *Person) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Person, error) {
	var persons []*Person
	defer rows.Close()
	for rows.Next() {
		var person Person
		if err := rows.StructScan(&person); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}

		persons = append(persons, &person)
	}
	for _, person := range persons {
		if err := person.ParseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct person")
			return nil, gqlerror.Errorf("Error scan response to struct person")
		}
	}
	return persons, nil
}

func (p *Person) ParseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, p)
}

func (p *Person) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(p)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if !v.IsValid() {
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		if err := restoreStructReflect(ctx, app, db, v, v.Field(i), v.Type().Field(i)); err != nil {
			return err
		}
	}
	return nil
}

func (p *Person) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Person, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	count := 0
	var person Person
	for rows.Next() {
		count++
		if err := rows.StructScan(&person); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	if count == 0 {
		return nil, gqlerror.Errorf("Error no person found by UUID")
	}
	return &person, nil
}

func (p *Person) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Person, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return p.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (p *Person) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Person, error) {
	rows, err := pglxqb.SelectAll().From("persons").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return p.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
