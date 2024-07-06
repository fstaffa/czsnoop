package rzp

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/fstaffa/czsnoop/internal/rzp/statement"
	"github.com/fstaffa/czsnoop/internal/rzp/subject-details"
	"github.com/fstaffa/czsnoop/internal/types"
	"golang.org/x/text/encoding/ianaindex"
)

const baseUrl = "https://www.rzp.cz"
const dateFormat = "02.01.2006"

type Rzp struct {
	sessionId string
	client    http.Client
	logger    *slog.Logger
	context   context.Context
}

// This field seem to be bound to session
type Ssarzp string

func CreateClient(context context.Context) (*Rzp, error) {
	logger := slog.Default()
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create cookie jar for client: %v", err)
	}

	client := http.Client{Jar: jar, Timeout: 60 * time.Second}
	sessionId, err := getSessionId(&client, context)
	if err != nil {
		return nil, fmt.Errorf("unable to get session id: %v", err)
	}
	logger.DebugContext(context, "Created RZP client", slog.String("rzpSessionId", sessionId))
	return &Rzp{
		sessionId: sessionId,
		client:    client,
		logger:    logger,
		context:   context,
	}, nil
}

type sessionResponse struct {
	SessionId string `json:"sesid"`
}

func getSessionId(c *http.Client, context context.Context) (string, error) {
	req, err := http.NewRequestWithContext(context, http.MethodGet, "https://www.rzp.cz/rzp/api-c/srv/session/v1/start", nil)

	if err != nil {
		return "", fmt.Errorf("unable to create request: %v", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d and status %s", resp.StatusCode, resp.Status)
	}

	var sessionResponse sessionResponse
	err = json.NewDecoder(resp.Body).Decode(&sessionResponse)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal response: %v", err)
	}
	return sessionResponse.SessionId, nil
}

type Subject struct {
	Name    string    `json:"nazev"`
	Ico     types.Ico `json:"ico"`
	Address string    `json:"sidlo"`
	Ssarzp  Ssarzp    `json:"ssarzp"`
}

type SearchSubjectResponse struct {
	MorePossibleMatches bool      `json:"seznamNeniKompletni"`
	Subjects            []Subject `json:"subjekty"`
}

type SearchSubjectQuery struct {
	Name           string
	StartsWithName bool
	Ico            types.Ico
	// either 'enterpreneur' or 'statutory body'
	SubjectType string
}

