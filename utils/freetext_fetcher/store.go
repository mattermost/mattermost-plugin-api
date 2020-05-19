package freetext_fetcher

type FreetextStore interface {
	StartFetching(userID, fetcherID string, payload string) error
	StopFetching(userID, fetcherID string) error
	ShouldProcessFreetext(userID, fetcherID string) (bool, string, error)
}
