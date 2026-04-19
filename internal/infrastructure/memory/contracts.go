package memory

import "cryptocore/internal/repository"

var (
	_ repository.UserRepository   = (*UserRepository)(nil)
	_ repository.FileRepository   = (*FileRepository)(nil)
	_ repository.ContainerStorage = (*ContainerStorage)(nil)
)
