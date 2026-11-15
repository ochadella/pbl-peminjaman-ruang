package service

import (
	"errors"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/app/repository"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrEmailTaken         = errors.New("email sudah terdaftar")
	ErrNimNipTaken        = errors.New("NIM sudah terdaftar")
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrAccountInactive    = errors.New("akun tidak aktif")
)

type RegisterRequest struct {
	NimNip   string `json:"nim_nip" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthService interface {
	Register(req RegisterRequest) (*model.User, error)
	Login(req LoginRequest) (string, *model.User, error)
}

type authService struct{ repo repository.UserRepository }

func NewAuthService(repo repository.UserRepository) AuthService { return &authService{repo} }

func (s *authService) Register(req RegisterRequest) (*model.User, error) {
	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if _, err := s.repo.FindByNimNip(req.NimNip); err == nil {
		return nil, ErrNimNipTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		NimNip: req.NimNip, Name: req.Name, Email: req.Email,
		PasswordHash: string(hash),
		Role:         model.RoleMahasiswa,
		IsActive:     true,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(req LoginRequest) (string, *model.User, error) {
	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return "", nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return "", nil, ErrAccountInactive
	}
	token, err := helper.GenerateToken(user.ID, user.Role)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}