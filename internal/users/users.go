package users

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/middleware"
	"github.com/sphera-erp/sphera/internal/models"
	"github.com/sphera-erp/sphera/internal/organizations"
	"github.com/sphera-erp/sphera/internal/utils"
	"github.com/sphera-erp/sphera/pkg/pglx/pglxqb"
	"github.com/sphera-erp/sphera/pkg/smsCommunicator"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/structpb"
)

type Resolver struct {
	env *app.App
	Resolvers
}

type Resolvers interface {
	RegUserByPhone(ctx context.Context, phone *string) (*string, error)
	Validate(ctx context.Context, pincode *string) (*string, error)
	AuthUserByPhone(ctx context.Context, phone *string, client *models.ClientType) (*string, error)
	User(ctx context.Context, user models.User) (*models.User, error)
	Users(ctx context.Context, user *models.User, filter *models.UserFilter, sort []models.UserSort, offset *int, limit *int) (*[]models.User, error)
	UserMutation(ctx context.Context, user *models.User) (*models.User, error)
	GetCurrentUser(ctx context.Context) (*models.User, error)
	ResetUser(ctx context.Context, phone *string) (bool, error)
	UpdateToken(ctx context.Context, token string) (bool, error)
	UsersByObject(ctx context.Context, object *models.Organization, user *models.User, filter *models.UserFilter, sort []models.UserSort, offset *int, limit *int) ([]*models.User, error)
	UserSub(ctx context.Context) (<-chan *models.User, error)
}

func NewUsersResolvers(app *app.App) (*Resolver, error) {
	return &Resolver{
		env: app,
	}, nil
}

func (r *Resolver) ResetUser(ctx context.Context, phone *string) (bool, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UserMutation").Err(err).Msg("Error run transaction")
		return false, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	_, err = pglxqb.Update("contacts").
		Set("presentation", fmt.Sprintf("%06d", utils.NewCryptoRand())).
		Where(pglxqb.Eq{"presentation": utils.PreparePhone(*phone)}).
		RunWith(tx).Exec(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "user").Str("func", "createEmptyUser").Err(err).Msg("Error create null user in DB")
		return false, gqlerror.Errorf("Error create null user in DB")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "createEmptyUser").Err(err).Msg("Error commit transaction")
		return false, gqlerror.Errorf("Error commit transaction")
	}
	return true, nil
}

func (r *Resolver) Users(ctx context.Context, user *models.User, filter *models.UserFilter, sort []models.UserSort, offset *int, limit *int) ([]*models.User, error) {
	var err error
	table := "users"
	logger := r.env.Logger.Error().Str("package", "users").Str("func", "Users")
	sql := pglxqb.Select(fmt.Sprintf("%s.*", table)).From(table)
	if filter != nil {
		sql = utils.ReflectFilter(table, sql, filter)
	} else if user != nil {
		var result map[string]interface{}
		result, sql, err = models.SqlGenSelectKeys(user, sql, table, 1)
		if err != nil {
			logger.Err(err).Msg("Error generate select relations")
			return nil, gqlerror.Errorf("Error generate select relations")
		}
		if len(result) > 0 {
			sql = sql.Where(pglxqb.Eq(result))
		}
		fmt.Println(sql.ToSql())
	}
	if sort != nil {
		for _, sortItem := range sort {
			sql = sql.OrderBy(fmt.Sprintf("%s.%s %s", table, sortItem.Field, sortItem.Order))
		}
	}
	if limit != nil {
		sql = sql.Limit(uint64(*limit))
	}
	if offset != nil {
		sql = sql.Offset(uint64(*offset))
	}
	rows, err := sql.RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error select users")
		return nil, gqlerror.Errorf("Error select users")
	}
	return user.ParseRows(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) GetCurrentUser(ctx context.Context) (*models.User, error) {
	var err error
	logger := r.env.Logger.Error().Str("package", "users").Str("func", "GetCurrentUser")

	UUIDUser, err := middleware.ExtractUserInTokenMetadata(ctx, r.env)
	if err != nil {
		logger.Err(err).Msg("Error select users")
		return nil, gqlerror.Errorf("Error select users")
	}

	rows, err := pglxqb.Select("users.*").From("users").Where(pglxqb.Eq{"uuid": UUIDUser}).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		logger.Err(err).Msg("Error select users")
		return nil, gqlerror.Errorf("Error select users")
	}
	var user models.User
	return user.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) User(ctx context.Context, user *models.User) (*models.User, error) {
	var err error
	sql := pglxqb.Select("users.*").From("users")
	result, sql, err := models.SqlGenSelectKeys(user, sql, "users", 1)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "Users").Err(err).Msg("Error generate select relations")
		return nil, gqlerror.Errorf("Error generate select relations")
	}
	rows, err := sql.Where(pglxqb.Eq(result)).RunWith(r.env.Cockroach).QueryX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "Users").Err(err).Msg("Error select users")
		return nil, gqlerror.Errorf("Error select users")
	}
	return user.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, r.env.Cockroach)
}

