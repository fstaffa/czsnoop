package rzp

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"testing"
	"time"
)

var rzp *Rzp

func TestMain(m *testing.M) {
	var err error
	rzp, err = CreateClient(context.Background(), slog.Default())
	if err != nil {
		fmt.Printf("Unable to create client %v", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func Test_CreateClient(t *testing.T) {
	t.Parallel()
	rzp, err := CreateClient(context.Background(), slog.Default())
	if err != nil {
		t.Errorf("Received unexpected error %v", err)
	}
	if rzp == nil {
		t.Fatalf("Expected rzp to be non-nil")
	}
	if rzp.sessionId == "" {
		t.Errorf("Expected session id to be non-zero")
	}
	u, _ := url.Parse(baseUrl)
	cookies := rzp.client.Jar.Cookies(u)
	if len(cookies) != 1 {
		t.Errorf("Expected exactly one cookie")
	}

	if cookies[0].Name != "mseidf" || cookies[0].Value == "" {
		t.Errorf("Expected cookie with name to be mseidf to be set")
	}
}

func Test_SearchSubject_ByName_Count(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		name                        string
		query                       SearchSubjectQuery
		resultLength                int
		resultHasMorePosibleMatches bool
	}{
		"too many matches": {query: SearchSubjectQuery{Name: "novak", SubjectType: "enterpreneur"}, resultLength: 50, resultHasMorePosibleMatches: true},
		"no matches":       {query: SearchSubjectQuery{Name: "asdfeeija", SubjectType: "enterpreneur"}, resultLength: 0, resultHasMorePosibleMatches: false},
		"few matches":      {query: SearchSubjectQuery{Name: "02930366", SubjectType: "enterpreneur"}, resultLength: 0, resultHasMorePosibleMatches: false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := rzp.SearchSubject(test.query)
			if err != nil {
				t.Fatalf("Received unexpected error %v", err)
			}
			if len(res.Subjects) != test.resultLength {
				t.Errorf("Expected %d subjects, got %d", test.resultLength, len(res.Subjects))
			}

			if res.MorePossibleMatches != test.resultHasMorePosibleMatches {
				t.Errorf("Expected more possible matches to be %v, got %v", test.resultHasMorePosibleMatches, res.MorePossibleMatches)
			}
		})

	}
}

func Test_SearchSubject_DetailedResult(t *testing.T) {
	t.Parallel()

	res, err := rzp.SearchSubject(SearchSubjectQuery{Ico: "01895541", SubjectType: "enterpreneur"})

	if err != nil {
		t.Fatalf("Unable to search subject %v", err)
	}

	if len(res.Subjects) != 1 {
		t.Fatalf("Expected exactly one subject, got %d", len(res.Subjects))
	}

	subject := res.Subjects[0]

	if subject.Name != "THOMAS SILVERTONNI s.r.o." {
		t.Errorf("Expected name to be THOMAS SILVERTONNI s.r.o., got %s", subject.Name)
	}
	if subject.Ico != "01895541" {
		t.Errorf("Expected ICO to be 01895541, got %s", subject.Ico)
	}

	if subject.Address != "Mazovská 479/8, 181 00, Praha 8 - Troja" {
		t.Errorf("Expected address to be Mazovská 479/8, 181 00, Praha 8 - Troja, got %s", subject.Address)
	}
	if subject.Ssarzp == "" {
		t.Errorf("Expected ssarzp to be non-empty")
	}
}

func Test_GetSubjectDetails(t *testing.T) {
	t.Parallel()

	result, err := rzp.SearchSubject(SearchSubjectQuery{Name: "ing phd novak", SubjectType: "enterpreneur"})
	if err != nil {
		t.Fatalf("Unable to search subject %v", err)
	}

	subject := result.Subjects[0]

	detail, err := rzp.GetSubjectDetails(subject.Ssarzp)
	if err != nil {
		t.Fatalf("Received unexpected error %v", err)
	}

	if detail.Ico != subject.Ico {
		t.Errorf("Expected ICO to be %s, got %s", subject.Ico, detail.Ico)
	}
	if detail.FullNameWithTitles != subject.Name {
		t.Errorf("Expected name to be %s, got %s", subject.Name, detail.FullNameWithTitles)
	}

	if detail.Address != subject.Address {
		t.Errorf("Expected address to be %s, got %s", subject.Address, detail.Address)
	}

	if detail.FirstName == "" {
		t.Errorf("Expected first name to be non-empty")
	}

	if detail.LastName == "" {
		t.Errorf("Expected last name to be non-empty")
	}

	if detail.BirthDate.IsZero() {
		t.Errorf("Expected birth date to be non-zero")
	}

	if detail.TitleAfterName != "Ph.D." {
		t.Errorf("Expected title after name to be Ph.D., got %s", detail.TitleAfterName)
	}

	if detail.TitleBeforeName != "Ing." {
		t.Errorf("Expected title before name to be 'Ing.', got '%s'", detail.TitleBeforeName)
	}

	if detail.Citizenship != "Česká republika" {
		t.Errorf("Expected citizenship to be 'Česká republika', got '%s'", detail.Citizenship)
	}
	if len(detail.Trades) == 0 {
		t.Errorf("Expected trades to be non-empty")
	}
	for _, trade := range detail.Trades {
		if trade.TradeType == "" {
			t.Errorf("Expected trade to be non-empty")
		}
		if trade.DateOfOrigin.IsZero() {
			t.Errorf("Expected valid from to be non-zero")
		}
	}
}

func Test_SearchPerson_ByName(t *testing.T) {
	t.Parallel()

	res, err := rzp.SearchPerson(SearchPersonQuery{Surname: "novak"})

	if err != nil {
		t.Fatalf("Unable to search person %v", err)
	}

	if len(res.People) == 0 {
		t.Fatalf("Expected at least one person")
	}

	for _, person := range res.People {
		if person.FirstName == "" {
			t.Errorf("Expected first name to be non-empty")
		}
		if person.LastName == "" {
			t.Errorf("Expected last name to be non-empty")
		}
		if time.Time(person.DateOfBirth).IsZero() {
			t.Errorf("Expected birth date to be non-zero")
		}
		if person.PersonId == "" {
			t.Errorf("Expected person id to be non-empty")
		}
		if person.DisplayName == "" {
			t.Errorf("Expected full name to be non-empty")
		}
	}
}

func Test_SearchPerson_ParseTitle(t *testing.T) {
	t.Parallel()

	initialSearch, err := rzp.SearchSubject(SearchSubjectQuery{Name: "novak csc", SubjectType: "enterpreneur"})
	if err != nil {
		t.Fatalf("Unable to search subject %v", err)
	}

	if len(initialSearch.Subjects) == 0 {
		t.Fatalf("Expected at least one subject")
	}

	detailedSubject, err := rzp.GetSubjectDetails(initialSearch.Subjects[0].Ssarzp)
	if err != nil {
		t.Fatalf("Unable to get subject details: %v", err)
	}

	res, err := rzp.SearchPerson(SearchPersonQuery{Surname: detailedSubject.LastName, FirstName: detailedSubject.FirstName, DateOfBirth: detailedSubject.BirthDate})
	if err != nil {
		t.Fatalf("Unable to search person %v", err)
	}
	if len(res.People) != 1 {
		t.Fatalf("Expected to match exactly one person, found %d", len(res.People))
	}

	person := res.People[0]
	if person.TitleBeforeName != detailedSubject.TitleBeforeName {
		t.Errorf("Expected title before name to be %s, got %s", detailedSubject.TitleBeforeName, person.TitleBeforeName)
	}
	if person.TitleAfterName != detailedSubject.TitleAfterName {
		t.Errorf("Expected title after name to be %s, got %s", detailedSubject.TitleAfterName, person.TitleAfterName)
	}
}

func Test_SearchSubject_ByPersonId(t *testing.T) {
	t.Parallel()

	res, err := rzp.SearchPerson(SearchPersonQuery{Surname: "novak"})
	if err != nil {
		t.Fatalf("Unable to search person %v", err)
	}
	if len(res.People) == 0 {
		t.Fatalf("Expected at least one person")
	}
	personId := res.People[0].PersonId

	subjects, err := rzp.SearchSubject(SearchSubjectQuery{PersonId: personId})
	if err != nil {
		t.Fatalf("Unable to search subject %v", err)
	}
	if len(subjects.Subjects) == 0 {
		t.Fatalf("Expected at least one subject")
	}
}

func Test_SearchAddress_SingleResult(t *testing.T) {
	t.Parallel()

	res, err := rzp.SearchAddress("milovicka 9")
	if err != nil {
		t.Fatalf("Unable to search address %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("Expected exactly one address, got %d", len(res))
	}

	address := res[0]
	if address.Address != "Milovická 197/9, 198 00, Praha 9 - Hloubětín" {
		t.Errorf("Expected address to be Milovická 197/9, 198 00, Praha 9 - Hloubětín, got %s", address.Address)
	}
	if address.Code != AddressCode(22421611) {
		t.Errorf("Expected code to be 22421611, got %d", address.Code)
	}
}

func Test_SearchAddress_ManyResults(t *testing.T) {
	t.Parallel()

	res, err := rzp.SearchAddress("zborovska 4")
	if err != nil {
		t.Fatalf("Unable to search address %v", err)
	}
	if len(res) < 11 {
		t.Fatalf("Expected at least 11 results, got %d", len(res))
	}
}

func Test_SearchAddress_SearchBySubjectAddress(t *testing.T) {
	t.Parallel()

	subjects, err := rzp.SearchSubject(SearchSubjectQuery{Name: "novak", SubjectType: "enterpreneur"})
	if err != nil {
		t.Fatalf("Unable to search subject %v", err)
	}
	if len(subjects.Subjects) == 0 {
		t.Fatalf("Expected at least one subject")
	}

	subject := subjects.Subjects[0]
	address, err := subject.Address.ToSearchableString()
	if err != nil {
		t.Fatalf("Unable to convert address '%s' to searchable string %v", subject.Address, err)
	}
	res, err := rzp.SearchAddress(address)
	if err != nil {
		t.Fatalf("Unable to search address %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("Expected at least one address when searching by subject address %s", subject.Address)
	}
	if res[0].Address != string(subject.Address) {
		t.Errorf("Expected address to be %s, got %s", subject.Address, res[0].Address)
	}
}
