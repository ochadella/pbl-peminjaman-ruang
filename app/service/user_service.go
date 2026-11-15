package service

import (
	"errors"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
)

var ErrInvalidRole = errors.New("role tidak valid")

type UserService interface {
	GetProfile(id uint) (*model.User, error)
	UpdateProfile(id uint, name string) (*model.User, error)
	List() ([]model.User, error)
	UpdateRole(id uint, role string) (*model.User, error)
	SetActive(id uint, active bool) (*model.User, error)
}

type userService struct{ repo repository.UserRepository }

func NewUserService(repo repository.UserRepository) UserService { return &userService{repo} }

func (s *userService) GetProfile(id uint) (*model.User, error) { return s.repo.FindByID(id) }

func (s *userService) UpdateProfile(id uint, name string) (*model.User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	u.Name = name
	return u, s.repo.Update(u)
}

func (s *userService) List() ([]model.User, error) { return s.repo.FindAll() }

func (s *userService) UpdateRole(id uint, role string) (*model.User, error) {
	if role != model.RoleMahasiswa && role != model.RolePetugas && role != model.RoleAdmin {
		return nil, ErrInvalidRole
	}
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	u.Role = role
	return u, s.repo.Update(u)
}

func (s *userService) SetActive(id uint, active bool) (*model.User, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	u.IsActive = active
	return u, s.repo.Update(u)
}