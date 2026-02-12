package dataload

type UserId int
type CommentId int
type AttachmentId int

type SesId string

func (sId SesId) IsEmpty() bool {
	return sId == ""
}

type Session struct {
	UserId    UserId
	SessionId SesId
}

type User struct {
	Id UserId
}

type Comment struct {
	Id     CommentId
	UserId UserId
	Text   string
}

type Attachment struct {
	Id        AttachmentId
	CommentId CommentId
	DataUrl   string // или слайс байт самого вложения
}
