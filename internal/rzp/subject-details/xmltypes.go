package subjectdetails

import "encoding/xml"

type Vypis struct {
	XMLName    xml.Name `xml:"Vypis"`
	Nadpis     string   `xml:"Nadpis"`
	Oduvodneni string   `xml:"Oduvodneni"`
	Subjekt    subjekt  `xml:"Subjekt"`
	InfoText   string   `xml:"InfoText"`
	Vydano     string   `xml:"Vydano"`
}

type subjekt struct {
	Podnikatel      podnikatel     `xml:"Podnikatel"`
	ObchodniJmenoFO string         `xml:"ObchodniJmenoFO"`
	Sidlo           sidlo          `xml:"Sidlo"`
	Ico             ico            `xml:"Ico"`
	ZivnostiSeznam  zivnostiSeznam `xml:"ZivnostiSeznam"`
	EvidujiciUrad   string         `xml:"EvidujiciUrad"`
	Odkazy          odkazy         `xml:"Odkazy"`
}

type podnikatel struct {
	Osoba osoba `xml:"Osoba"`
}

type osoba struct {
	OsobaPoradi           string                `xml:"osobaPoradi,attr"`
	JmenoPrijmeniPlatnost jmenoPrijmeniPlatnost `xml:"JmenoPrijmeniPlatnost"`
	DatumNarozeniPlatnost datumNarozeniPlatnost `xml:"DatumNarozeniPlatnost"`
	ObcanstviPlatnost     obcanstviPlatnost     `xml:"ObcanstviPlatnost"`
}

type jmenoPrijmeniPlatnost struct {
	Descr   string `xml:"descr,attr"`
	Hodnota string `xml:"Hodnota"`
}

type datumNarozeniPlatnost struct {
	Descr   string `xml:"descr,attr"`
	Hodnota string `xml:"Hodnota"`
}

type obcanstviPlatnost struct {
	Descr   string `xml:"descr,attr"`
	Hodnota string `xml:"Hodnota"`
}

type sidlo struct {
	Descr  string `xml:"descr,attr"`
	Adresa adresa `xml:"Adresa"`
}

type adresa struct {
	Hodnota string `xml:"Hodnota"`
}

type ico struct {
	Descr   string `xml:"descr,attr"`
	Hodnota string `xml:"Hodnota"`
}

type zivnostiSeznam struct {
	Zivnost []zivnost `xml:"Zivnost"`
}

type zivnost struct {
	Descr            string           `xml:"descr,attr"`
	ZivnostPoradi    string           `xml:"zivnostPoradi,attr"`
	Predmet          predmet          `xml:"Predmet"`
	Druh             druh             `xml:"Druh"`
	DatumVzniku      string           `xml:"DatumVzniku"`
	PlatnostZivnosti platnostZivnosti `xml:"PlatnostZivnosti"`
}

type predmet struct {
	Descr              string             `xml:"descr,attr"`
	HodnotaPredmetDruh hodnotaPredmetDruh `xml:"HodnotaPredmetDruh"`
}

type hodnotaPredmetDruh struct {
	Hodnota string `xml:"Hodnota"`
}

type druh struct {
	Descr              string             `xml:"descr,attr"`
	HodnotaPredmetDruh hodnotaPredmetDruh `xml:"HodnotaPredmetDruh"`
}

type platnostZivnosti struct {
	Descr   string `xml:"descr,attr"`
	Hodnota string `xml:"Hodnota"`
}

type odkazy struct {
	VypisPDF string `xml:"VypisPDF"`
	VypisXML string `xml:"VypisXML"`
}
