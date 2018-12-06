package user

// User Structure that stores user information retrieved from database or
// entered by user during registration
type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Avatar    string `json:"avatar"`
	Ready     bool
}
