package user

/* Model for data profile */
type UserProfileModel struct {
	Email string `json:"email" binding:"required"`
	Data  string `json:"data" binding:"required" db:"data"`
}

/* Model for request update profile user */
type UserProfileDataModel struct {
	Email      string `json:"email" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Surname    string `json:"surname" binding:"required"`
	Patronymic string `json:"patronymic"`
	Gender     bool   `json:"gender"`
	Phone      string `json:"phone"`
	Nickname   string `json:"nickname" binding:"required"`
	DateBirth  string `json:"date_birth"`
}