func (r *Resolver) UserMutation(ctx context.Context, user *models.User) (*models.User, error) {
	tx, err := r.env.Cockroach.BeginX(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UserMutation").Err(err).Msg("Error run transaction")
		return nil, gqlerror.Errorf("Error run transaction")
	}
	defer tx.Rollback(ctx)

	columns := make(map[string]interface{})
	rows, _, err := user.Mutation(ctx, tx, r.env, nil, columns)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UserMutation").Err(err).Msg("Error mutation user")
		return nil, err
	}
	user, err = user.ParseRow(ctx, r.env, graphql.CollectFieldsCtx(ctx, nil), rows, tx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UserMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}
	err = tx.Commit(ctx)
	if err != nil {
		r.env.Logger.Error().Str("module", "users").Str("func", "UserMutation").Err(err).Msg("Error commit transaction")
		return nil, gqlerror.Errorf("Error commit transaction")
	}

	// получим еще раз пользователя
	var organization models.Organization
	oRows, err := pglxqb.Select("organizations.*").
		From("organizations").
		LeftJoin("users u on organizations.uuid = u.uuid_organization").
		Where(pglxqb.Eq{"u.uuid": user.UUID}).
		RunWith(r.env.Cockroach).QueryX(ctx)

	if err != nil {
		r.env.Logger.Error().Str("module", "flow").Str("func", "UserMutation").Err(err).Msg("Error Select Org from user ")
		return nil, gqlerror.Errorf("Error Select Org from user")
	}

	for oRows.Next() {
		if err := oRows.StructScan(&organization); err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "UserMutation").Err(err).Msg("Error parse org ")
			return nil, gqlerror.Errorf("Error parse org")
		}
	}

	for _, c := range organizations.SubscriptionsMutateOrganizationResults.MutateOrganizationResults[uuid.Nil] {
		if err := organization.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		}
		c.Chanel <- &organization
	}

	for _, c := range SubscriptionsMutateUserResults.MutateUserResults[uuid.Nil] {
		userSub := user
		if err := userSub.ParseRequestedFields(ctx, graphql.CollectFieldsCtx(c.SubContext, nil), r.env, r.env.Cockroach); err != nil {
			r.env.Logger.Error().Str("module", "medicalBooks").Str("func", "MedicalBookMutation").Err(err).Msg("Error parse row in medicalBook")
		}
		c.Chanel <- userSub
	}

	return user, nil
}

