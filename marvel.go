package marvel

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/go-querystring/query"
)

type Client struct {
	PublicKey, PrivateKey string
	Client                *http.Client
}

func (c Client) fetch(path string, params interface{}, out interface{}) error {
	u := c.baseURL(path, params)
	if u.RawQuery != "" {
		u.RawQuery += "&"
	}
	ts, hash := c.hash()
	u.RawQuery += url.Values(map[string][]string{
		"ts":     []string{fmt.Sprintf("%d", ts)},
		"apikey": []string{c.PublicKey},
		"hash":   []string{hash},
	}).Encode()
	if c.Client == nil {
		c.Client = &http.Client{}
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		slurp, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error response from API: %d\n%s", resp.StatusCode, slurp)
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c Client) baseURL(path string, params interface{}) url.URL {
	u := url.URL{
		Scheme: "https",
		Host:   "gateway.marvel.com",
		Path:   "/v1/public/" + path,
	}
	if params != nil {
		q, _ := query.Values(params)
		u.RawQuery += "&" + q.Encode()
	}
	return u
}

// See http://developer.marvel.com/documentation/authorization
func (c Client) hash() (int64, string) {
	ts := time.Now().Unix()
	hash := md5.New()
	io.WriteString(hash, fmt.Sprintf("%d%s%s", ts, c.PrivateKey, c.PublicKey))
	return ts, fmt.Sprintf("%x", hash.Sum(nil))
}

type URL struct {
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Fields common to all request parameter entities
type CommonParams struct {
	OrderBy       string `url:"orderBy,omitempty"`
	Offset        int    `url:"offset,omitempty"`
	Limit         int    `url:"limit,omitempty"`
	ModifiedSince string `url:"modifiedSince,omitempty"`
}

// Fields common to all response entities
type CommonResponse struct {
	Code            int    `json:"code,omitempty"`
	ETag            string `json:"etag,omitempty"`
	Status          string `json:"status,omitempty"`
	Copyright       string `json:"copyright,omitempty"`
	AttributionText string `json:"attributionText,omitempty"`
	AttributionHTML string `json:"attributionHtml,omitempty"`
}

// Fields common to data that lists entities, with pagination
type CommonList struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
	Total  int `json:"total,omitempty"`
	Count  int `json:"count,omitempty"`
}

type Image struct {
	Path      string `json:"path,omitempty"`
	Extension string `json:"extension,omitempty"`
}

type Variant string

var (
	PortraitSmall       = Variant("portrait_small")
	PortraitMedium      = Variant("portrait_medium")
	PortraitXLarge      = Variant("portrait_xlarge")
	PortraitFantastic   = Variant("portrait_fantastic")
	PortraitUncanny     = Variant("portrait_uncanny")
	PortraitIncredible  = Variant("portrait_incredible")
	StandardSmall       = Variant("standard_small")
	StandardMedium      = Variant("standard_medium")
	StandardXLarge      = Variant("standard_xlarge")
	StandardFantastic   = Variant("standard_fantastic")
	StandardUncanny     = Variant("standard_uncanny")
	StandardIncredible  = Variant("standard_incredible")
	LandscapeSmall      = Variant("landscape_small")
	LandscapeMedium     = Variant("landscape_medium")
	LandscapeXLarge     = Variant("landscape_xlarge")
	LandscapeFantastic  = Variant("landscape_fantastic")
	LandscapeUncanny    = Variant("landscape_uncanny")
	LandscapeIncredible = Variant("landscape_incredible")
)

func (i Image) URL(v Variant) string {
	return fmt.Sprintf("%s/%s.%s", i.Path, string(v), i.Extension)
}

type Date string

const dateLayout = "2006-01-02T15:04:05-0700"

func (d Date) Parse() time.Time {
	t, err := time.Parse(dateLayout, string(d))
	if err != nil {
		panic(err)
	}
	return t
}

type ResourceList struct {
	Available     int    `json:"available,omitempty"`
	Returned      int    `json:"returned,omitempty"`
	CollectionURI string `json:"collectionUri,omitempty"`
}

/////
// Characters
/////

func (c Client) Character(id int) CharacterResource {
	return CharacterResource{basePath: fmt.Sprintf("/characters/%d", id), client: c}
}

type CharacterResource struct {
	basePath string
	client   Client
}

func (c Client) Characters(params CharactersParams) (resp *CharactersResponse, err error) {
	err = c.fetch("/characters", params, &resp)
	return
}

func (s CharacterResource) Get() (resp *CharactersResponse, err error) {
	err = s.client.fetch(s.basePath, nil, &resp)
	return
}

func (s CharacterResource) Comics(params ComicsParams) (resp *ComicsResponse, err error) {
	err = s.client.fetch(s.basePath+"/comics", params, &resp)
	return
}

func (s CharacterResource) Events(params EventsParams) (resp *EventsResponse, err error) {
	err = s.client.fetch(s.basePath+"/events", params, &resp)
	return
}

func (s CharacterResource) Series(params SeriesParams) (resp *SeriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

func (s CharacterResource) Stories(params StoriesParams) (resp *StoriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

type CharactersParams struct {
	CommonParams
	Name           string `url:"name,omitempty"`
	NameStartsWith string `url:"nameStartsWith,omitempty"`
	Comics         string `url:"comics,omitempty"`
	Events         string `url:"events,omitempty"`
	Stories        string `url:"stories,omitempty"`
}

type CharactersResponse struct {
	CommonResponse
	Data struct {
		CommonList
		Results []Character `json:"results,omitempty"`
	} `json:"data,omitempty"`
}

type Character struct {
	ResourceURI string       `json:"resourceURI,omitempty"`
	ID          *int         `json:"id,omitempty"`
	Name        *string      `json:"name,omitempty"`
	Description *string      `json:"description,omitempty"`
	Modified    *Date        `json:"modified,omitempty"`
	URLs        []URL        `json:"urls,omitempty"`
	Thumbnail   *Image       `json:"thumbnail,omitempty"`
	Comics      *ComicsList  `json:"comics,omitempty"`
	Stories     *StoriesList `json:"stories,omitempty"`
	Events      *EventsList  `json:"events,omitempty"`
	Series      *SeriesList  `json:"series,omitempty"`
}

type CharactersList struct {
	ResourceList
	Items []Character `json:"items,omitempty"`
}

/////
// Comics
/////

func (c Client) Comic(id int) ComicResource {
	return ComicResource{basePath: fmt.Sprintf("comics/%d", id), client: c}
}

type ComicResource struct {
	basePath string
	client   Client
}

func (c Client) Comics(params ComicsParams) (resp *ComicsResponse, err error) {
	err = c.fetch("/comics", params, &resp)
	return
}

func (s ComicResource) Get() (resp *ComicsResponse, err error) {
	err = s.client.fetch(s.basePath, nil, &resp)
	return
}

func (s ComicResource) Characters(params CharactersParams) (resp *CharactersResponse, err error) {
	err = s.client.fetch(s.basePath+"/characters", params, &resp)
	return
}

func (s ComicResource) Events(params EventsParams) (resp *EventsResponse, err error) {
	err = s.client.fetch(s.basePath+"/events", params, &resp)
	return
}

func (s ComicResource) Series(params SeriesParams) (resp *SeriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

func (s ComicResource) Stories(params StoriesParams) (resp *StoriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

type ComicsParams struct {
	CommonParams
	Format            string `url:"format,omitempty"`
	FormatType        string `url:"formatType,omitempty"`
	NoVariants        bool   `url:"noVariants,omitempty"`
	DateDescriptor    string `url:"dateDescriptor,omitempty"`
	DateRange         string `url:"dateRange,omitempty"`
	DiamondCode       string `url:"diamondCode,omitempty"`
	DigitalID         string `url:"digitalId,omitempty"`
	UPC               string `url:"upc,omitempty"`
	ISBN              string `url:"isbn,omitempty"`
	EAN               string `url:"ean,omitempty"`
	ISSN              string `url:"issn,omitempty"`
	HasDigitalIssue   bool   `url:"hasDigitalIssue,omitempty"`
	Creators          string `url:"creators,omitempty"`
	Characters        string `url:"characters,omitempty"`
	Events            string `url:"events,omitempty"`
	Stories           string `url:"stories,omitempty"`
	SharedAppearances string `url:"sharedAppearances,omitempty"`
	Collaborators     string `url:"collaborators,omitempty"`
}

type ComicsResponse struct {
	CommonResponse
	Data struct {
		CommonList
		Results []Comic `json:"results,omitempty"`
	} `json:"data,omitempty"`
}

type Comic struct {
	ResourceURI        string          `json:"resourceURI,omitempty"`
	ID                 *int            `json:"id,omitempty"`
	Name               *string         `json:"id,omitempty"`
	DigitalID          *int            `json:"digitalId,omitempty"`
	Title              *string         `json:"title,omitempty"`
	IssueNumber        *float64        `json:"issueNumber,omitempty"`
	VariantDescription *string         `json:"variantDescription,omitEmpty"`
	Description        *string         `json:"description,omitempty"`
	Modified           *Date           `json:"modified,omitempty"`
	ISBN               *string         `json:"isbn,omitempty"`
	UPC                *string         `json:"upc,omitempty"`
	DiamondCode        *string         `json:"diamondCode,omitempty"`
	EAN                *string         `json:"ean,omitempty"`
	ISSN               *string         `json:"issn,omitempty"`
	Format             *string         `json:"format,omitempty"`
	PageCount          *int            `json:"pageCount,omitEmpty"`
	TextObjects        []TextObject    `json:"textObjects,omitempty"`
	URLs               []URL           `json:"urls,omitempty"`
	Series             *Series         `json:"series,omitempty"`
	Variants           []Comic         `json:"variants,omitempty"`
	Collections        []Comic         `json:"collections,omitempty"`
	CollectedIssues    []Comic         `json:"collectedIssues,omitempty"`
	Dates              []ComicDate     `json:"dates,omitempty"`
	Prices             []ComicPrice    `json:"prices,omitempty"`
	Thumbnail          *Image          `json:"thumbnail,omitempty"`
	Images             []Image         `json:"images,omitempty"`
	Creators           *CreatorsList   `json:"creators,omitempty"`
	Characters         *CharactersList `json:"characters,omitempty"`
	Stories            *StoriesList    `json:"stories,omitempty"`
	Events             *EventsList     `json:"events,omitempty"`
}

type TextObject struct {
	Type     string `json:"text,omitempty"`
	Language string `json:"language,omitempty"`
	Text     string `json:"text,omitempty"`
}

type ComicDate struct {
	Type string `json:"type,omitempty"`
	Date Date   `json:"date,omitempty"`
}

type ComicPrice struct {
	Type  string  `json:"type,omitempty"`
	Price float64 `json:"price,omitempty"`
}

type ComicsList struct {
	ResourceList
	Items []Comic `json:"items,omitempty"`
}

/////
// Stories
/////

func (c Client) Story(id int) StoryResource {
	return StoryResource{basePath: fmt.Sprintf("stories/%d", id), client: c}
}

type StoryResource struct {
	basePath string
	client   Client
}

func (c Client) Stories(params StoriesParams) (resp *StoriesResponse, err error) {
	err = c.fetch("/stories", params, &resp)
	return
}

func (s StoryResource) Get() (resp *StoriesResponse, err error) {
	err = s.client.fetch(s.basePath, nil, &resp)
	return
}

func (s StoryResource) Characters(params CharactersParams) (resp *CharactersResponse, err error) {
	err = s.client.fetch(s.basePath+"/characters", params, &resp)
	return
}

func (s StoryResource) Comics(params ComicsParams) (resp *ComicsResponse, err error) {
	err = s.client.fetch(s.basePath+"/comics", params, &resp)
	return
}

func (s StoryResource) Creators(params CreatorsParams) (resp *CreatorsResponse, err error) {
	err = s.client.fetch(s.basePath+"/creators", params, &resp)
	return
}

func (s StoryResource) Events(params EventsParams) (resp *EventsResponse, err error) {
	err = s.client.fetch(s.basePath+"/events", params, &resp)
	return
}

func (s StoryResource) Series(params SeriesParams) (resp *SeriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

type StoriesParams struct {
	CommonParams
	Comics     string `url:"comics,omitempty"`
	Events     string `url:"events,omitempty"`
	Creators   string `url:"creators,omitempty"`
	Characters string `url:"characters,omitempty"`
}

type StoriesResponse struct {
	CommonResponse
	Data struct {
		CommonList
		Results []Story `json:"results,omitempty"`
	} `json:"data,omitempty"`
}

type Story struct {
	ResourceURI   *string         `json:"resourceURI,omitempty"`
	ID            *int            `json:"id,omitempty"`
	Name          *string         `json:"name,omitempty"`
	Title         *string         `json:"title,omitempty"`
	Description   *string         `json:"description,omitempty"`
	Type          *string         `json:"type,omitempty"`
	Modified      *Date           `json:"date,omitempty"`
	Thumbnail     *Image          `json:"image,omitempty"`
	Comics        *ComicsList     `json:"comics,omitempty"`
	Series        *SeriesList     `json:"series,omitempty"`
	Events        *EventsList     `json:"events,omitempty"`
	Characters    *CharactersList `json:"characters,omitempty"`
	Creators      *CreatorsList   `json:"creators,omitempty"`
	OriginalIssue Comic
}

type StoriesList struct {
	ResourceList
	Items []Story `json:"items,omitempty"`
}

/////
// Events
/////

func (c Client) Event(id int) EventResource {
	return EventResource{basePath: fmt.Sprintf("events/%d", id), client: c}
}

type EventResource struct {
	basePath string
	client   Client
}

func (c Client) Events(params EventsParams) (resp *EventsResponse, err error) {
	err = c.fetch("/events", params, &resp)
	return
}

func (s EventResource) Get() (resp *EventsResponse, err error) {
	err = s.client.fetch(s.basePath, nil, &resp)
	return
}

func (s EventResource) Characters(params CharactersParams) (resp *CharactersResponse, err error) {
	err = s.client.fetch(s.basePath+"/characters", params, &resp)
	return
}

func (s EventResource) Comics(params ComicsParams) (resp *ComicsResponse, err error) {
	err = s.client.fetch(s.basePath+"/comics", params, &resp)
	return
}

func (s EventResource) Creators(params CreatorsParams) (resp *CreatorsResponse, err error) {
	err = s.client.fetch(s.basePath+"/creators", params, &resp)
	return
}

func (s EventResource) Series(params SeriesParams) (resp *SeriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

func (s EventResource) Stories(params StoriesParams) (resp *StoriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

type EventsParams struct {
	CommonParams
	Name           string `url:"name,omitempty"`
	NameStartsWith string `url:"nameStartsWith,omitempty"`
	Creators       string `url:"creators,omitempty"`
	Characters     string `url:"characters,omitempty"`
	Comics         string `url:"comics,omitempty"`
	Stories        string `url:"stories,omitempty"`
}

type EventsResponse struct {
	CommonResponse
	Data struct {
		CommonList
		Results []Event `json:"results,omitempty"`
	} `json:"data,omitempty"`
}

type Event struct {
	ResourceURI *string         `json:"resourceURI,omitempty"`
	ID          *int            `json:"id,omitempty"`
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	URLs        []URL           `json:"urls,omitempty"`
	Modified    *Date           `json:"modified,omitempty"`
	Start       *Date           `json:"start,omitempty"`
	End         *Date           `json:"end,omitempty"`
	Thumbnail   *Image          `json:"thumbnail,omitempty"`
	Comics      *ComicsList     `json:"comics,omitempty"`
	Stories     *StoriesList    `json:"stories,omitempty"`
	Series      *SeriesList     `json:"series,omitempty"`
	Characters  *CharactersList `json:"characters,omitempty"`
	Creators    *CreatorsList   `json:"creators,omitempty"`
	Next        *Event          `json:"next,omitempty"`
	Previous    *Event          `json:"next,omitempty"`
}

type EventsList struct {
	ResourceList
	Items []Event `json:"items,omitempty"`
}

/////
// Series
/////

// Named SingleSeries because Series (plural) will search for series'
func (c Client) SingleSeries(id int) SeriesResource {
	return SeriesResource{basePath: fmt.Sprintf("/series/%d", id), client: c}
}

type SeriesResource struct {
	basePath string
	client   Client
}

func (c Client) Series(params SeriesParams) (resp *SeriesResponse, err error) {
	err = c.fetch("/series", params, &resp)
	return
}

func (s SeriesResource) Get() (resp *SeriesResponse, err error) {
	err = s.client.fetch(s.basePath, nil, &resp)
	return
}

func (s SeriesResource) Characters(params CharactersParams) (resp *CharactersResponse, err error) {
	err = s.client.fetch(s.basePath+"/characters", params, resp)
	return
}

func (s SeriesResource) Comics(params ComicsParams) (resp *ComicsResponse, err error) {
	err = s.client.fetch(s.basePath+"/comics", params, &resp)
	return
}

func (s SeriesResource) Creators(params CreatorsParams) (resp *CreatorsResponse, err error) {
	err = s.client.fetch(s.basePath+"/creators", params, &resp)
	return
}

func (s SeriesResource) Events(params EventsParams) (resp *EventsResponse, err error) {
	err = s.client.fetch(s.basePath+"/events", params, &resp)
	return
}

func (s SeriesResource) Stories(params StoriesParams) (resp *StoriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

type SeriesParams struct {
	CommonParams
	Events          string `url:"events,omitempty"`
	Title           string `url:"title,omitempty"`
	TitleStartsWith string `url:"titleStartsWith,omitempty"`
	StartYear       string `url:"startYear,omitempty"`
	SeriesType      string `url:"seriesType,omitempty"`
	Contains        string `url:"contains,omitempty"`
	Comics          string `url:"comics,omitempty"`
	Creators        string `url:"creators,omitempty"`
	Characters      string `url:"characters,omitempty"`
}

type SeriesResponse struct {
	CommonResponse
	Data struct {
		CommonList
		Results []Series `json:"results,omitempty"`
	} `json:"data,omitempty"`
}

type Series struct {
	ResourceURI *string         `json:"resourceURI,omitempty"`
	ID          *int            `json:"id,omitempty"`
	Name        *string         `json:"name,omitempty"`
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	URLs        []URL           `json:"urls,omitempty"`
	StartYear   *int            `json:"startYear,omitempty"`
	EndYear     *int            `json:"endYear,omitempty"`
	Rating      *string         `json:"rating,omitempty"`
	Modified    *Date           `json:"modified,omitempty"`
	Thumbnail   *Image          `json:"thumbnail,omitempty"`
	Comics      *ComicsList     `json:"comics,omitempty"`
	Stories     *StoriesList    `json:"stories,omitempty"`
	Events      *EventsList     `json:"events,omitempty"`
	Characters  *CharactersList `json:"characters,omitempty"`
	Creators    *CreatorsList   `json:"creators,omitempty"`
	Next        *Series         `json:"next,omitempty"`
	Previous    *Series         `json:"next,omitempty"`
}

type SeriesList struct {
	ResourceList
	Items []Series
}

/////
// Creators
/////

func (c Client) Creator(id int) CreatorResource {
	return CreatorResource{basePath: fmt.Sprintf("/creators/%d", id), client: c}
}

type CreatorResource struct {
	basePath string
	client   Client
}

func (c Client) Creators(params CreatorsParams) (resp *CreatorsResponse, err error) {
	err = c.fetch("/creators", params, &resp)
	return
}

func (s CreatorResource) Get() (resp *CreatorsResponse, err error) {
	err = s.client.fetch(s.basePath, nil, &resp)
	return
}

func (s CreatorResource) Comics(params ComicsParams) (resp *ComicsResponse, err error) {
	err = s.client.fetch(s.basePath+"/comics", params, &resp)
	return
}

func (s CreatorResource) Events(params EventsParams) (resp *EventsResponse, err error) {
	err = s.client.fetch(s.basePath+"/events", params, &resp)
	return
}

func (s CreatorResource) Series(params SeriesParams) (resp *SeriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/series", params, &resp)
	return
}

func (s CreatorResource) Stories(params StoriesParams) (resp *StoriesResponse, err error) {
	err = s.client.fetch(s.basePath+"/stories", params, &resp)
	return
}

type CreatorsParams struct {
	CommonParams
	FirstName            string `url:"firstName,omitempty"`
	MiddleName           string `url:"middleName,omitempty"`
	LastName             string `url:"lastName,omitempty"`
	Suffix               string `url:"suffix,omitempty"`
	NameStartsWith       string `url:"nameStartsWith,omitempty"`
	FirstNameStartsWith  string `url:"firstNameStartsWith,omitempty"`
	MiddleNameStartsWith string `url:"middleNameStartsWith,omitempty"`
	LastNameStartsWith   string `url:"lastNameStartsWith,omitempty"`
	Comics               string `url:"comics,omitempty"`
	Events               string `url:"events,omitempty"`
	Stories              string `url:"stories,omitempty"`
}

type CreatorsResponse struct {
	CommonResponse
	Data struct {
		CommonList
		Results []Creator `json:"results,omitempty"`
	} `json:"data,omitempty"`
}

type Creator struct {
	ResourceURI *string      `json:"resourceURI,omitempty"`
	ID          *int         `json:"id,omitempty"`
	Name        *string      `json:"name,omitempty"`
	FirstName   *string      `json:"firstName,omitempty"`
	MiddleName  *string      `json:"middleName,omitempty"`
	LastName    *string      `json:"lastName,omitempty"`
	Suffix      *string      `json:"suffix,omitempty"`
	FullName    *string      `json:"fullName,omitempty"`
	Modified    *Date        `json:"modified,omitempty"`
	URLs        []URL        `json:"urls,omitempty"`
	Thumbnail   *Image       `json:"thumbnail,omitempty"`
	Series      *SeriesList  `json:"series,omitempty"`
	Stories     *StoriesList `json:"stories,omitempty"`
	Comics      *ComicsList  `json:"comics,omitempty"`
	Events      *EventsList  `json:"events,omitempty"`
}

type CreatorsList struct {
	ResourceList
	Items []Creator
}
