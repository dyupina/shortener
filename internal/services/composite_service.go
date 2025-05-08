package services

import (
	"shortener/internal/storage"
)

type CompositeService struct {
	URLService     URLService
	UserService    UserService
	StorageService storage.StorageService
}

func NewCompositeService(urlService URLService, userService UserService, storageService storage.StorageService) *CompositeService {
	return &CompositeService{
		URLService:     urlService,
		UserService:    userService,
		StorageService: storageService,
	}
}
