package panel

type PanelStore interface {
	SetPanelPostID(userID string, postIDs string) error
	GetPanelPostID(userID string) (string, error)
	DeletePanelPostID(userID string) error
}
