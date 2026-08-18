package models

type Group struct {
	ID        string
	Name      string
	CreatedAt int64
}

type GroupResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
}

type GroupMember struct {
	GroupID  string
	UserID   string
	JoinedAt int64
}

type GroupMemberResponse struct {
	UserID   string `json:"user_id"`
	JoinedAt int64  `json:"joined_at"`
}

type CreateGroupRequest struct {
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
}

type RenameGroupRequest struct {
	Name string `json:"name"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id"`
}

type GroupProfilePicture struct {
	GroupID   string
	ImageData []byte
	Version   int64
	UpdatedAt int64
}

type GroupProfilePictureRequest struct {
	ImageData string `json:"image_data"`
	Version   int64  `json:"version"`
}

type GroupProfilePictureResponse struct {
	ImageData string `json:"image_data"`
	Version   int64  `json:"version"`
}
