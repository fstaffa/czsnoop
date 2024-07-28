package search

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
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
	Subjects        []rzp.Subject
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
	client, err := rzp.CreateClient(context, logger.With("client", "rzp"))
	if err != nil {
		return nil, fmt.Errorf("unable to create RZP client: %v", err)
	}
	var searchQuery rzp.SearchSubjectQuery
	if input.Query != "" {
		searchQuery.Name = input.Query
	}
	personsSearch, err := rzpPersonSearch(searchQuery, client, cancel, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to search persons in RZP: %v", err)
	}

	wg := sync.WaitGroup{}
	resultChan := make(chan types.Result[Person])
	for _, person := range personsSearch {
		logger.Debug("Searching subjects for person", slog.String("person", person.DisplayName))
		wg.Add(1)
		go func() {
			subjects, err := client.SearchSubject(rzp.SearchSubjectQuery{
				PersonId: person.PersonId,
			})
			if err != nil {
				cancel(err)
				resultChan <- types.Result[Person]{Err: err}
			}
			person := Person{
				BirthDate:       time.Time(person.DateOfBirth),
				FirstName:       person.FirstName,
				LastName:        person.LastName,
				TitleBeforeName: person.TitleBeforeName,
				TitleAfterName:  person.TitleAfterName,
				FullName:        person.DisplayName,
				Subjects:        subjects.Subjects,
			}
			for _, subject := range subjects.Subjects {
				if subject.Type == "F" {
					subjectDetail, err := client.GetSubjectDetails(subject.Ssarzp)
					if err != nil {
						cancel(err)
						resultChan <- types.Result[Person]{Err: err}
					}
					person.Address = subject.Address
					person.Citizenship = subjectDetail.Citizenship
				}
			}

			resultChan <- types.Result[Person]{Result: person}
			wg.Done()
			logger.Debug("Done searching subjects for person", slog.String("person", person.FullName))
		}()
	}

	go func() {
		logger.Debug("Waiting for subjects for all persons to be found")
		wg.Wait()
		logger.Debug("Subjects for all persons found")
		close(resultChan)
	}()

	persons := make([]Person, 0, len(personsSearch))
	for result := range resultChan {
		persons = append(persons, result.Result)
	}

	return persons, nil
}

func rzpPersonSearch(searchQuery rzp.SearchSubjectQuery, client *rzp.Rzp, cancel context.CancelCauseFunc, logger *slog.Logger) ([]rzp.Person, error) {
	personQuery := rzp.SearchPersonQuery{}
	if searchQuery.Name != "" {
		parts := strings.Split(searchQuery.Name, " ")
		firstName := strings.Join(parts[0:(len(parts)-1)], " ")
		surname := parts[len(parts)-1]
		personQuery.FirstName = firstName
		personQuery.Surname = surname
	}

	persons, err := client.SearchPerson(personQuery)
	if err != nil {
		err := fmt.Errorf("unable to search persons in RZP: %v", err)
		cancel(err)
		return nil, err
	}
	if persons.MorePossibleMatches {
		return nil, fmt.Errorf("too many possible matches, please provide more details")
	}

	logger.Debug("Found persons", slog.Int("count", len(persons.People)))
	return persons.People, nil
}
