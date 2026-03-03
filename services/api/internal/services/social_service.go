package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"example.com/api/internal/repositories"
)

var (
	ErrCannotLikeSelf   = errors.New("cannot like yourself")
	ErrCannotBlockSelf  = errors.New("cannot block yourself")
	ErrInteractionBlock = errors.New("interaction blocked")
)

type DiscoverUser struct {
	ID        string
	Email     string
	CreatedAt time.Time
}

type SocialService struct {
	repo repositories.SocialRepository
}

func NewSocialService(repo repositories.SocialRepository) *SocialService {
	return &SocialService{repo: repo}
}

func (s *SocialService) Discover(ctx context.Context, userID string) ([]DiscoverUser, error) {
	users, err := s.repo.ListDiscoverUsers(ctx, userID, 50)
	if err != nil {
		return nil, err
	}

	payload := make([]DiscoverUser, 0, len(users))
	for _, user := range users {
		payload = append(payload, DiscoverUser{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}
	return payload, nil
}

func (s *SocialService) Like(ctx context.Context, fromUserID, toUserID string) (bool, error) {
	normalizedToUserID := strings.TrimSpace(toUserID)
	if normalizedToUserID == fromUserID {
		return false, ErrCannotLikeSelf
	}

	exists, err := s.repo.UserExists(ctx, normalizedToUserID)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, ErrUserNotFound
	}

	blocked, err := s.repo.UsersBlocked(ctx, fromUserID, normalizedToUserID)
	if err != nil {
		return false, err
	}
	if blocked {
		return false, ErrInteractionBlock
	}

	if err := s.repo.InsertLike(ctx, fromUserID, normalizedToUserID); err != nil {
		return false, err
	}

	matched, err := s.repo.HasReciprocalLike(ctx, fromUserID, normalizedToUserID)
	if err != nil {
		return false, err
	}

	return matched, nil
}

func (s *SocialService) Block(ctx context.Context, blockerUserID, blockedUserID string) error {
	normalizedBlockedUserID := strings.TrimSpace(blockedUserID)
	if normalizedBlockedUserID == blockerUserID {
		return ErrCannotBlockSelf
	}

	exists, err := s.repo.UserExists(ctx, normalizedBlockedUserID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	return s.repo.BlockAndDeleteLikes(ctx, blockerUserID, normalizedBlockedUserID)
}

func (s *SocialService) Matches(ctx context.Context, userID string) ([]DiscoverUser, error) {
	matches, err := s.repo.ListMatches(ctx, userID)
	if err != nil {
		return nil, err
	}

	payload := make([]DiscoverUser, 0, len(matches))
	for _, user := range matches {
		payload = append(payload, DiscoverUser{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		})
	}
	return payload, nil
}
