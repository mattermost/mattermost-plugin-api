package settings

//SettingStore defines the behaviour needed to set and get settings
type SettingStore interface {
	SetSetting(userID, settingID string, value interface{}) error
	GetSetting(userID, settingID string) (interface{}, error)
}
