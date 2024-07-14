package repository

import sq "github.com/Masterminds/squirrel"

func getUserByID(userID int) sq.SelectBuilder {
	return sq.Select("u.id", "u.first_name", "u.last_name", "u.date_of_birth", "u.gender", "u.education", "u.address").
		From("deu.users u").
		Where(sq.Eq{"u.id": userID}).
		PlaceholderFormat(sq.Dollar)
}
