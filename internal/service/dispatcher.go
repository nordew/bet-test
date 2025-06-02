package service

import (
	"context"
	"log"
	"strings"

	"github.com/nordew/bet-test/internal/client"
	"github.com/nordew/bet-test/internal/model"
)

type Dispatcher interface {
	ProcessAndDispatchUsers(ctx context.Context) error
}

type dispatcher struct {
	apiClient client.APIClient
	apiBURL   string
}

func NewDispatcher(apiClient client.APIClient, apiBURL string) Dispatcher {
	return &dispatcher{
		apiClient: apiClient,
		apiBURL:   apiBURL,
	}
}

func (d *dispatcher) ProcessAndDispatchUsers(ctx context.Context) error {
	users, err := d.apiClient.FetchUsers(ctx)
	if err != nil {
		log.Printf("Fetch users error: %v", err)
		return err
	}

	log.Printf("Fetched %d users", len(users))

	for _, user := range users {
		if strings.HasSuffix(user.Email, ".biz") {
			log.Printf("User %s (.biz): sending to API B", user.Email)
			payload := model.UserPayload{
				Name:  user.Name,
				Email: user.Email,
			}

			if err := d.apiClient.SendUserToAPIB(ctx, payload, d.apiBURL); err != nil {
				log.Printf("Send to API B error for %s (%s): %v", user.Name, user.Email, err)
			} else {
				log.Printf("User %s (%s) sent to API B", user.Name, user.Email)
			}
		} else {
			log.Printf("User %s (not .biz): skipping", user.Email)
		}
	}

	log.Println("User processing finished.")
	return nil
} 