func (r *Resolver) RegUserByPhone(ctx context.Context, phone string) (*string, error) {
	if r.env.Cfg.UseTracer {
		tr := otel.Tracer("RegUserByPhone")
		_, span := tr.Start(ctx, "RegUserByPhone")
		defer span.End()
	}
	if phone == "" {
		return nil, gqlerror.Errorf("Empty phone")
	}
	// проверим что пользовател уже есть
	count, err := countIdenticalPhones(ctx, r.env, utils.PreparePhone(phone))
	if err != nil {
		return nil, err
	}
	if count > 0 {
		r.env.Logger.Debug().Str("module", "user").Str("func", "RegUserByPhone").Msg("Phone number is already registered in the system")
		return nil, gqlerror.Errorf("Phone number is already registered in the system")
	}

	token, err := middleware.CreateToken(uuid.Nil, "", r.env.Cfg.Api.AccessSecret)
	if err != nil {
		graphql.AddErrorf(ctx, "Error generate token")
		return nil, err
	}
	pinCode := fmt.Sprintf("%06d", utils.NewCryptoRand())

	JSON := Token{
		Token:    token,
		UUIDUser: uuid.Nil,
		UserType: "SelfEmployed",
		PinCode:  pinCode,
		Phone:    utils.PreparePhone(phone),
	}
	rBody, err := json.Marshal(JSON)

	result, err := utils.SendRequest(fmt.Sprintf("%s/token/auth", r.env.Cfg.TarantoolURL), "POST", rBody, nil)
	if err != nil {
		graphql.AddErrorf(ctx, "Error save token and pin code to Tarantool")
		return nil, err
	}
	var body TokenTarantool

	err = json.Unmarshal(result, &body)
	if err != nil {
		graphql.AddErrorf(ctx, "Error unmarshalling response from Tarantool")
		return nil, err
	}

	if body.Status == "success" {
		_ = smsCommunicator.SendMessage(utils.PreparePhone(phone), pinCode)
		str := *token
		return &str, nil
	} else {
		graphql.AddErrorf(ctx, "Error put token to Tarantool: %s, code: %s", body.Message, body.Code)
		return nil, err
	}
}

func (r *Resolver) AuthUserByPhone(ctx context.Context, phone string, client *models.ClientType) (*string, error) {
	if r.env.Cfg.UseTracer {
		tr := otel.Tracer("AuthUserByPhone")
		_, span := tr.Start(ctx, "AuthUserByPhone")
		defer span.End()
	}
	if phone == "" {
		return nil, gqlerror.Errorf("Empty phone")
	}
	// проверим что пользовател уже есть
	UUIDUser, userType, err := getUserByContact(ctx, r.env, utils.PreparePhone(phone))
	if err != nil {
		return nil, err
	}
	if UUIDUser == uuid.Nil {
		r.env.Logger.Debug().Str("module", "user").Str("func", "RegUserByPhone").Msg("Phone number is not registered in the system")
		return nil, gqlerror.Errorf("Phone number is not registered in the system")
	}

	if client == nil {
		r.env.Logger.Debug().Str("module", "user").Str("func", "RegUserByPhone").Msg("Unknown client, access denied")
		return nil, gqlerror.Errorf("Unknown client, access denied")
	}

	if userType == "SelfEmployed" && client.String() != models.ClientTypeMobile.String() {
		r.env.Logger.Debug().Str("module", "user").Str("func", "RegUserByPhone").Msg("Self-signed cannot log into web application and client application")
		return nil, gqlerror.Errorf("Self-signed cannot log into web application and client application")
	}

	if userType == "SystemUser" && client.String() == models.ClientTypeMobile.String() {
		r.env.Logger.Debug().Str("module", "user").Str("func", "RegUserByPhone").Msg("The system user cannot log into the self-assigned application")
		return nil, gqlerror.Errorf("The system user cannot log into the self-assigned application")
	}

	token, err := middleware.CreateToken(UUIDUser, userType, r.env.Cfg.Api.AccessSecret)
	if err != nil {
		return nil, gqlerror.Errorf("Error generate token")
	}

	pinCode := fmt.Sprintf("%06d", utils.NewCryptoRand())

	JSON := Token{
		Token:    token,
		UUIDUser: UUIDUser,
		UserType: userType,
		PinCode:  pinCode,
		Phone:    utils.PreparePhone(phone),
	}

	rBody, err := json.Marshal(JSON)
	if err != nil {
		return nil, gqlerror.Errorf("Marshaling error of body parameters for sending to Tarantool")
	}

	result, err := utils.SendRequest(fmt.Sprintf("%s/token/auth", r.env.Cfg.TarantoolURL), "POST", rBody, nil)
	if err != nil {
		return nil, gqlerror.Errorf("Error save token and pin code to Tarantool")
	}

	var body TokenTarantool

	err = json.Unmarshal(result, &body)
	if err != nil {
		return nil, gqlerror.Errorf("Error unmarshalling response from Tarantool")
	}

	if body.Status == "success" {

		var pushToken *string
		err = pglxqb.Select("notification_token").
			From("users").
			Where(pglxqb.Eq{"uuid": UUIDUser}).RunWith(r.env.Cockroach).QueryRow(ctx).Scan(&pushToken)
		if err != nil {
			r.env.Logger.Error().Str("module", "flow").Str("func", "AgreeToJob").Err(err).Msg("Error select status from jobs")
			return nil, gqlerror.Errorf("Error select status from jobs")
		}

		// pin = 123123, type= pin_code_type
		if pushToken != nil {
			r.env.SendDataPush("192.168.10.244:9999", []string{*pushToken}, "pincode:"+pinCode, &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"type": {
						Kind: &structpb.Value_StringValue{StringValue: "pin_code_type"},
					},
					"pin": {
						Kind: &structpb.Value_StringValue{StringValue: pinCode},
					},
				},
			})
		}
		_ = smsCommunicator.SendMessage(utils.PreparePhone(phone), pinCode)

		str := *token
		return &str, nil
	} else {
		return nil, gqlerror.Errorf("Error put token to Tarantool: %s, code: %s", body.Message, body.Code)
	}
}

