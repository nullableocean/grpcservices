package dataload

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

/*
# Параллельная загрузка данных из нескольких источников
---
## Описание задачи
Реализовать систему параллельной загрузки данных из независимых источников:
1. Асинхронная загрузка комментариев из БД
2. Параллельная загрузка данных пользователей на основе полученных комментариев
3. Загрузка данных сессии и условная загрузка вложений

**Цель:**
Освоить работу с горутинами, `sync.Once` и синхронизацией через `sync.WaitGroup`.

---
## Требования
1. Загрузка комментариев и данных сессии должна выполняться параллельно
2. Загрузка данных пользователей должна стартовать только после получения комментариев
3. Загрузка вложений должна выполняться только при наличии session-id
4. Использовать минимум 3 горутины для разных этапов
5. Синхронизировать все операции перед завершением
*/

var (
	canRetries = 3
)

var (
	ErrLoading      = errors.New("error in process load")
	ErrPanicLoading = errors.New("error in process load with panic")
)

type CommentStore interface {
	GetComments(context.Context) ([]*Comment, error)
}

type UserStore interface {
	GetUsers(context.Context, ...UserId) []*User
}

type SessionStore interface {
	GetCurrent() *Session
}

type AttachmentStore interface {
	GetAttachements(context.Context, ...CommentId) []*Attachment
}

type LoadedData struct {
	Comments    []*Comment
	Users       map[UserId]*User
	Attachments map[CommentId][]*Attachment
	Session     *Session
}

type OnceCommentsLoader struct {
	userStore    UserStore
	commentStore CommentStore
	sesStore     SessionStore
	attachStore  AttachmentStore

	loaded   *LoadedData
	loadOnce *sync.Once

	loadErr error
	retries int

	mu sync.Mutex
}

func NewOnceCommentLoader(us UserStore, cs CommentStore, ss SessionStore, atts AttachmentStore) *OnceCommentsLoader {
	return &OnceCommentsLoader{
		userStore:    us,
		commentStore: cs,
		sesStore:     ss,
		attachStore:  atts,

		loaded:   nil,
		loadOnce: &sync.Once{},

		loadErr: nil,
		retries: canRetries,

		mu: sync.Mutex{},
	}
}

func (loader *OnceCommentsLoader) Load(ctx context.Context) (*LoadedData, error) {
	loader.mu.Lock()
	defer loader.mu.Unlock()

	loader.loadOnce.Do(func() {
		loader.loadErr = loader.load(ctx)
	})

	if loader.loadErr != nil {
		if loader.retries > 0 {
			loader.resetOnce()
			loader.retries -= 1
		}
	}

	return loader.loaded, loader.loadErr
}

func (loader *OnceCommentsLoader) resetOnce() {
	loader.loadOnce = &sync.Once{}
}

func (loader *OnceCommentsLoader) load(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	atomicErr := atomic.Value{}
	errCh := make(chan error, 1)

	// обрабатываем ошибку в горутине
	// если есть ошибка отменяем контекст
	go func() {
		select {
		case <-errCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	var comments []*Comment
	var users map[UserId]*User
	var session *Session
	var attachments map[CommentId][]*Attachment

	wg := sync.WaitGroup{}

	// комменты
	wg.Add(1)
	commentsDone := make(chan struct{})
	go func() {
		defer wg.Done()
		defer close(commentsDone)

		// на случай паники
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%w: panic in comments loading: %v", ErrPanicLoading, r)
				if atomicErr.CompareAndSwap(nil, err) {
					errCh <- err
				}
			}
		}()

		comms, err := loader.getComments(ctx)

		if err != nil {
			if atomicErr.CompareAndSwap(nil, err) {
				errCh <- err
			}
			return
		}

		comments = comms
	}()

	// сессия
	wg.Add(1)
	sessionDone := make(chan struct{})
	go func() {
		defer wg.Done()
		defer close(sessionDone)

		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%w: panic in session loading: %v", ErrPanicLoading, r)
				if atomicErr.CompareAndSwap(nil, err) {
					errCh <- err
				}
			}
		}()

		session = loader.getSession()
	}()

	// юзеры
	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%w: panic in users loading: %v", ErrPanicLoading, r)
				if atomicErr.CompareAndSwap(nil, err) {
					errCh <- err
				}
			}
		}()

		select {
		case <-ctx.Done():
			return
		case <-commentsDone:
		}

		users = loader.getUsersByComments(ctx, comments)
	}()

	// вложения
	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%w: panic in attachments loading: %v", ErrPanicLoading, r)
				if atomicErr.CompareAndSwap(nil, err) {
					errCh <- err
				}
			}
		}()

		select {
		case <-ctx.Done():
			return
		case <-commentsDone:
		}

		select {
		case <-ctx.Done():
			return
		case <-sessionDone:
		}

		if session == nil || session.SessionId.IsEmpty() {
			return
		}

		attachments = loader.getAttachmentsForComments(ctx, comments)
	}()

	// ждем завершения
	loadDone := make(chan struct{})
	go func() {
		defer close(loadDone)
		wg.Wait()
	}()

	select {
	case <-ctx.Done():
		e := ctx.Err()

		// если отменили локальный контекст из-за ошибки при загрузке
		if err := atomicErr.Load(); err != nil {
			e = err.(error)
		}

		return e
	case <-loadDone:
		if err := atomicErr.Load(); err != nil {
			return err.(error)
		}
	}

	loader.loaded = &LoadedData{
		Comments:    comments,
		Users:       users,
		Attachments: attachments,
		Session:     session,
	}

	return nil
}

func (loader *OnceCommentsLoader) getComments(ctx context.Context) ([]*Comment, error) {
	return loader.commentStore.GetComments(ctx)
}

func (loader *OnceCommentsLoader) getSession() *Session {
	return loader.sesStore.GetCurrent()
}

func (loader *OnceCommentsLoader) getUsersByComments(ctx context.Context, comments []*Comment) map[UserId]*User {
	if len(comments) == 0 {
		return nil
	}

	ids := make([]UserId, 0, len(comments))
	usersMap := make(map[UserId]*User, len(comments))

	for _, c := range comments {
		if _, ex := usersMap[c.UserId]; ex {
			continue
		}

		usersMap[c.UserId] = nil
		ids = append(ids, c.UserId)
	}

	loadedUsers := loader.userStore.GetUsers(ctx, ids...)

	for _, u := range loadedUsers {
		usersMap[u.Id] = u
	}

	return usersMap
}

func (loader *OnceCommentsLoader) getAttachmentsForComments(ctx context.Context, comments []*Comment) map[CommentId][]*Attachment {
	if len(comments) == 0 {
		return nil
	}

	commentsIds := make([]CommentId, 0, len(comments))
	attachMap := make(map[CommentId][]*Attachment, len(comments))

	for _, c := range comments {
		commentsIds = append(commentsIds, c.Id)
	}

	attachments := loader.attachStore.GetAttachements(ctx, commentsIds...)
	attchAvg := len(attachments)/len(comments) + 1 // среднее значение для капасити в слайсе

	for _, a := range attachments {
		commAts, ex := attachMap[a.CommentId]

		if !ex {
			commAts = make([]*Attachment, 0, attchAvg)
		}

		attachMap[a.CommentId] = append(commAts, a)
	}

	return attachMap
}
