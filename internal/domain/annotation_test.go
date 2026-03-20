package domain_test

import (
	"errors"
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
)

func TestNewBookMark(t *testing.T) {
	// i guess i have to create a new bookmark first , let's go
	_, bookmarkCreationError := domain.NewBookMark("3ubiosdfo24", domain.AnnotationBookmark, 30)
	if bookmarkCreationError != nil {
		t.Fatalf("expected no errors, got %v", bookmarkCreationError)
	}

	// now let's test the case of improper bookID
	_, bookmarkCreationWrongBookId := domain.NewBookMark("", domain.AnnotationBookmark, 30)
	if !errors.Is(bookmarkCreationWrongBookId, domain.ErrInvalidBookId) {
		t.Errorf("expected error to be ErrInvalidBookId, instead got %v", bookmarkCreationWrongBookId)
	}

	// now let's test the case of the improper page Number
	_, pageNumberErr := domain.NewBookMark("3ubiosdfo24", domain.AnnotationBookmark, -2)
	if !errors.Is(pageNumberErr, domain.ErrInvalidPageNumber) {
		t.Errorf("expected error to be ErrInvalidPageNumber, instead got %v", pageNumberErr)
	}

}

func TestBookMarkType(t *testing.T) {
	// we are checking if the book got the correct value for the bookmark type
	bookmark, bookmarkCreationError := domain.NewBookMark("kln934nalj4fs", domain.AnnotationBookmark, 30)
	if bookmarkCreationError != nil {
		t.Fatalf("expected no errors, got %v", bookmarkCreationError)
	}

	if bookmark.AnnotationType != "bookmark" {
		t.Fatalf("bookmark.Type should  be 'bookmark', instead got %v", bookmark.AnnotationType)
	}
	// now let's cover for the other type of bookmark
	bookmark2, bookmarkCreationError2 := domain.NewBookMark("kln934nalj4fs", domain.AnnotationHighlight, 30)
	if bookmarkCreationError2 != nil {
		t.Fatalf("expected no errors, got %v", bookmarkCreationError2)
	}

	if bookmark2.AnnotationType != "highlight" {
		t.Fatalf("bookmark.Type should be 'hightlight', instead got %v", bookmark2.AnnotationType)
	}
}
