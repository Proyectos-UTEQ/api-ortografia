package data

import (
	"Proyectos-UTEQ/api-ortografia/internal/db"
	"Proyectos-UTEQ/api-ortografia/internal/utils"
	"Proyectos-UTEQ/api-ortografia/pkg/types"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName            string
	LastName             string
	Email                string `gorm:"uniqueIndex"`
	Password             string
	BirthDate            time.Time
	PointsEarned         int
	Whatsapp             string
	Telegram             string
	TelegramID           int64
	URLAvatar            string
	Status               Status
	TypeUser             TypeUser
	PerfilUpdateRequired bool
}

type Status string

const (
	Actived         Status = "actived"
	Blocked         Status = "blocked"
	PendingApproval Status = "pending_approval"
)

type TypeUser string

const (
	Admin   TypeUser = "admin"
	Student TypeUser = "student"
	Teacher TypeUser = "teacher"
)

func (User) TableName() string {
	return "users"
}

func UserToAPI(user User) *types.UserAPI {
	return &types.UserAPI{
		ID:                   user.ID,
		FirstName:            user.FirstName,
		LastName:             user.LastName,
		Email:                user.Email,
		Password:             "",
		BirthDate:            utils.GetDate(user.BirthDate),
		PointsEarned:         user.PointsEarned,
		Whatsapp:             user.Whatsapp,
		Telegram:             user.Telegram,
		TelegramID:           user.TelegramID,
		URLAvatar:            user.URLAvatar,
		Status:               string(user.Status),
		TypeUser:             string(user.TypeUser),
		PerfilUpdateRequired: user.PerfilUpdateRequired,
	}
}

func UsersToAPI(users []User) []types.UserAPI {
	var usersApi []types.UserAPI
	for _, user := range users {
		usersApi = append(usersApi, *UserToAPI(user))
	}
	return usersApi
}

func Login(login types.Login) (*types.UserAPI, bool, error) {
	var user User
	result := db.DB.First(&user, "email = ?", login.Email)

	// Controlar el error de record not found.
	if result.Error != nil {
		return nil, false, errors.New("las credenciales son incorrectas")
	}

	// Comparar las contraseñas con un hash.
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		return nil, false, errors.New("las credenciales son incorrectas")
	}

	// Convertir a un usuario api
	userAPI := &types.UserAPI{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Password:     "",
		BirthDate:    user.BirthDate.String(),
		PointsEarned: user.PointsEarned,
		Whatsapp:     user.Whatsapp,
		Telegram:     user.Telegram,
		URLAvatar:    user.URLAvatar,
		Status:       string(user.Status),
		TypeUser:     string(user.TypeUser),
	}

	return userAPI, true, nil
}

func Register(userAPI *types.UserAPI) (*User, error) {

	// crear un hash apartir de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userAPI.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// en caso de ser admin o profesor se pone en pendiente de aprobacion.
	status := Actived
	if userAPI.TypeUser == "admin" || userAPI.TypeUser == "teacher" {
		status = PendingApproval
	}

	// rellenamos los datos con la entidad.
	user := User{
		FirstName:            userAPI.FirstName,
		Email:                userAPI.Email,
		Password:             string(hashedPassword),
		Status:               status,
		TypeUser:             TypeUser(userAPI.TypeUser),
		PerfilUpdateRequired: true,
	}

	result := db.DB.Create(&user)

	if result.Error != nil {
		if pgerr, ok := result.Error.(*pgconn.PgError); ok {
			if pgerr.Code == "23505" {
				return nil, errors.New("el email ya existe")
			}
		}
		return nil, result.Error
	}

	userAPI.ID = user.ID

	return &user, nil
}

func ExisteEmail(email string) (bool, types.UserAPI) {
	var user User
	result := db.DB.First(&user, "email = ?", email)
	if result.Error != nil {
		return false, types.UserAPI{}
	}
	return true, types.UserAPI{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		BirthDate:    user.BirthDate.String(),
		PointsEarned: user.PointsEarned,
		Whatsapp:     user.Whatsapp,
		Telegram:     user.Telegram,
		TelegramID:   user.TelegramID,
		URLAvatar:    user.URLAvatar,
		Status:       string(user.Status),
		TypeUser:     string(user.TypeUser),
	}
}

func UpdatePassword(userid uint, newPassword string) error {
	// hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// actualizar la contraseña en la base de datos.
	result := db.DB.Model(&User{}).Where("id = ?", userid).Update("password", string(hashedPassword))
	fmt.Println("Rows affected: ", result.RowsAffected)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func SetTelegramChat(username string, chatid int64) error {

	// actualizar la contraseña en la base de datos.
	result := db.DB.Model(&User{}).Where("telegram = ?", username).Update("telegram_id", chatid)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
