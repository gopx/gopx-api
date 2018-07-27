package types

// User represents a single user.
type User struct {
	Name     string             `json:"name"`
	Email    string             `json:"email"`
	Username string             `json:"username"`
	Avatar   string             `json:"avator"`
	Company  string             `json:"company"`
	Social   UserSocialAccounts `json:"social"`
}

// UserSocialAccounts represents the collection of social accounts of the user.
type UserSocialAccounts struct {
	Github        string `json:"github"`
	Twitter       string `json:"twitter"`
	StackOverflow string `json:"stackOverflow"`
	Linkedin      string `json:"linkedin"`
}
