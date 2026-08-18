package models

type ProfilePicture struct {
	UserID    string
	ImageData []byte
	Version   int64
	UpdatedAt int64
}

type ProfilePictureRequest struct {
	ImageData string `json:"image_data"`
	Version   int64  `json:"version"`
}

type ProfilePictureResponse struct {
	ImageData string `json:"image_data"`
	Version   int64  `json:"version"`
}
