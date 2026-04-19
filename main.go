package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cryptocore/internal/infrastructure/fs"
	"cryptocore/internal/infrastructure/memory"
	"cryptocore/internal/infrastructure/postgres"
	"cryptocore/internal/repository"
	"cryptocore/internal/service"
	"cryptocore/internal/transport"
)

func main() {
	var (
		userRepo      repository.UserRepository
		fileRepo      repository.FileRepository
		containerRepo repository.ContainerStorage
		sessionRepo   repository.SessionRepository
		challengeRepo repository.ChallengeRepository
	)

	dsn := os.Getenv("DATABASE_URL")
	if dsn != "" {
		pool, err := postgres.NewPool(context.Background(), dsn)
		if err != nil {
			log.Fatalf("connect to postgres: %v", err)
		}
		defer pool.Close()

		if err := postgres.RunMigrations(pool); err != nil {
			log.Fatalf("run migrations: %v", err)
		}

		dataDir := os.Getenv("DATA_DIR")
		if dataDir == "" {
			dataDir = "./data"
		}

		store, err := fs.NewContainerStorage(dataDir)
		if err != nil {
			log.Fatalf("init container storage: %v", err)
		}

		userRepo = postgres.NewUserRepository(pool)
		fileRepo = postgres.NewFileRepository(pool)
		containerRepo = store
		sessionRepo = postgres.NewSessionRepository(pool)
		challengeRepo = postgres.NewChallengeRepository(pool)

		log.Println("using postgres + filesystem storage")
	} else {
		userRepo = memory.NewUserRepository()
		fileRepo = memory.NewFileRepository()
		containerRepo = memory.NewContainerStorage()
		sessionRepo = memory.NewSessionRepository()
		challengeRepo = memory.NewChallengeRepository()

		log.Println("using in-memory storage (no DATABASE_URL set)")
	}

	userService := service.NewUserService(userRepo)
	fileService := service.NewFileService(userRepo, fileRepo, containerRepo)
	authService := service.NewAuthService(userRepo, sessionRepo, challengeRepo)

	router := transport.NewRouter(userService, fileService, authService)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	log.Printf("server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
