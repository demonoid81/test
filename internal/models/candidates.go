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

type CandidateTags string

const (
	Primary    CandidateTags = "primary"
	Secondary  CandidateTags = "secondary"
	Refused    CandidateTags = "refused"
	Rejected   CandidateTags = "rejected"
	NotConfirm CandidateTags = "notConfirm"
	NoTraining CandidateTags = "noTraining"
)

func (e CandidateTags) IsValid() bool {
	switch e {
	case Primary, Secondary, Refused, Rejected, NotConfirm, NoTraining:
		return true
	}
	return false
}

func (e CandidateTags) String() string {
	return string(e)
}

func (e CandidateTags) Point() *CandidateTags {
	return &e
}

func (e *CandidateTags) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}
	*e = CandidateTags(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid gender", str)
	}
	return nil
}

func (e CandidateTags) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type Candidate struct {
	UUID       *uuid.UUID `json:"uuid" db:"uuid"`
	Created    *time.Time `json:"created" db:"created"`
	Updated    *time.Time `json:"updated" db:"updated"`
	UUIDPerson *uuid.UUID `db:"uuid_person"`
	Person     *Person    `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	UUIDJob    *uuid.UUID `db:"uuid_job"`
	Job        *Job       `json:"job" relay:"uuid_job" link:"UUIDJob"`
	//"отказался", "отклоненили", "не подтвердил", "нет обучения"
	CandidateTags *CandidateTags `json:"candidateTag" db:"candidate_tag"`
	IsDeleted     *bool          `json:"isDeleted" db:"is_deleted"`
}

func (c *Candidate) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	if c.UUID != nil {
		candidate, err := c.GetByUUID(ctx, app, db, c.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// востановим все ссылки
		utils.RestoreUUID(c, candidate)
		// востановим подчиненные структуры
		if err = candidate.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		updateOrDelete = true
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		c.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// дополним пропущеные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, c, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(c, setColumns)
	delete(setColumns, "is_deleted")
	if len(setColumns) > 0 {
		if updateOrDelete {
			// Обновляем иначе
			rows, err := pglxqb.Update("candidates").
				SetMap(setColumns).
				Where("uuid = ?", c.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
				return nil, nil, gqlerror.Errorf("Error update contact")
			}
			return rows, c.UUID, nil
		} else {
			rows, err := pglxqb.Insert("candidates").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, nil, gqlerror.Errorf("Error insert contact")
			}
			return rows, c.UUID, nil
		}
	}
	return nil, c.UUID, nil
}

func (c *Candidate) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Candidate, error) {
	var candidates []*Candidate
	defer rows.Close()
	for rows.Next() {
		var candidate Candidate
		if err := rows.StructScan(&candidate); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}

		candidates = append(candidates, &candidate)
	}
	for _, candidate := range candidates {
		if err := candidate.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return candidates, nil
}

func (c *Candidate) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Candidate, error) {
	var err error
	var candidate Candidate
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&candidate)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	err = candidate.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &candidate, nil
}

func (c *Candidate) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, c)
}

func (c *Candidate) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(c)
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

func (c *Candidate) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Candidate, error) {
	rows, err := pglxqb.SelectAll().From("candidates").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var candidate Candidate
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&candidate); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &candidate, nil
}

func (c *Candidate) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Candidate, error) {
	rows, err := pglxqb.SelectAll().From("candidates").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (c *Candidate) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Candidate, error) {
	rows, err := pglxqb.SelectAll().From("candidates").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return c.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
