package dataload

import (
	"context"
	"errors"
	"testing"
)

// Моки для интерфейсов

type MockCommentStore struct {
	comments []*Comment
	err      error
}

func (m *MockCommentStore) GetComments(ctx context.Context) ([]*Comment, error) {
	return m.comments, m.err
}

type MockUserStore struct {
	users map[UserId]*User
}

func (m *MockUserStore) GetUsers(ctx context.Context, ids ...UserId) []*User {
	var result []*User
	for _, id := range ids {
		if user, ok := m.users[id]; ok {
			result = append(result, user)
		}
	}
	return result
}

type MockSessionStore struct {
	session *Session
}

func (m *MockSessionStore) GetCurrent() *Session {
	return m.session
}

type MockAttachmentStore struct {
	attachments map[CommentId][]*Attachment
	err         error
}

func (m *MockAttachmentStore) GetAttachements(ctx context.Context, ids ...CommentId) []*Attachment {
	var result []*Attachment
	for _, id := range ids {
		if attachments, ok := m.attachments[id]; ok {
			result = append(result, attachments...)
		}
	}
	return result
}

// Тесты

func TestOnceCommentsLoader_Load(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		testComments := []*Comment{
			{Id: 1, UserId: 101},
			{Id: 2, UserId: 102},
		}
		testUsers := map[UserId]*User{
			101: {Id: 101},
			102: {Id: 102},
		}
		testSession := &Session{SessionId: "session123"}
		testAttachments := map[CommentId][]*Attachment{
			1: {{Id: 1001, CommentId: 1}},
			2: {{Id: 1002, CommentId: 2}},
		}

		mockCommentStore := &MockCommentStore{comments: testComments}
		mockUserStore := &MockUserStore{users: testUsers}
		mockSessionStore := &MockSessionStore{session: testSession}
		mockAttachmentStore := &MockAttachmentStore{attachments: testAttachments}

		loader := NewOnceCommentLoader(mockUserStore, mockCommentStore, mockSessionStore, mockAttachmentStore)

		data, err := loader.Load(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(data.Comments) != len(testComments) {
			t.Errorf("expected %d comments, got %d", len(testComments), len(data.Comments))
		}
		if len(data.Users) != len(testUsers) {
			t.Errorf("expected %d users, got %d", len(testUsers), len(data.Users))
		}
		if data.Session.SessionId != testSession.SessionId {
			t.Errorf("expected session ID %s, got %s", testSession.SessionId, data.Session.SessionId)
		}
		if len(data.Attachments) != len(testAttachments) {
			t.Errorf("expected %d attachment groups, got %d", len(testAttachments), len(data.Attachments))
		}
	})

	t.Run("empty comments", func(t *testing.T) {
		mockCommentStore := &MockCommentStore{comments: []*Comment{}}
		mockUserStore := &MockUserStore{users: map[UserId]*User{}}
		mockSessionStore := &MockSessionStore{session: &Session{SessionId: "session123"}}
		mockAttachmentStore := &MockAttachmentStore{attachments: map[CommentId][]*Attachment{}}

		loader := NewOnceCommentLoader(mockUserStore, mockCommentStore, mockSessionStore, mockAttachmentStore)

		data, err := loader.Load(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if data.Users != nil {
			t.Error("expected users to be nil for empty comments")
		}
		if data.Attachments != nil {
			t.Error("expected attachments to be nil for empty comments")
		}
	})

	t.Run("nil stores interfases", func(t *testing.T) {
		testComments := []*Comment{
			{Id: 1, UserId: 101},
		}
		mockCommentStore := &MockCommentStore{comments: testComments}
		mockSessionStore := &MockSessionStore{session: nil}

		var userNilPtr *MockUserStore
		var userStore UserStore = userNilPtr

		var attchStore AttachmentStore

		loader := NewOnceCommentLoader(userStore, mockCommentStore, mockSessionStore, attchStore)
		data, err := loader.Load(context.Background())
		if err == nil {
			t.Fatalf("expected load error. got err: %v, got data: %v", err, data)
		}

		if !errors.Is(err, ErrPanicLoading) {
			t.Fatalf("expected panic load error. got: %v", err)
		}
	})

	t.Run("get comments error and retry", func(t *testing.T) {
		testUsers := map[UserId]*User{
			101: {Id: 101},
			102: {Id: 102},
		}
		testSession := &Session{SessionId: "session123"}
		testAttachments := map[CommentId][]*Attachment{
			1: {{Id: 1001, CommentId: 1}},
			2: {{Id: 1002, CommentId: 2}},
		}

		mockUserStore := &MockUserStore{users: testUsers}
		mockSessionStore := &MockSessionStore{session: testSession}
		mockAttachmentStore := &MockAttachmentStore{attachments: testAttachments}
		mockCommentStore := &MockCommentStore{err: errors.New("first error")}

		loader := NewOnceCommentLoader(mockUserStore, mockCommentStore, mockSessionStore, mockAttachmentStore)

		_, err := loader.Load(context.Background())
		if err == nil {
			t.Error("expected error on first comments load, got nil")
		}

		mockCommentStore.err = nil
		mockCommentStore.comments = []*Comment{{Id: 1, UserId: 101}}

		data, err := loader.Load(context.Background())
		if err != nil {
			t.Fatalf("unexpected error on second load: %v", err)
		}

		if len(data.Comments) != 1 {
			t.Errorf("expected 1 comment, got %d", len(data.Comments))
		}
	})
}
