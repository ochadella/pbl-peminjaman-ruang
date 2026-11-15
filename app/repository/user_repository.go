package repository

import (
	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByNimNip(nimNip string) (*model.User, error)
	FindAll() ([]model.User, error)
	Update(user *model.User) error
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db} }

func (r *userRepository) Create(u *model.User) error { return r.db.Create(u).Error }

func (r *userRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	return &u, err
}

func (r *userRepository) FindByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *userRepository) FindByNimNip(nimNip string) (*model.User, error) {
	var u model.User
	err := r.db.Where("nim_nip = ?", nimNip).First(&u).Error
	return &u, err
}

func (r *userRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.db.Order("id").Find(&users).Error
	return users, err
}

func (r *userRepository) Update(u *model.User) error { return r.db.Save(u).Error }