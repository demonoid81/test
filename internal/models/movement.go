package models

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Movement struct {
	UUID             *uuid.UUID           `json:"uuid" db:"uuid"`
	Created          *time.Time           `json:"created" db:"created"`
	Updated          *time.Time           `json:"updated" db:"updated"`
	IsDeleted        *bool                `json:"isDeleted" db:"is_deleted"`
	UUIDOrganization *uuid.UUID           `db:"uuid_organization"`
	Organization     *Organization        `json:"organization" relay:"uuid_organization" link:"UUIDOrganization"`
	Direct           *Direct              `json:"direct" db:"direct"`
	Destination      *DestinationMovement `json:"destination" db:"destination"`
	UUIDPerson       *uuid.UUID           `db:"uuid_person"`
	Person           *Person              `json:"person" relay:"uuid_person" link:"UUIDPerson"`
	UUIDJob          *uuid.UUID           `db:"uuid_job"`
	Job              *Job                 `json:"job" relay:"uuid_job" link:"UUIDJob"`
	Amount           *float64             `json:"amount" db:"amount"`
	Link             *string              `json:"Link" db:"link"`
}

type MovementFilter struct {
	UUID         *UUIDFilter          `json:"uuid"`
	IsDeleted    *bool                `json:"isDeleted"`
	Organization *OrganizationFilter  `json:"organization"`
	Direct       *Direct              `json:"direct"`
	Destination  *DestinationMovement `json:"destination"`
	Person       *PersonFilter        `json:"person"`
	Job          *JobFilter           `json:"base"`
	Amount       *float64             `json:"amount"`
	And          []*MovementFilter    `json:"and"`
	Or           []*MovementFilter    `json:"or"`
	Not          *MovementFilter      `json:"not"`
}

type MovementInput struct {
	UUID         *uuid.UUID           `json:"uuid"`
	IsDeleted    *bool                `json:"isDeleted"`
	Organization *Organization        `json:"organization"`
	Direct       *Direct              `json:"direct"`
	Destination  *DestinationMovement `json:"destination"`
	Person       *Person              `json:"person"`
	Job          *Job                 `json:"base"`
	Amount       *float64             `json:"amount"`
}

type DestinationMovement string

const (
	DestinationMovementSelfEmployer DestinationMovement = "selfEmployer"
	DestinationMovementTaxing       DestinationMovement = "taxing"
	DestinationMovementCommission   DestinationMovement = "commission"
	DestinationMovementReward       DestinationMovement = "reward"
	DestinationMovementRewardTax    DestinationMovement = "rewardTax"
)

func (e DestinationMovement) IsValid() bool {
	switch e {
	case DestinationMovementSelfEmployer,
		DestinationMovementTaxing,
		DestinationMovementCommission,
		DestinationMovementReward,
		DestinationMovementRewardTax:
		return true
	}
	return false
}

func (e DestinationMovement) String() string {
	return string(e)
}

func (e *DestinationMovement) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = DestinationMovement(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid DestinationMovement", str)
	}
	return nil
}

func (e DestinationMovement) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func (m *Movement) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	updateOrDelete := false
	if m.UUID != nil {
		movement, err := m.GetByUUID(ctx, app, db, m.UUID)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error get person")
			return nil, nil, gqlerror.Errorf("Error get person")
		}
		// восстановим все ссылки
		utils.RestoreUUID(m, movement)
		// восстановим подчиненные структуры
		if err = movement.restoreStruct(ctx, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "Mutation").Err(err).Msg("Error restore struct person")
			return nil, nil, gqlerror.Errorf("Error restore struct person")
		}
		updateOrDelete = true
	} else {
		// иначе создадим с нуля объект
		newUUID := uuid.New()
		m.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	// дополним пропущенные поля, если они есть
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, m, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем объект
	setColumns = utils.ClearSQLFields(m, setColumns)
	if len(setColumns) > 0 {
		if updateOrDelete {
			// Обновляем иначе
			rows, err := pglxqb.
				Update("movements").
				SetMap(setColumns).
				Where("uuid = ?", m.UUID).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).
				QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error update contact")
				return nil, nil, gqlerror.Errorf("Error update contact")
			}
			return rows, m.UUID, nil
		} else {
			rows, err := pglxqb.
				Insert("movements").
				SetMap(setColumns).
				Suffix(utils.PrepareSuffix(rColumns)).
				RunWith(db).
				QueryX(ctx)
			if err != nil {
				app.Logger.Error().Str("module", "contact").Str("func", "Mutation").Err(err).Msg("Error insert contact")
				return nil, nil, gqlerror.Errorf("Error insert contact")
			}
			return rows, m.UUID, nil
		}
	}
	return nil, m.UUID, nil
}

func (m *Movement) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Movement, error) {
	var movements []*Movement
	defer rows.Close()
	for rows.Next() {
		var movement Movement
		err := rows.StructScan(&movement)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		movements = append(movements, &movement)
	}
	for _, movement := range movements {
		if err := movement.parseRequestedFields(ctx, fields, app, db); err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	return movements, nil
}

func (m *Movement) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Movement, error) {
	var err error
	var movement Movement
	for rows.Next() {
		err = rows.StructScan(&movement)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
	}
	err = movement.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &movement, nil
}

func (m *Movement) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, m)
}

func (m *Movement) restoreStruct(ctx context.Context, app *app.App, db pglxqb.BaseRunner) error {
	v := reflect.ValueOf(m)
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

func (m *Movement) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Movement, error) {
	rows, err := pglxqb.SelectAll().From("movements").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var movement Movement
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&movement); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &movement, nil
}

func (m *Movement) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Movement, error) {
	rows, err := pglxqb.Select("movements.*").From("movements").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return m.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (m *Movement) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Movement, error) {
	rows, err := pglxqb.Select("movements.*").From("movements").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return m.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}
