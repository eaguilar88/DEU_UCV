package repository

const getUserQuery = `SELECT ` +
	`%s ` +
	`FROM users u ` +
	`WHERE u.ud = ?`

const userCommon = `u.first_name, u.last_name, u.date_of_birth, u.gender, u.education_level, u.address`
