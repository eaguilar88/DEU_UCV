package users

import "github.com/eaguilar88/deu/internal/entities"

type GetUserResponse struct {
	ID             string `json:"id,omitempty"`
	CI             string `json:"ci,omitempty"`
	Username       string `json:"nombre_usuario,omitempty"`
	FirstName      string `json:"primer_nombre,omitempty"`
	LastName       string `json:"apellido,omitempty"`
	DateOfBirth    string `json:"fecha_nacimiento,omitempty"`
	Age            int    `json:"edad,omitempty"`
	Gender         string `json:"genero,omitempty"`
	EducationLevel string `json:"nivel_educacion,omitempty"`
	Code           string `json:"codigo,omitempty"`
	Address        string `json:"direccion,omitempty"`
	CreatedAt      string `json:"creado_en,omitempty"`
}

type GetUsersResponse struct {
	Users []GetUserResponse  `json:"usuarios"`
	Pages entities.PageScope `json:"paginas"`
}

type CreateUsersResponse struct {
	ID string `json:"id,"`
}
