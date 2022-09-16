package rbac

type RoleModel struct {
	Id          int     `json:"id" db:"id"`
	Uuid        string  `json:"uuid" db:"uuid"`
	Value       string  `json:"value" db:"value"`
	Description string  `json:"description" db:"description"`
	UsersId     *string `json:"users_id" db:"users_id"`
	DomainsId   *int    `json:"domains_id" db:"domains_id"`
}

type UsersRolesModel struct {
	Id      int    `json:"id" db:"id"`
	UsersId string `json:"users_id" db:"users_id"`
	RolesId string `json:"roles_id" db:"roles_id"`
}