func (r *Rzp) SearchSubject(query SearchSubjectQuery) (SearchSubjectResponse, error) {
	req, err := http.NewRequestWithContext(r.context, http.MethodGet, "https://www.rzp.cz/rzp/api3-c/srv/vw/v1/subjekty", nil)
	if err != nil {
		return SearchSubjectResponse{}, fmt.Errorf("unable to create request: %v", err)
	}
	req.Header.Set("Sesid", r.sessionId)
	req.Header.Set("Accept-Language", "cs")
	q := req.URL.Query()
	if query.Name != "" {
		q.Add("s-obchjm", query.Name)
	}

	q.Add("s-ico", string(query.Ico))
	// without true, it throws error that last part of word has to be at least 4 characters
	q.Add("s-presvyber", "true")
	q.Add("pouzeplatne", "true")
	if query.SubjectType == "statutory body" {
		q.Add("s-role", "S")
	} else if query.SubjectType == "enterpreneur" {
		q.Add("s-role", "P")
	} else {
		return SearchSubjectResponse{}, fmt.Errorf("unknown subject type: %s", query.SubjectType)
	}

	req.URL.RawQuery = q.Encode()
	r.logger.DebugContext(r.context, "Searching for subject", slog.String("url", req.URL.String()))
	resp, err := r.client.Do(req)
	if err != nil {
		return SearchSubjectResponse{}, fmt.Errorf("unable to do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		contentB, err := io.ReadAll(resp.Body)
		var contentS string
		if err != nil {
			contentS = "Unable to read error response string"
		}
		contentS = string(contentB)

		return SearchSubjectResponse{}, fmt.Errorf("unexpected status code: %d and status %s, with response %s", resp.StatusCode, resp.Status, contentS)
	}

	var searchResult SearchSubjectResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {

		return SearchSubjectResponse{}, fmt.Errorf("unable to unmarshal response: %v", err)
	}
	r.logger.DebugContext(r.context, "Search result", slog.Any("result", searchResult))

	return searchResult, nil
}

type SubjectDetail struct {
	Address            string
	Ico                types.Ico
	FullNameWithTitles string
	Trades             []Trade
	FirstName          string
	LastName           string
	TitleBeforeName    string
	TitleAfterName     string
	BirthDate          time.Time
	Citizenship        string
}

type Trade struct {
	TradeType         string
	DateOfOrigin      time.Time
	ValidityOfLicense string
}

func (r *Rzp) GetSubjectDetails(ssarzp Ssarzp) (SubjectDetail, error) {
	req, err := http.NewRequestWithContext(r.context, http.MethodGet, fmt.Sprintf("%s%s%s%s", baseUrl, `/rzp/api3-c/srv/vw/v1/subjekty/isvs/`, ssarzp, ".xml"), nil)
	if err != nil {
		return SubjectDetail{}, fmt.Errorf("unable to create request: %v", err)
	}
	req.Header.Set("Accept", "text/xml")
	req.Header.Set("Sesid", r.sessionId)
	req.Header.Set("Accept-Language", "cs")
	resp, err := r.client.Do(req)
	if err != nil {
		return SubjectDetail{}, fmt.Errorf("unable to do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return SubjectDetail{}, fmt.Errorf("unexpected status code: %d and status %s", resp.StatusCode, resp.Status)
	}

	var v subjectdetails.Vypis
	err = xml.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return SubjectDetail{}, fmt.Errorf("unable to unmarshal response: %v", err)
	}

	trades := make([]Trade, 0, len(v.Subjekt.ZivnostiSeznam.Zivnost))
	for _, z := range v.Subjekt.ZivnostiSeznam.Zivnost {
		date, err := time.Parse(dateFormat, z.DatumVzniku)
		if err != nil {
			return SubjectDetail{}, fmt.Errorf("unable to parse date: %v", err)
		}
		trades = append(trades, Trade{
			TradeType:         z.Predmet.HodnotaPredmetDruh.Hodnota,
			DateOfOrigin:      date,
			ValidityOfLicense: z.PlatnostZivnosti.Hodnota,
		})
	}

	deeperDetails, err := r.getSubjectStatement(v.Subjekt.Odkazy.VypisXML)
	if err != nil {
		return SubjectDetail{}, fmt.Errorf("unable to get deeper subject details: %v", err)
	}

	return deeperDetails, nil
}

func (r *Rzp) getSubjectStatement(path string) (SubjectDetail, error) {
	req, err := http.NewRequestWithContext(r.context, http.MethodGet, fmt.Sprintf("%s%s", baseUrl, path), nil)
	if err != nil {
		return SubjectDetail{}, fmt.Errorf("unable to create deeper subject details request: %v", err)
	}
	req.Header.Set("Sesid", r.sessionId)
	resp, err := r.client.Do(req)
	if err != nil {
		return SubjectDetail{}, fmt.Errorf("unable to do deeper subject details request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SubjectDetail{}, fmt.Errorf("unexpected status code: %d and status %s", resp.StatusCode, resp.Status)
	}
	var l statement.Listiny
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = func(charset string, reader io.Reader) (io.Reader, error) {
		enc, err := ianaindex.IANA.Encoding(charset)
		if err != nil {
			return nil, fmt.Errorf("charset %s: %s", charset, err.Error())
		}
		return enc.NewDecoder().Reader(reader), nil
	}
	err = decoder.Decode(&l)
	if err != nil {
		return SubjectDetail{}, fmt.Errorf("unable to unmarshal deeper subject details response: %v", err)
	}

	enterpreneuerDetail := l.Verweb.PodnikatelDetail
	birthDate, err := time.Parse("02.01.2006", enterpreneuerDetail.PodnikatelOsoba.ZucastnenaOsobaDetail.DatumNarozeni.Hodnota)
	if err != nil {
		return SubjectDetail{}, fmt.Errorf("unable to parse birth date in statement detail: %v", err)
	}

	trades := make([]Trade, 0, len(enterpreneuerDetail.SeznamZivnosti.Zivnost[0].Obor.Vycet.Drive))
	for _, zivnost := range enterpreneuerDetail.SeznamZivnosti.Zivnost {
		date, err := time.Parse(dateFormat, zivnost.Vznik)
		if err != nil {
			return SubjectDetail{}, fmt.Errorf("unable to parse date: %v", err)
		}
		for _, trade := range zivnost.Obor.Vycet.Drive {
			trades = append(trades, Trade{
				TradeType:    trade.Hodnota,
				DateOfOrigin: date,
			})
		}
	}
	return SubjectDetail{
		Address:            enterpreneuerDetail.AdresaPodnikani.PlatnostAdresy.ZmenaAdresy.TextAdresy,
		Ico:                types.Ico(enterpreneuerDetail.IdentifikacniCislo.PlatnostHodnoty.Hodnota),
		FullNameWithTitles: enterpreneuerDetail.PodnikatelOsoba.ZucastnenaOsobaDetail.JmenoPrijmeni.Hodnota,
		FirstName:          enterpreneuerDetail.PodnikatelOsoba.ZucastnenaOsobaDetail.Jmeno.Hodnota,
		LastName:           enterpreneuerDetail.PodnikatelOsoba.ZucastnenaOsobaDetail.Prijmeni.Hodnota,
		TitleBeforeName:    enterpreneuerDetail.PodnikatelOsoba.ZucastnenaOsobaDetail.TitulPredJmenem.Hodnota,
		TitleAfterName:     enterpreneuerDetail.PodnikatelOsoba.ZucastnenaOsobaDetail.TitulZaJmenem.Hodnota,
		BirthDate:          birthDate,
		Citizenship:        enterpreneuerDetail.PodnikatelOsoba.ZucastnenaOsobaDetail.Obcanstvi.Hodnota,
		Trades:             trades,
	}, nil
}

type SearchPersonQuery struct {
	FirstName   string
	Surname     string
	DateOfBirth time.Time
}

type SearchPersonResponse struct {
	MorePossibleMatches bool     `json:"seznamNeniKompletni"`
	People              []Person `json:"osoby"`
}

type Person struct {
	FirstName       string      `json:"jmeno"`
	LastName        string      `json:"prijmeni"`
	DisplayName     string      `json:"zobrazeneJmeno"`
	TitleBeforeName string      `json:"titulPred"`
	TitleAfterName  string      `json:"titulZa"`
	DateOfBirth     Iso8601Date `json:"datum"`
	PersonId        string      `json:"idOsoby"`

	//not sure what this is
	PersonRole string `json:"roleOsoby"`
}

type Iso8601Date time.Time

func (d *Iso8601Date) UnmarshalJSON(data []byte) error {
	var date string
	err := json.Unmarshal(data, &date)
	if err != nil {
		return err
	}
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}
	*d = Iso8601Date(parsed)
	return nil
}

func (r *Rzp) SearchPerson(query SearchPersonQuery) (SearchPersonResponse, error) {
	req, err := http.NewRequestWithContext(r.context, http.MethodGet, "https://www.rzp.cz/rzp/api3-c/srv/vw/v1/osoby", nil)
	if err != nil {
		return SearchPersonResponse{}, fmt.Errorf("unable to create request: %v", err)
	}
	req.Header.Set("Sesid", r.sessionId)
	req.Header.Set("Accept-Language", "cs")
	q := req.URL.Query()
	q.Add("pouzeplatne", "true")
	if query.FirstName != "" {
		q.Add("o-jmeno", query.FirstName)
	}
	if query.Surname != "" {
		q.Add("o-prijmeni", query.Surname)
	}
	if !query.DateOfBirth.IsZero() {
		q.Add("o-datum", query.DateOfBirth.Format(time.DateOnly))
	}
	req.URL.RawQuery = q.Encode()
	resp, err := r.client.Do(req)
	if err != nil {
		return SearchPersonResponse{}, fmt.Errorf("unable to do search person request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return SearchPersonResponse{}, fmt.Errorf("unexpected status code: %d and status %s", resp.StatusCode, resp.Status)
	}

	var searchResult SearchPersonResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResult)
	if err != nil {
		return SearchPersonResponse{}, fmt.Errorf("unable to unmarshal response: %v", err)
	}

	return searchResult, nil
}
