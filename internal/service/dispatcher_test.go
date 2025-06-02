package service

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/nordew/bet-test/internal/model"
)

type MockAPIClient struct {
	FetchUsersFunc     func(ctx context.Context) ([]model.User, error)
	SendUserToAPIBFunc func(ctx context.Context, payload model.UserPayload, apiBURL string) error
	SentPayloads       []model.UserPayload
}

func (m *MockAPIClient) FetchUsers(ctx context.Context) ([]model.User, error) {
	if m.FetchUsersFunc != nil {
		return m.FetchUsersFunc(ctx)
	}
	return nil, errors.New("mock FetchUsers not implemented")
}

func (m *MockAPIClient) SendUserToAPIB(
	ctx context.Context,
	payload model.UserPayload,
	apiBURL string,
) error {
	m.SentPayloads = append(m.SentPayloads, payload)
	if m.SendUserToAPIBFunc != nil {
		return m.SendUserToAPIBFunc(ctx, payload, apiBURL)
	}
	
	return nil
}

func TestDispatcher_ProcessAndDispatchUsers_FilterBizEmails(t *testing.T) {
	users := []model.User{
		{ID: 1, Name: "Leanne Graham", Email: "Sincere@april.biz"},
		{ID: 2, Name: "Ervin Howell", Email: "Shanna@melissa.tv"},
		{ID: 3, Name: "Clementine Bauch", Email: "Nathan@yesenia.net"},
		{ID: 4, Name: "Patricia Lebsack", Email: "Julianne.OConner@kory.org"},
		{ID: 5, Name: "Chelsey Dietrich", Email: "Lucio_Hettinger@annie.ca"},
		{ID: 6, Name: "Mrs. Dennis Schulist", Email: "Karley_Dach@jasper.info"},
		{ID: 7, Name: "Kurtis Weissnat", Email: "Telly.Hoeger@billy.biz"},
	}

	mockClient := &MockAPIClient{
		FetchUsersFunc: func(ctx context.Context) ([]model.User, error) {
			return users, nil
		},
		SendUserToAPIBFunc: func(ctx context.Context, payload model.UserPayload, apiBURL string) error {
			return nil
		},
		SentPayloads: []model.UserPayload{},
	}

	dispatcher := NewDispatcher(mockClient, "https://mock.api.b")

	var logOutput strings.Builder
	log.SetOutput(&logOutput)
	originalFlags := log.Flags()
	log.SetFlags(0)
	defer func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(originalFlags)
	}()

	err := dispatcher.ProcessAndDispatchUsers(context.Background())
	if err != nil {
		t.Fatalf("ProcessAndDispatchUsers error: %v", err)
	}

	expectedSentCount := 2
	if len(mockClient.SentPayloads) != expectedSentCount {
		t.Errorf("Expected %d payloads sent, got %d", expectedSentCount, len(mockClient.SentPayloads))
	}

	wasSent := func(email string) bool {
		for _, p := range mockClient.SentPayloads {
			if p.Email == email {
				return true
			}
		}
		return false
	}

	if !wasSent("Sincere@april.biz") {
		t.Errorf("Email Sincere@april.biz not sent")
	}
	if !wasSent("Telly.Hoeger@billy.biz") {
		t.Errorf("Email Telly.Hoeger@billy.biz not sent")
	}
	if wasSent("Shanna@melissa.tv") {
		t.Errorf("Email Shanna@melissa.tv should not have been sent")
	}

	logs := logOutput.String()
	if !strings.Contains(logs, "User Shanna@melissa.tv (not .biz): skipping") {
		t.Errorf("Log for Shanna@melissa.tv not found. Logs: %s", logs)
	}
	if !strings.Contains(logs, "User Sincere@april.biz (.biz): sending to API B") {
		t.Errorf("Log for Sincere@april.biz processing not found. Logs: %s", logs)
	}
	if !strings.Contains(logs, "User Leanne Graham (Sincere@april.biz) sent to API B") {
		t.Errorf("Success log for Sincere@april.biz not found. Logs: %s", logs)
	}
}

func TestDispatcher_ProcessAndDispatchUsers_FetchError(t *testing.T) {
	mockClient := &MockAPIClient{
		FetchUsersFunc: func(ctx context.Context) ([]model.User, error) {
			return nil, errors.New("mock fetch error")
		},
	}

	dispatcher := NewDispatcher(mockClient, "https://mock.api.b")

	err := dispatcher.ProcessAndDispatchUsers(context.Background())
	if err == nil {
		t.Fatal("Expected error on FetchUsers failure, got nil")
	}
	if !strings.Contains(err.Error(), "mock fetch error") {
		t.Errorf("Expected error message 'mock fetch error', got: %s", err.Error())
	}
}

func TestDispatcher_ProcessAndDispatchUsers_SendError(t *testing.T) {
	users := []model.User{
		{ID: 1, Name: "Biz User", Email: "test@example.biz"},
	}

	mockClient := &MockAPIClient{
		FetchUsersFunc: func(ctx context.Context) ([]model.User, error) {
			return users, nil
		},
		SendUserToAPIBFunc: func(
			ctx context.Context, 
			payload model.UserPayload, 
			apiBURL string,
		) error {
			return errors.New("mock send error")
		},
		SentPayloads: []model.UserPayload{},
	}

	dispatcher := NewDispatcher(mockClient, "https://mock.api.b")

	var logOutput strings.Builder
	log.SetOutput(&logOutput)
	originalFlags := log.Flags()
	log.SetFlags(0)
	defer func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(originalFlags)
	}()

	err := dispatcher.ProcessAndDispatchUsers(context.Background())
	if err != nil {
		t.Fatalf("ProcessAndDispatchUsers error (send failure): %v", err)
	}

	if len(mockClient.SentPayloads) != 1 {
		t.Errorf("Expected SendUserToAPIB call count 1, got %d", len(mockClient.SentPayloads))
	}

	logs := logOutput.String()
	if !strings.Contains(logs, "Send to API B error for Biz User (test@example.biz): mock send error") {
		t.Errorf("Log for failed send not found. Logs: %s", logs)
	}
} 