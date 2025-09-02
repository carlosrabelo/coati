package services

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGistFetcher struct {
	mock.Mock
}

func (m *MockGistFetcher) Fetch(gistID, token, gistFile string) ([]byte, error) {
	args := m.Called(gistID, token, gistFile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestCachedGistFetcher_Fetch(t *testing.T) {
	tmpDir := t.TempDir()
	mockDelegate := new(MockGistFetcher)
	ttl := 1 * time.Second

	fetcher := NewCachedGistFetcher(mockDelegate, tmpDir, ttl)

	gistID := "test-gist"
	token := "token"
	content := []byte("remote-content")

	// Case 1: Cache Miss (First Call)
	mockDelegate.On("Fetch", gistID, token, "").Return(content, nil).Once()

	result, err := fetcher.Fetch(gistID, token, "")
	assert.NoError(t, err)
	assert.Equal(t, content, result)

	// Verify file created
	cachePath := filepath.Join(tmpDir, gistID+".json")
	assert.FileExists(t, cachePath)

	// Case 2: Cache Hit (Immediate Second Call)
	// Delegate should NOT be called again (asserted by .Once() above)
	result2, err := fetcher.Fetch(gistID, token, "")
	assert.NoError(t, err)
	assert.Equal(t, content, result2)

	// Case 3: Cache Expiration
	time.Sleep(1100 * time.Millisecond) // Wait for TTL

	newContent := []byte("new-remote-content")
	mockDelegate.On("Fetch", gistID, token, "").Return(newContent, nil).Once()

	result3, err := fetcher.Fetch(gistID, token, "")
	assert.NoError(t, err)
	assert.Equal(t, newContent, result3)

	mockDelegate.AssertExpectations(t)
}
