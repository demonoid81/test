package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sphera-erp/sphera/app"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"strings"
	"time"
)

var tokenCtxKey = &contextKey{"userToken"}

type contextKey struct {
	token string
}

func PackTokenToCtx(ctx context.Context, tokenString string) context.Context {
	return context.WithValue(ctx, tokenCtxKey, tokenString)
}

func Middleware(app *app.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if app.Cfg.UseTracer {
				var span trace.Span
				tracer := otel.GetTracerProvider().Tracer("tracerName")
				ctx, span = tracer.Start(ctx, "auth.middleware")
				defer span.End()
			}

			defer func() {
				if err := r.Body.Close(); err != nil {
					app.Logger.Error().Err(err).Msgf("Closing body failed due to an error: %s", err)
				}
			}()
			//wsProtocolHeader := r.Header.Get("sec-websocket-protocol")
			//if wsProtocolHeader != "" {
			//	wsProtocolParts := strings.Split(wsProtocolHeader, ",")
			//	if len(wsProtocolParts) > 2 {
			//		http.Error(w, "sec-websocket-protocol malformed", http.StatusBadRequest)
			//		return
			//	}
			//	if len(wsProtocolParts) == 2 {
			//		wsProtocol, wsToken := strings.TrimSpace(wsProtocolParts[0]), strings.TrimSpace(wsProtocolParts[1])
			//		fmt.Println(wsProtocolParts)
			//		r.Header.Set("Authorization", "Bearer "+wsToken)
			//		r.Header.Set("sec-websocket-protocol", wsProtocol)
			//	}
			//}

			var tokenString string
			tokens, ok := r.Header["Authorization"]
			if ok && len(tokens) >= 1 {
				tokenString = tokens[0]
				tokenString = strings.TrimPrefix(tokenString, "Bearer ")
			}
			// Allow unauthenticated users in
			if tokenString == "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx = context.WithValue(ctx, tokenCtxKey, tokenString)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func CreateToken(userUUID uuid.UUID, userType, accessSecret string) (*string, error) {
	var err error
	atClaims := jwt.MapClaims{}
	atClaims["user"] = userUUID.String()
	atClaims["type"] = userType
	atClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(accessSecret))
	if err != nil {
		return nil, gqlerror.Errorf("Auth Error. Token generate error")
	}
	return &token, nil
}

func TokenForContext(ctx context.Context) (string, error) {
	raw := ctx.Value(tokenCtxKey)
	if raw == nil {
		return "", errors.New("Unable to find user UUID in request context")
	}
	token, ok := raw.(string)
	if !ok {
		return "", errors.New("User UUID from request context does not comply with uuid interface")
	}
	return token, nil
}

func VerifyToken(ctx context.Context, app *app.App) (*jwt.Token, error) {
	tokenString, err := TokenForContext(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Println(tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(app.Cfg.Api.AccessSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractUserInTokenMetadata(ctx context.Context, app *app.App) (uuid.UUID, error) {
	token, err := VerifyToken(ctx, app)
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return uuid.Parse(claims["user"].(string))
	}
	return uuid.Nil, err
}

func ExtractUserTypeInTokenMetadata(ctx context.Context, app *app.App) (string, error) {
	token, err := VerifyToken(ctx, app)
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return claims["type"].(string), nil
	}
	return "", err
}

//// appContext.go
//const (
//	AppContextKey = "appContext"
//)
//
//type AppContext struct {
//	Token          string
//	UserId         string
//	Cancel         context.CancelFunc
//}
//
//func ForAppContext(ctx context.Context) *AppContext {
//	c, _ := ctx.Value(AppContextKey).(*AppContext)
//	return c
//}
//// init.go
//h := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver.Resolver{}}))
//// ...
//h.AddTransport(transport.Websocket{
//Upgrader: websocket.Upgrader{
//HandshakeTimeout: time.Minute,
//CheckOrigin: func(r *http.Request) bool {
//// we are already checking for CORS
//return true
//},
//EnableCompression: true,
//},
//InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
//if token := initPayload.Authorization(); middleware.CouldBetoken(token) {
//if intro, err := oauth2.IntrospectToken(token[7:], false); err == nil && intro != nil && oauth2.IsIntrospectionValid(intro) {
//nctx, cancel := context.WithCancel(ctx)
//return context.WithValue(nctx, middleware.AppContextKey, &middleware.AppContext{
//Token:          token[7:],
//UserId:         intro.Sub,
//Cancel:         cancel,
//}), nil
//}
//}
//return ctx, errors.New("AUTHORIZATION_REQUIRED")
//},
//KeepAlivePingInterval: viper.GetDuration(config.WebsocketKeepAliveKey),
//})
//// ...
//// resolvers.go
//func (r *subscriptionResolver) Notification(ctx context.Context, id string) (<-chan *model.Notification, error) {
//	appContext := middleware.ForAppContext(ctx)
//	if !oauth2.IsValid(appContext.Token) {
//		appContext.Cancel() // stop sending keep alive
//		return nil, nil
//	}
//	// ...
//}
//I would love to have the ability to close the websocket entirely instead of hoping the client disconnects. In a way like this:
//
//// resolvers.go
//func (r *subscriptionResolver) Notification(ctx context.Context, id string) (<-chan *model.Notification, error) {
//	appContext := middleware.ForAppContext(ctx)
//	if !oauth2.IsValid(appContext.Token) {
//		conn := middleware.ForWsConnection(ctx)
//		conn.Close(websocket.CloseNormalClosure, "unauthorized")
//		return nil, nil
//	}
//	// ...
//}

//func newGQLServer(allowed allowedOrigin, allowOriginFunc func(string) bool) *handler.Server {
//	srv := handler.New(generated.NewExecutableSchema(resolver.New()))
//	srv.AddTransport(transport.Websocket{
//		KeepAlivePingInterval: 10 * time.Second,
//		Upgrader: websocket.Upgrader{
//			CheckOrigin: func(r *http.Request) bool {
//				origin := r.Header["Origin"]
//				if len(origin) == 0 {
//					return true
//				}
//				u, err := url.Parse(origin[0])
//				if strings.EqualFold(u.Host, r.Host) && err == nil {
//					return true
//				}
//				return allowOriginFunc(origin[0])
//			},
//		},
//	})
//	srv.AddTransport(transport.Options{})
//	srv.AddTransport(transport.GET{})
//	srv.AddTransport(transport.POST{})
//	srv.AddTransport(transport.MultipartForm{})
//	srv.SetQueryCache(lru.New(1000))
//	srv.Use(extension.Introspection{})
//	srv.Use(extension.AutomaticPersistedQuery{
//		Cache: lru.New(100),
//	})
//	return srv
//}
