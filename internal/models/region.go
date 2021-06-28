package models

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/ekomobile/dadata/v2/api/model"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Region struct {
	UUID        *uuid.UUID `json:"uuid" db:"uuid" auto:"false"`
	Name        *string    `json:"name" db:"region"`
	Created     *time.Time `json:"created" db:"created"`
	Updated     *time.Time `json:"updated" db:"updated"`
	UUIDCountry *uuid.UUID `db:"uuid_country"`
	Country     *Country   `json:"country" relay:"uuid_country" link:"UUIDCountry"`
	IsDeleted   *bool      `json:"isDeleted" db:"is_deleted"`
}

type RegionFilter struct {
	UUID      *uuid.UUID `json:"uuid" db:"uuid"`
	Name      *string    `json:"name" db:"name"`
	Created   *time.Time `json:"created" db:"created"`
	Updated   *time.Time `json:"updated" db:"updated"`
	IsDeleted *bool      `json:"isDeleted" db:"is_deleted"`
}

func (r *Region) Mutation(ctx context.Context, db pglxqb.BaseRunner, app *app.App, rColumns interface{}, columns map[string]interface{}) (*pglx.Rows, *uuid.UUID, error) {
	//updateOrDelete := false
	// если есть uuid значит манипулируем обектом
	fmt.Println(r)
	if r.UUID != nil {
		if utils.CountFillFields(r) == 1 && len(columns) == 0 {
			return nil, r.UUID, nil
		}
	} else {
		// иначе создадим с нуля Обьект
		newUUID := uuid.New()
		r.UUID = &newUUID
		columns["uuid"] = newUUID
	}
	parent := make(map[string]interface{})
	setColumns, err := SqlGenKeys(ctx, app, db, r, columns, parent)
	if err != nil {
		app.Logger.Error().Str("module", "persons").Str("func", "Mutation").Err(err).Msg("Error generate map of keys")
		return nil, nil, err
	}
	// только одна колонка, и это uuid то удаляем обьект
	setColumns = utils.ClearSQLFields(r, setColumns)
	if len(columns) > 0 {
		rows, err := pglxqb.Insert("regions").
			SetMap(columns).
			OnConflictUpdateMap(setColumns, "uuid").
			Upset("updated", "now()").
			Suffix(utils.PrepareSuffix(rColumns)).
			RunWith(db).QueryX(ctx)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "Mutation").Err(err).Msg("Error insert region")
			return nil, nil, gqlerror.Errorf("Error insert region")
		}
		return rows, r.UUID, nil
	}
	return nil, r.UUID, nil
}

func (r *Region) ParseRows(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) ([]*Region, error) {
	var regions []*Region
	defer rows.Close()
	for rows.Next() {
		var region Region
		err := rows.StructScan(&region)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "ParseRows").Err(err).Msg("Error scan response to struct region")
			return nil, gqlerror.Errorf("Error scan response to struct region")
		}
		err = region.parseRequestedFields(ctx, fields, app, db)
		if err != nil {
			app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
			return nil, gqlerror.Errorf("Error scan response to struct user")
		}
		regions = append(regions, &region)
	}
	return regions, nil
}

func (r *Region) ParseRow(ctx context.Context, app *app.App, fields []graphql.CollectedField, rows *pglx.Rows, db pglxqb.BaseRunner) (*Region, error) {
	var err error
	var region Region
	defer rows.Close()
	for rows.Next() {
		err = rows.StructScan(&region)
		if err != nil {
			app.Logger.Error().Str("module", "models").Str("func", "manipulate").Err(err).Msg("Error scan response to struct region")
			return nil, gqlerror.Errorf("Error scan response to struct region")
		}
	}
	err = region.parseRequestedFields(ctx, fields, app, db)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return &region, nil
}

func (r *Region) parseRequestedFields(ctx context.Context, fields []graphql.CollectedField, app *app.App, db pglxqb.BaseRunner) error {
	return parseRequestedFields(ctx, app, db, fields, r)
}

func (r *Region) GetByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID) (*Region, error) {
	rows, err := pglxqb.SelectAll().From("regions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error get person from DB")
		return nil, gqlerror.Errorf("Error get person from DB")
	}
	var region Region
	for rows.Next() {
		if err := rows.StructScan(&region); err != nil {
			app.Logger.Error().Str("module", "persons").Str("func", "ParseRow").Err(err).Msg("Error scan response to struct Person")
			return nil, gqlerror.Errorf("Error scan response to struct Person")
		}
	}
	return &region, nil
}

func (r *Region) GetParsedObjectByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid *uuid.UUID, column graphql.CollectedField) (*Region, error) {
	rows, err := pglxqb.SelectAll().From("regions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return r.ParseRow(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func (r *Region) GetParsedObjectsByUUID(ctx context.Context, app *app.App, db pglxqb.BaseRunner, uuid []*uuid.UUID, column graphql.CollectedField) ([]*Region, error) {
	rows, err := pglxqb.SelectAll().From("regions").Where(pglxqb.Eq{"uuid": uuid}).RunWith(db).QueryX(ctx)
	if err != nil {
		app.Logger.Error().Str("module", "users").Str("func", "parseRequestedFields").Err(err).Msg("Error scan response to struct user")
		return nil, gqlerror.Errorf("Error scan response to struct user")
	}
	return r.ParseRows(ctx, app, graphql.CollectFields(graphql.GetOperationContext(ctx), column.Selections, nil), rows, db)
}

func parserRegion(suggestion *model.Address, address *Address, app *app.App, ctx context.Context) (err error) {
	region := new(Region)
	regionUUID := uuid.New()
	if suggestion.RegionFiasID != "" {
		regionUUID, err = uuid.Parse(suggestion.RegionFiasID)
		if err != nil {
			app.Logger.Err(err).Msg("Error parse uuid region")
			return gqlerror.Errorf("Error parse uuid region")
		}
	}
	var regionDB Region
	rows, err := pglxqb.SelectAll().
		From("regions").
		Where(pglxqb.Eq{"region": suggestion.Region}).
		RunWith(app.Cockroach).
		QueryX(ctx)
	if err != nil {
		app.Logger.Err(err).Msg("Error get region from db")
		return gqlerror.Errorf("Error get region from db")
	}
	for rows.Next() {
		if err := rows.StructScan(&regionDB); err != nil {
			app.Logger.Err(err).Msg("Error scan response to struct Region")
			return gqlerror.Errorf("Error scan response to struct Region")
		}
	}
	// у нас уже записан такой регион
	if regionDB.UUID != nil {
		suggestion.RegionFiasID = regionDB.UUID.String()
		region.UUID = regionDB.UUID
		address.Region = region
		return nil
	}
	region.UUID = &regionUUID
	region.Name = &suggestion.Region
	address.Region = region
	return nil
}
