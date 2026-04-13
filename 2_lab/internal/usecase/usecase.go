package usecase

import (
	"context"
	"encoding/json"
	"log"

	"lab2/internal/repo/pg"
)

type Repo interface {
	GetAll(ctx context.Context) ([]pg.Note, error)
	Create(ctx context.Context, title, body string) (pg.Note, error)
	Delete(ctx context.Context, id int) error
}

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
	Delete(ctx context.Context, key string) error
	NotesKey() string
}

type UseCase struct {
	repo  Repo
	cache Cache
}

func New(repo Repo, cache Cache) *UseCase {
	return &UseCase{repo: repo, cache: cache}
}

func (uc *UseCase) GetNotes(ctx context.Context) ([]pg.Note, error) {
	// 1. Try cache
	if raw, err := uc.cache.Get(ctx, uc.cache.NotesKey()); err == nil {
		var notes []pg.Note
		if json.Unmarshal(raw, &notes) == nil {
			log.Println("[cache] HIT notes:all")
			return notes, nil
		}
	}

	// 2. Cache miss — go to DB
	log.Println("[cache] MISS notes:all — querying DB")
	notes, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if notes == nil {
		notes = []pg.Note{}
	}

	// 3. Populate cache
	if raw, err := json.Marshal(notes); err == nil {
		_ = uc.cache.Set(ctx, uc.cache.NotesKey(), raw)
	}

	return notes, nil
}

func (uc *UseCase) CreateNote(ctx context.Context, title, body string) (pg.Note, error) {
	note, err := uc.repo.Create(ctx, title, body)
	if err != nil {
		return pg.Note{}, err
	}
	// Invalidate cache
	log.Println("[cache] INVALIDATE notes:all (create)")
	_ = uc.cache.Delete(ctx, uc.cache.NotesKey())
	return note, nil
}

func (uc *UseCase) DeleteNote(ctx context.Context, id int) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	// Invalidate cache
	log.Println("[cache] INVALIDATE notes:all (delete)")
	_ = uc.cache.Delete(ctx, uc.cache.NotesKey())
	return nil
}
