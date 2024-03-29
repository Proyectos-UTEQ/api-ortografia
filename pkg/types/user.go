package types

import "fmt"

type UserAPI struct {
	ID                   uint   `json:"id"`
	FirstName            string `json:"first_name" validate:"required,min=6,max=100"`
	LastName             string `json:"last_name"`
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password,omitempty" validate:"required,min=6,max=100"`
	BirthDate            string `json:"birth_date"`
	PointsEarned         int    `json:"points_earned"`
	Whatsapp             string `json:"whatsapp"`
	Telegram             string `json:"telegram"`
	TelegramID           int64  `json:"telegram_id"`
	URLAvatar            string `json:"url_avatar"`
	Status               string `json:"status"`
	TypeUser             string `json:"type_user" validate:"required,oneof=student teacher admin"`
	PerfilUpdateRequired bool   `json:"perfil_update_required"`
}

func (user *UserAPI) ValidateUpdateUser() error {

	if user.FirstName == "" {
		return fmt.Errorf("el nombre es requerido")
	}

	if user.LastName == "" {
		return fmt.Errorf("el apellido es requerido")
	}

	if user.Whatsapp == "" {
		return fmt.Errorf("el whatsapp es requerido")
	}

	if user.URLAvatar == "" {
		return fmt.Errorf("la url del avatar es requerida")
	}

	return nil
}

type Login struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ResetPassword sirve para obtener el email y enviar el correo con el token.
type ResetPassword struct {
	Email string `json:"email" validate:"required,email"`
}

type ChangePassword struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}
