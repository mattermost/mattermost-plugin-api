package flow

type FlowStore interface {
	SetProperty(userID, propertyName string, value interface{}) error
	SetPostID(userID, propertyName, postID string) error
	GetPostID(userID, propertyName string) (string, error)
	RemovePostID(userID, propertyName string) error
	GetCurrentStep(userID string) (int, error)
	SetCurrentStep(userID string, step int) error
	DeleteCurrentStep(userID string) error
}
