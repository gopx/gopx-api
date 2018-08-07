package types

import (
	"time"
)

// User represents a single user.
type User struct {
	Username      string             `json:"username"`
	ID            uint64             `json:"id"`
	Name          string             `json:"name"`
	Email         string             `json:"email"`
	JoinedAt      time.Time          `json:"joinedAt"`
	AvatarURL     string             `json:"avatarURL"`
	Blog          string             `json:"blog"`
	Organization  string             `json:"organization"`
	Location      string             `json:"location"`
	PackagesCount uint64             `json:"packagesCount"`
	Social        UserSocialAccounts `json:"social"`
}

// UserSocialAccounts represents the collection of social accounts of the user.
type UserSocialAccounts struct {
	Github        string `json:"github"`
	Twitter       string `json:"twitter"`
	StackOverflow string `json:"stackOverflow"`
	LinkedIn      string `json:"linkedin"`
}

// UserMutation holds user mutation data.
type UserMutation struct {
	Name         *string                     `json:"name"`
	Blog         *string                     `json:"blog"`
	Organization *string                     `json:"organization"`
	Location     *string                     `json:"location"`
	Social       *UserSocialAccountsMutation `json:"social"`
}

// UserSocialAccountsMutation holds user social account mutation data.
type UserSocialAccountsMutation struct {
	Github        *string `json:"github"`
	Twitter       *string `json:"twitter"`
	StackOverflow *string `json:"stackOverflow"`
	LinkedIn      *string `json:"linkedin"`
}
