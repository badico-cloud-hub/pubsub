package utils

import "github.com/badico-cloud-hub/pubsub/dto"

//FilterEvents is return filters if existed in service
func FilterEvents(events []string) []string {
	newEvents := []string{}
	allEvents := dto.AllEvents
	for _, ev := range events {
		for _, a := range allEvents {
			if ev == a {
				newEvents = append(newEvents, ev)
			}
		}
	}
	return newEvents
}

func VerifyIfUrlIsDuplicated(elems []dto.SubscriptionDTO, url, event string) bool {
	for _, s := range elems {
		if url == s.SubscriptionUrl && event == s.SubscriptionEvent {
			return true
		}
	}
	return false
}
