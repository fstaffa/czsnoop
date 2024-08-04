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

type PersonSearchInput struct {
	Ico        types.Ico
	Query      string
	BornAfter  time.Time
	BornBefore time.Time
}

type Person struct {
	Citizenship         string
	BirthDate           time.Time
	FirstName           string
	LastName            string
	TitleBeforeName     string
	TitleAfterName      string
	FullName            string
	Address             string
	Subjects            []EconomicSubject
	SameAddressSubjects []string
}

type EconomicSubject struct {
	Name    string
	Address string
	Ico     types.Ico
}

func Rzp(input PersonSearchInput, logger *slog.Logger) ([]Person, error) {
	ctx, cancel := context.WithCancelCause(context.Background())
	logger = logger.With("search", "rzp")
	defer cancel(nil)
	client, err := rzp.CreateClient(ctx, logger.With("client", "rzp"))
	if err != nil {
		return nil, fmt.Errorf("unable to create RZP client: %v", err)
	}
	var searchQuery rzp.SearchSubjectQuery
	if input.Query != "" {
		searchQuery.Name = input.Query
	}
	rzpPersons, err := rzpPersonSearch(searchQuery, client, cancel, logger)
	if err != nil {
		return nil, fmt.Errorf("unable to search persons in RZP: %v", err)
	}

	wg := sync.WaitGroup{}
	resultChan := make(chan types.Result[Person])
	for _, rzpPerson := range rzpPersons {
		logger.Debug("Searching subjects for person", slog.String("person", rzpPerson.DisplayName))
		wg.Add(1)
		go func() {
			subjects, err := client.SearchSubject(rzp.SearchSubjectQuery{
				PersonId: rzpPerson.PersonId,
			})
			if err != nil {
				cancel(err)
				resultChan <- types.Result[Person]{Err: err}
			}
			economicSubjects := make([]EconomicSubject, 0, len(subjects.Subjects))
			for _, subject := range subjects.Subjects {
				economicSubjects = append(economicSubjects, EconomicSubject{
					Name:    subject.Name,
					Address: string(subject.Address),
					Ico:     subject.Ico,
				})
			}

			person := Person{
				BirthDate:       time.Time(rzpPerson.DateOfBirth),
				FirstName:       rzpPerson.FirstName,
				LastName:        rzpPerson.LastName,
				TitleBeforeName: rzpPerson.TitleBeforeName,
				TitleAfterName:  rzpPerson.TitleAfterName,
				FullName:        rzpPerson.DisplayName,
				Subjects:        economicSubjects,
			}
			for _, subject := range subjects.Subjects {
				if subject.Type == "F" {
					subjectDetail, err := client.GetSubjectDetails(subject.Ssarzp)
					if err != nil {
						cancel(err)
						resultChan <- types.Result[Person]{Err: err}
					}
					addressCode, err := rzpAddressToCode(subjectDetail.Address, client, cancel, logger)
					if err != nil {
						logger.Warn("Unable to get address code for subject", slog.String("name", subject.Name), slog.String("address", string(subject.Address)))
					}

					sameAddressSubjects, err := client.SearchSubject(rzp.SearchSubjectQuery{
						AddressCode: addressCode,
					})
					if err != nil {
						logger.Warn("Unable to search subjects with same address", slog.String("address", string(subject.Address)))
					}
					person.SameAddressSubjects = make([]string, 0, len(sameAddressSubjects.Subjects))
					for _, sameAddressSubject := range sameAddressSubjects.Subjects {
						person.SameAddressSubjects = append(person.SameAddressSubjects, sameAddressSubject.Name)
					}

					person.Address = string(subject.Address)
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

	persons := make([]Person, 0, len(rzpPersons))
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

func rzpAddressToCode(address rzp.RzpAddress, client *rzp.Rzp, cancel context.CancelCauseFunc, logger *slog.Logger) (rzp.AddressCode, error) {
	searchable, err := address.ToSearchableString()
	if err != nil {
		err := fmt.Errorf("unable to convert address to searchable string: %v", err)
		return 0, err
	}

	addresses, err := client.SearchAddress(searchable)
	if err != nil {
		err := fmt.Errorf("unable to search address in RZP: %v", err)
		cancel(err)
		return 0, err
	}
	if len(addresses) != 1 {
		return 0, fmt.Errorf("expected exactly one address, got %d", len(addresses))
	}

	return addresses[0].Code, nil
}
