package types

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type GetUserRequest struct {
	Login string `json:"login"`
}

type UploadFileRequest struct {
	Name string `json:"name"`
}

type FollowRequest struct {
	User     *User `json:"user"`
	Follower *User `json:"follower"`
}