func (r *Resolver) Validate(ctx context.Context, pincode string) (*string, error) {
	if r.env.Cfg.UseTracer {
		tr := otel.Tracer("Validate")
		_, span := tr.Start(ctx, "Validate")
		defer span.End()
	}

	var body TokenTarantool

	queue := map[string]string{
		"pin_code": pincode,
	}

	result, err := utils.SendRequest(fmt.Sprintf("%s/token/getByPinCode", r.env.Cfg.TarantoolURL), "GET", nil, queue)
	if err != nil {
		graphql.AddErrorf(ctx, "Error get user uuid from Tarantool")
		return nil, err
	}

	err = json.Unmarshal(result, &body)

	if err != nil {
		graphql.AddErrorf(ctx, "Error unmarshalling response from Tarantool")
		return nil, err
	}

	var UUIDUser = uuid.Nil
	var UserType = "SelfEmployed"
	var phone = ""
	for _, Token := range body.Tokens {
		UUIDUser = Token.UUIDUser
		UserType = Token.UserType
		phone = Token.Phone
	}

	//если нулевой uuid то создадим юзера
	if UUIDUser == uuid.Nil {
		UUIDUser, err = createEmptyUser(ctx, r.env, phone)
	}
	if err != nil {
		return nil, err
	}

	token, err := middleware.CreateToken(UUIDUser, UserType, r.env.Cfg.Api.AccessSecret)
	if err != nil {
		graphql.AddErrorf(ctx, "Error generate token")
		return nil, err
	}

	JSON := Token{
		Token:    token,
		UUIDUser: UUIDUser,
		UserType: UserType,
		Phone:    phone,
	}
	rBody, err := json.Marshal(JSON)
	result, err = utils.SendRequest(fmt.Sprintf("%s/token/session", r.env.Cfg.TarantoolURL), "POST", rBody, nil)
	if err != nil {
		graphql.AddErrorf(ctx, "Error save token and pin code to Tarantool")
		return nil, err
	}

	err = json.Unmarshal(result, &body)

	if err != nil {
		graphql.AddErrorf(ctx, "Error unmarshalling response from Tarantool")
		return nil, err
	}

	if body.Status == "success" {
		return token, nil
	} else {
		graphql.AddErrorf(ctx, "Error put token to Tarantool: %s, code: %s", body.Message, body.Code)
		return nil, err
	}
}
