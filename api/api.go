package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	// "github.com/jordan-wright/unindexed"
	"github.com/sphera-erp/sphera/app"
	"github.com/sphera-erp/sphera/internal/directives"
	"github.com/sphera-erp/sphera/internal/middleware"
	// "github.com/sphera-erp/sphera/pkg/otelgqlgen"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/sphera-erp/sphera/internal"
	"github.com/sphera-erp/sphera/internal/resolvers"
)

var userCtxKey = "userKey"

var mb int64 = 1 << 20

func Api(ctx context.Context, cancel context.CancelFunc, app *app.App) {

	cfg := app.Cfg

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	resolvers, err := resolvers.New(app)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()

	// ui := http.FileServer(unindexed.Dir("./ui/dist/"))
	// router.Handle("/", ui)

	//ui := http.FileServer(unindexed.Dir("./ui/dist/"))
	//router.Handle("/", ui)

	router.Handle("/storage/{bucket}/{object}", getObjectHandler(app))

	gqlCfg := internal.Config{Resolvers: resolvers}
	gqlCfg.Directives.Private = directives.NewPrivate(app)
	gqlCfg.Directives.HasAccess = directives.NewHasAccess(app)
	gqlCfg.Directives.BlockParsePerson = directives.NewBlockParsePerson(app)

	schema := internal.NewExecutableSchema(gqlCfg)

	graphQLHandler := newGqlServer(schema, cfg.Api.AllowedOrigins, app)
	// graphQLHandler.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	// 	oc := graphql.GetOperationContext(ctx)
	// 	fmt.Printf("around: %s %s", oc.OperationName, oc.RawQuery)
	// 	fmt.Println("variables: ", oc.Variables)
	// 	return next(ctx)
	// })

	router.Use(middleware.Middleware(app))
	router.Use(UserMiddleware)
	router.Handle("/gql", graphQLHandler)  // date, userUuid, req uuid, reqbody
	router.Handle("/wgql", graphQLHandler) // date, userUuid, req uuid, reqbody
	router.Handle("/playground", playground.Handler("API", "/gql"))
	router.Handle("/voyager", VoyagerHandler())
	router.Handle("/qr/{code}", QrcodeHandler())
	router.Handle("/static/{icon}", IconHandler())
	router.Handle("/succeeded", succeededHandler())
	router.Handle("/failed", failedHandler())
	router.Handle("/check", checkHandler())
	router.Handle("/pay", payHandler())

	serverHandler := cors.New(cors.Options{
		AllowedOrigins: cfg.Api.AllowedOrigins,
		AllowedMethods: []string{
			"POST", "GET", "OPTIONS",
		},
		AllowCredentials:   true,
		AllowedHeaders:     []string{"*"},
		OptionsPassthrough: false,
	}).Handler(router)

	//loggedRouter := LoggingHandler(app.Logger, serverHandler)

	//addr := fmt.Sprintf("%s:%d", cfg.Api.Address, cfg.Api.Port)
	addr := fmt.Sprintf(":%d", cfg.Api.Port)
	//srv := &http.Server{Addr: addr, Handler: loggedRouter}
	srv := &http.Server{Addr: addr, Handler: serverHandler}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			app.Logger.Err(err).Str("module", "api").Str("func", "api").Msgf("HTTP server start error", err)
		}
	}()

	for {
		oscall := <-c
		switch oscall {
		case os.Interrupt:
			if err := srv.Shutdown(ctx); err != nil {
				app.Logger.Fatal().Str("module", "api").Str("func", "api").Msgf("Server Shutdown Failed:%+v", err)
			}
			app.Logger.Info().Str("module", "api").Str("func", "api").Msgf("Server Exited Properly")
			cancel()
			return
		}
	}
}

func UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("wfmt", "123")
		next.ServeHTTP(w, r)
	})
}

func newGqlServer(es graphql.ExecutableSchema, allowedOrigins []string, app *app.App) *handler.Server {
	srv := handler.New(es)

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin:     checkFn(allowedOrigins),
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		//инициализация websocket, в случае ошибки сокет не соединиться
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			token, ok := initPayload["authorization"]
			if ok {
				tokenString := strings.TrimPrefix(token.(string), "Bearer ")
				fmt.Println(tokenString)
				return middleware.PackTokenToCtx(ctx, tokenString), nil
			}
			token, ok = initPayload["Authorization"]
			if ok {
				tokenString := strings.TrimPrefix(token.(string), "Bearer ")
				fmt.Println(tokenString)
				return middleware.PackTokenToCtx(ctx, tokenString), nil
			}
			return ctx, nil
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{
		//MaxMemory:     32 * mb,
		//MaxUploadSize: 50 * mb,
	})

	srv.SetQueryCache(lru.New(1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New(100),
	})
	// srv.Use(otelgqlgen.NewTracer())

	return srv
}

func checkFn(allowedOrigins []string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		if r == nil {
			return false
		}

		requestOrigin := r.Header.Get("Origin")
		if requestOrigin == "" {
			return true
		}

		for _, allowed := range allowedOrigins {
			if match(requestOrigin, allowed) {
				return true
			}
		}

		return false
	}
}

func match(s, pattern string) bool {
	left, right := split(pattern)
	return strings.HasPrefix(s, left) && strings.HasSuffix(s, right)
}

func split(pattern string) (string, string) {
	spliced := strings.SplitN(pattern, "*", 2)

	if len(spliced) == 2 {
		return spliced[0], spliced[1]
	}

	if strings.HasPrefix(pattern, "*") {
		return "", spliced[0]
	}

	return spliced[0], ""
}
