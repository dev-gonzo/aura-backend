package config

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	autoresrepository "sistema-editorial/editora/backend/src/autores/repository"
	autoresroutes "sistema-editorial/editora/backend/src/autores/routes"
	autoresservice "sistema-editorial/editora/backend/src/autores/service"
	editaisrepository "sistema-editorial/editora/backend/src/editais/repository"
	editaisroutes "sistema-editorial/editora/backend/src/editais/routes"
	editalservice "sistema-editorial/editora/backend/src/editais/service"

	authrepository "sistema-editorial/editora/backend/src/auth/repository"
	authroutes "sistema-editorial/editora/backend/src/auth/routes"
	authservice "sistema-editorial/editora/backend/src/auth/service"
	healthrepository "sistema-editorial/editora/backend/src/health/repository"
	healthroutes "sistema-editorial/editora/backend/src/health/routes"
	healthservice "sistema-editorial/editora/backend/src/health/service"
	livrosrepository "sistema-editorial/editora/backend/src/livros/repository"
	livrosroutes "sistema-editorial/editora/backend/src/livros/routes"
	livrosservice "sistema-editorial/editora/backend/src/livros/service"
	logisticarepository "sistema-editorial/editora/backend/src/logistica/repository"
	logisticaroutes "sistema-editorial/editora/backend/src/logistica/routes"
	logisticaservice "sistema-editorial/editora/backend/src/logistica/service"
	lojarepository "sistema-editorial/editora/backend/src/loja/repository"
	lojaroutes "sistema-editorial/editora/backend/src/loja/routes"
	lojaservice "sistema-editorial/editora/backend/src/loja/service"
	pagamentosrepository "sistema-editorial/editora/backend/src/pagamentos/repository"
	pagamentosroutes "sistema-editorial/editora/backend/src/pagamentos/routes"
	pagamentosservice "sistema-editorial/editora/backend/src/pagamentos/service"
	pedidosrepository "sistema-editorial/editora/backend/src/pedidos/repository"
	pedidosroutes "sistema-editorial/editora/backend/src/pedidos/routes"
	pedidosservice "sistema-editorial/editora/backend/src/pedidos/service"
	usuariosrepository "sistema-editorial/editora/backend/src/usuarios/repository"
	usuariosroutes "sistema-editorial/editora/backend/src/usuarios/routes"
	usuariosservice "sistema-editorial/editora/backend/src/usuarios/service"
)

func NewHTTPServer(cfg AppConfig, pool *pgxpool.Pool) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:4200,http://127.0.0.1:4200,http://localhost:4201,http://127.0.0.1:4201,http://localhost:4202,http://127.0.0.1:4202",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	var editalStorage ObjectStorage
	if cfg.HasSupabaseStorageConfig() {
		storage, err := NewSupabaseS3Storage(context.Background(), cfg)
		if err != nil {
			log.Printf("nao foi possivel inicializar storage do Supabase: %v", err)
		} else {
			editalStorage = storage
		}
	}

	usuariosRepository := usuariosrepository.NewPostgresRepository(pool)
	usuariosDomainService := usuariosservice.NewService(
		usuariosservice.InitialAdminConfig{
			Email:    cfg.InitialAdminEmail,
			CPF:      cfg.InitialAdminCPF,
			Password: cfg.InitialAdminPassword,
		},
		usuariosRepository,
	)
	bootstrapCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := usuariosDomainService.EnsureInitialAdmin(bootstrapCtx); err != nil {
		log.Printf("nao foi possivel garantir admin inicial: %v", err)
	}

	authRepository := authrepository.NewPostgresRepository(pool)
	authDomainService := authservice.NewService(cfg.JWTSecret, authRepository)
	requireAuth := authroutes.RequireAuth(authDomainService)

	healthRepository := healthrepository.NewDatabaseStatusRepository(pool)
	healthDomainService := healthservice.NewService(cfg.AppName, healthRepository)
	healthroutes.Register(app, healthDomainService)

	autoresRepository := autoresrepository.NewPostgresRepository(pool)
	autoresDomainService := autoresservice.NewService(autoresRepository)
	editaisRepository := editaisrepository.NewPostgresRepository(pool)
	livrosRepository := livrosrepository.NewPostgresRepository(pool)
	livrosDomainService := livrosservice.NewService(livrosRepository)
	lojaRepository := lojarepository.NewRepository(pool)
	lojaDomainService := lojaservice.NewService(lojaRepository)
	logisticaCatalogRepository := logisticarepository.NewCatalogRepository(pool)
	logisticaSettingsRepository := logisticarepository.NewSettingsRepository(pool)
	logisticaDomainService := logisticaservice.NewService(
		logisticaSettingsRepository,
		logisticaCatalogRepository,
		logisticaservice.NewSuperFreteProvider(),
		logisticaservice.NewMelhorEnvioProvider(),
		logisticaservice.NewLoggiProvider(),
		logisticaservice.NewFrenetProvider(),
	)
	pagamentosSettingsRepository := pagamentosrepository.NewSettingsRepository(pool)
	pagamentosDomainService := pagamentosservice.NewService(
		pagamentosSettingsRepository,
		pagamentosservice.NewMercadoPagoProvider(),
	)
	pedidosRepository := pedidosrepository.NewPostgresRepository(pool)
	pedidosDomainService := pedidosservice.NewService(pedidosRepository)
	editaisDomainService := editalservice.NewUploadService(nil, editaisRepository)
	if editalStorage != nil {
		editaisDomainService = editalservice.NewUploadService(editalStorageAdapter{storage: editalStorage}, editaisRepository)
	}
	autoresroutes.Register(app, autoresDomainService, requireAuth, authroutes.RequireRoles("ADMIN"))
	lojaroutes.Register(app, lojaDomainService, requireAuth, authroutes.RequireRoles("ADMIN"))
	logisticaroutes.Register(app, logisticaDomainService, requireAuth, authroutes.RequireRoles("ADMIN"))
	pagamentosroutes.Register(app, pagamentosDomainService, requireAuth, authroutes.RequireRoles("ADMIN"))
	livrosroutes.Register(app, livrosDomainService, requireAuth, authroutes.RequireRoles("ADMIN"))
	pedidosroutes.Register(app, pedidosDomainService, requireAuth, authroutes.RequireRoles("ADMIN"))

	authroutes.Register(app, authDomainService, requireAuth)
	editaisroutes.Register(app, editaisDomainService, requireAuth, authroutes.RequireRoles("ADMIN"))
	usuariosroutes.Register(
		app,
		usuariosDomainService,
		requireAuth,
		authroutes.RequireRoles("ADMIN"),
		authroutes.RequireRolesOrSelfParam("id", "ADMIN"),
	)

	return app
}

type editalStorageAdapter struct {
	storage ObjectStorage
}

func (adapter editalStorageAdapter) Upload(
	ctx context.Context,
	key string,
	body io.Reader,
	contentType string,
	cacheControl string,
	contentLength int64,
) (editalservice.StoredObject, error) {
	object, err := adapter.storage.Upload(ctx, StorageUploadInput{
		Key:           key,
		Body:          body,
		ContentType:   contentType,
		CacheControl:  cacheControl,
		ContentLength: contentLength,
	})
	if err != nil {
		return editalservice.StoredObject{}, err
	}

	return editalservice.StoredObject{
		Bucket: object.Bucket,
		Key:    object.Key,
		URL:    object.URL,
	}, nil
}

func (adapter editalStorageAdapter) Delete(ctx context.Context, key string) error {
	return adapter.storage.Delete(ctx, key)
}
