package usecase

import (
	"context"

	"notes-api/internal/repo/pg"
)

type Repo interface {
	GetAll(ctx context.Context) ([]pg.Note, error)
	Create(ctx context.Context, title, body string) (pg.Note, error)
	Delete(ctx context.Context, id int) error
}

type Publisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
}

type UseCase struct {
	repo      Repo
	publisher Publisher
}

func New(repo Repo, publisher Publisher) *UseCase {
	return &UseCase{repo: repo, publisher: publisher}
}

func (uc *UseCase) GetNotes(ctx context.Context) ([]pg.Note, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *UseCase) CreateNote(ctx context.Context, title, body string) (pg.Note, error) {
	note, err := uc.repo.Create(ctx, title, body)
	if err != nil {
		return pg.Note{}, err
	}
	_ = uc.publisher.Publish(ctx, "note.created", note)
	return note, nil
}

func (uc *UseCase) DeleteNote(ctx context.Context, id int) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = uc.publisher.Publish(ctx, "note.deleted", map[string]int{"id": id})
	return nil
}
