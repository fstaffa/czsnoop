package search

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fstaffa/czsnoop/internal/rzp"
	"github.com/fstaffa/czsnoop/internal/types"
)

type SearchInput struct {
	Ico        types.Ico
	Query      string
	BornAfter  time.Time
	BornBefore time.Time
}

type Person struct {
	Citizenship     string
	BirthDate       time.Time
	FirstName       string
	LastName        string
	TitleBeforeName string
	TitleAfterName  string
	FullName        string
	Address         string
	Ico             types.Ico
	Trades          []string
}

type SearchResult struct {
	// If true, this search was not able to go through all matches. If relevant match wasn't found
	// try to add more details
	MorePossibleNatches bool
	People              []Person
}

func SearchRzp(input SearchInput, logger *slog.Logger) ([]Person, error) {
	context, cancel := context.WithCancelCause(context.Background())
	logger = logger.With("search", "rzp")
	defer cancel(nil)
	client, err := rzp.CreateClient(context)
	if err != nil {
		return nil, fmt.Errorf("unable to create RZP client: %v", err)
	}
	var searchQuery rzp.SearchSubjectQuery
	if input.Ico != "" {
		searchQuery.Ico = input.Ico
	} else {
		searchQuery.Name = input.Query
	}

	searchResult, err := rzpWideSearch(searchQuery, client, cancel, logger)
	if err != nil {
		return nil, err
	}
	logger.Debug("Found subjects", slog.Int("count", len(searchResult)))

	resultWithDetails := rzpDeepSearch(client, searchResult, logger)
	persons := make([]Person, 0, len(resultWithDetails))
	for _, person := range resultWithDetails {
		if !input.BornAfter.IsZero() && person.BirthDate.Before(input.BornAfter) {
			logger.Debug("Skipping person because they were born before required time",
				slog.String("name", person.FullNameWithTitles), slog.String("ico", string(person.Ico)), slog.Time("birthDate", person.BirthDate), slog.Time("bornAfter", input.BornAfter))
			continue
		}
		if !input.BornBefore.IsZero() && person.BirthDate.After(input.BornBefore) {
			logger.Debug("Skipping person because they were born after required time",
				slog.String("name", person.FullNameWithTitles), slog.String("ico", string(person.Ico)), slog.Time("birthDate", person.BirthDate), slog.Time("bornBefore", input.BornBefore))
			continue
		}
		trades := make([]string, 0, len(person.Trades))
		for _, trade := range person.Trades {
			trades = append(trades, trade.TradeType)
		}
		persons = append(persons, Person{
			FirstName:       person.FirstName,
			LastName:        person.LastName,
			TitleBeforeName: person.TitleBeforeName,
			TitleAfterName:  person.TitleAfterName,
			Address:         person.Address,
			Ico:             person.Ico,
			Citizenship:     person.Citizenship,
			Trades:          trades,
			BirthDate:       person.BirthDate,
			FullName:        person.FullNameWithTitles,
		})
	}

	return persons, nil
}

func rzpWideSearch(searchQuery rzp.SearchSubjectQuery, client *rzp.Rzp, cancel context.CancelCauseFunc, logger *slog.Logger) ([]rzp.Subject, error) {
	results := make(chan types.Result[rzp.SearchSubjectResponse])
	searchFunc := func(subjectType string) {
		enterpreneurQuery := searchQuery
		enterpreneurQuery.SubjectType = subjectType
		subjects, err := client.SearchSubject(enterpreneurQuery)
		if err != nil {
			err := fmt.Errorf("unable to search enterpreneurs in RZP: %v", err)
			cancel(err)
			results <- types.Result[rzp.SearchSubjectResponse]{
				Err: err,
			}
		} else {
			results <- types.Result[rzp.SearchSubjectResponse]{Result: subjects}
		}
	}
	go searchFunc("enterpreneur")
	go searchFunc("statutory body")

	searchResult := make([]rzp.Subject, 0)
	for i := 0; i < 2; i++ {
		result := <-results
		if result.Err != nil {
			return nil, result.Err
		}
		if result.Result.MorePossibleMatches {
			return nil, fmt.Errorf("too many possible matches, please provide more details")
		}
		for _, subject := range result.Result.Subjects {
			searchResult = append(searchResult, subject)
		}

	}
	return searchResult, nil
}

func rzpDeepSearch(r *rzp.Rzp, subjects []rzp.Subject, logger *slog.Logger) []rzp.SubjectDetail {
	parallelization := min(6, len(subjects))
	inputs := make(chan rzp.Subject, parallelization)
	resultChan := make(chan rzp.SubjectDetail)

	go func() {
		for _, subject := range subjects {
			inputs <- subject
		}
		close(inputs)
	}()

	var wg sync.WaitGroup
	for i := 0; i < parallelization; i++ {
		wg.Add(1)
		go func() {
			for input := range inputs {
				logger.Debug("Getting details for subject", slog.String("ico", string(input.Ico)), slog.Int("worker", i), slog.Int("remaining", len(inputs)))
				result, err := r.GetSubjectDetails(input.Ssarzp)
				logger.Debug("Got details for subject", slog.String("ico", string(input.Ico)), slog.Int("worker", i), slog.Int("remaining", len(inputs)))
				if err != nil {
					panic(err)
				} else {
					resultChan <- result
				}
			}
			logger.Debug("Worker done", slog.Int("worker", i))
			wg.Done()
		}()
	}

	go func() {
		logger.Debug("Waiting for workers to finish")
		wg.Wait()
		logger.Debug("All workers done")
		close(resultChan)
	}()

	results := make([]rzp.SubjectDetail, 0, len(subjects))
	for result := range resultChan {
		results = append(results, result)
	}
	return results
}
