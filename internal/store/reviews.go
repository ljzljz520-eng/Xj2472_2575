package store

import (
	"fmt"

	"coldchain-alert/internal/domain"
	bolt "go.etcd.io/bbolt"
)

func (s *Store) SaveReview(review domain.AlertReview) error {
	if err := domain.ValidateReview(review); err != nil {
		return err
	}
	if alert, err := s.GetAlert(review.AlertID); err != nil {
		return fmt.Errorf("review alert: %w", err)
	} else if alert == nil {
		return domain.ErrNotFound
	}
	return s.withUpdate(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(BucketReviews)).Get([]byte(review.ID)) != nil {
			return fmt.Errorf("%w: review %s", domain.ErrAlreadyExists, review.ID)
		}
		return putJSON(tx, BucketReviews, review.ID, review)
	})
}

func (s *Store) GetReview(id string) (*domain.AlertReview, error) {
	var review domain.AlertReview
	err := s.withView(func(tx *bolt.Tx) error { return getJSON(tx, BucketReviews, id, &review) })
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (s *Store) ListReviews() ([]domain.AlertReview, error) {
	var reviews []domain.AlertReview
	err := s.withView(func(tx *bolt.Tx) error {
		var err error
		reviews, err = listJSON[domain.AlertReview](tx, BucketReviews)
		return err
	})
	return reviews, err
}

func (s *Store) ListReviewsByAlert(alertID string) ([]domain.AlertReview, error) {
	reviews, err := s.ListReviews()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.AlertReview, 0, len(reviews))
	for _, review := range reviews {
		if review.AlertID == alertID {
			filtered = append(filtered, review)
		}
	}
	return filtered, nil
}
