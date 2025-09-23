package main

// This file is conceptually separate but for simplicity, its content (initDB)
// is included in main.go for this project. In a larger application,
// you would have functions here like:
//
// func GetUserByEmail(email string) (*User, error) { ... }
// func CreateUser(email, password string) (int, error) { ... }
// func UpdateUserAvatar(userID int, avatarURL string) error { ... }
//
// This helps abstract database logic away from the HTTP handlers.

func getUserById(id int) (*User, error) {
	var user User
	row := db.QueryRow("SELECT id, prenom, nom, telephone, email, avatar_url FROM users WHERE id = $1", id)
	if err := row.Scan(&user.ID, &user.Prenom, &user.Nom, &user.Telephone, &user.Email, &user.AvatarURL); err != nil {
		return nil, err
	}

	return &user, nil
}
