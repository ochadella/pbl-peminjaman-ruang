package service

import (
	"testing"

	"github.com/ochadella/pbl-peminjaman-ruang/app/model"
	"github.com/ochadella/pbl-peminjaman-ruang/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockRepo struct{ users []*model.User }

func (m *mockRepo) Create(u *model.User) error {
	u.ID = uint(len(m.users) + 1)
	m.users = append(m.users, u)
	return nil
}
func (m *mockRepo) FindByID(id uint) (*model.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockRepo) FindByEmail(e string) (*model.User, error) {
	for _, u := range m.users {
		if u.Email == e {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockRepo) FindByNimNip(n string) (*model.User, error) {
	for _, u := range m.users {
		if u.NimNip == n {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (m *mockRepo) FindAll() ([]model.User, error) { return nil, nil }
func (m *mockRepo) Update(u *model.User) error      { return nil }

func userRoem(t *testing.T, aktif bool) *model.User {
	hash, err := bcrypt.GenerateFromPassword([]byte("PasswordRoem123"), bcrypt.MinCost)
	require.NoError(t, err)
	return &model.User{ID: 1, NimNip: "2301010001", Name: "Roem Hanan",
		Email: "roem.hanan@student.ac.id", PasswordHash: string(hash),
		Role: model.RoleMahasiswa, IsActive: aktif}
}

func TestRegister_Berhasil(t *testing.T) {
	repo := &mockRepo{}
	u, err := NewAuthService(repo).Register(RegisterRequest{
		NimNip: "2301010002", Name: "Kevin Rio",
		Email: "kevin.rio@student.ac.id", Password: "PasswordKevin456"})

	require.NoError(t, err)
	assert.Equal(t, "Kevin Rio", u.Name)
	assert.Equal(t, model.RoleMahasiswa, u.Role)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("PasswordKevin456")))
	assert.Len(t, repo.users, 1)
}

func TestRegister_EmailAtauNimSudahAda(t *testing.T) {
	svc := NewAuthService(&mockRepo{users: []*model.User{userRoem(t, true)}})

	_, err := svc.Register(RegisterRequest{NimNip: "2301010099", Name: "Asilla Cesa",
		Email: "roem.hanan@student.ac.id", Password: "PasswordAsilla789"})
	assert.ErrorIs(t, err, ErrEmailTaken)

	_, err = svc.Register(RegisterRequest{NimNip: "2301010001", Name: "Asilla Cesa",
		Email: "asilla.cesa@student.ac.id", Password: "PasswordAsilla789"})
	assert.ErrorIs(t, err, ErrNimNipTaken)
}

func TestLogin_Berhasil_TokenBerisiIdentitas(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret-untuk-testing-minimal-32-karakter")
	svc := NewAuthService(&mockRepo{users: []*model.User{userRoem(t, true)}})

	token, _, err := svc.Login(LoginRequest{Email: "roem.hanan@student.ac.id", Password: "PasswordRoem123"})

	require.NoError(t, err)
	claims, err := helper.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, model.RoleMahasiswa, claims.Role)
}

func TestLogin_Gagal(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret-untuk-testing-minimal-32-karakter")

	aktif := NewAuthService(&mockRepo{users: []*model.User{userRoem(t, true)}})
	_, _, err := aktif.Login(LoginRequest{Email: "roem.hanan@student.ac.id", Password: "salah"})
	assert.ErrorIs(t, err, ErrInvalidCredentials, "password salah")

	_, _, err = aktif.Login(LoginRequest{Email: "kevin.rio@student.ac.id", Password: "PasswordRoem123"})
	assert.ErrorIs(t, err, ErrInvalidCredentials, "email tidak terdaftar")

	nonaktif := NewAuthService(&mockRepo{users: []*model.User{userRoem(t, false)}})
	_, _, err = nonaktif.Login(LoginRequest{Email: "roem.hanan@student.ac.id", Password: "PasswordRoem123"})
	assert.ErrorIs(t, err, ErrAccountInactive, "akun dinonaktifkan")
}