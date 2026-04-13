package usecase

import (
	"context"

	"arch/internal/repo/pg"
)

type Repo interface {
	GetAll(ctx context.Context) ([]pg.Note, error)
	Create(ctx context.Context, title, body string) (pg.Note, error)
}

type UseCase struct {
	repo Repo
}

func New(repo Repo) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) GetNotes(ctx context.Context) ([]pg.Note, error) {
	return uc.repo.GetAll(ctx)
}

func (uc *UseCase) CreateNote(ctx context.Context, title, body string) (pg.Note, error) {
	return uc.repo.Create(ctx, title, body)
}
