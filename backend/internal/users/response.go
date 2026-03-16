package users

import "github.com/eaguilar88/deu/internal/entities"

type GetUserResponse struct {
	ID             string `json:"id,omitempty"`
	CI             string `json:"cedula,omitempty"`
	Email          string `json:"email,omitempty"`
	FirstName      string `json:"nombres,omitempty"`
	LastName       string `json:"apellidos,omitempty"`
	Bio            string `json:"biografia,omitempty"`
	DateOfBirth    string `json:"fecha_de_nacimiento,omitempty"`
	Age            int    `json:"edad,omitempty"`
	Gender         string `json:"genero,omitempty"`
	EducationLevel string `json:"nivel_educativo,omitempty"`
	Code           string `json:"codigo_proveedor,omitempty"`
	Address        string `json:"direccion,omitempty"`
	CreatedAt      string `json:"creado_en,omitempty"`
	//TODO agregar el rol
}

type GetUsersResponse struct {
	Users []GetUserResponse  `json:"usuarios"`
	Pages entities.PageScope `json:"paginas"`
}

type CreateUsersResponse struct {
	ID string `json:"id"`
}
