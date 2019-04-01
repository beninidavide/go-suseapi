package bugzilla

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/browser"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

type RequestError struct{ error }

func (e RequestError) Error() string {
	return fmt.Sprintf("cannot build request: %v", e.error)
}

type ConnectionError struct{ error }

func (e ConnectionError) Error() string {
	return fmt.Sprintf("cannot communicate with server: %v", e.error)
}

type Config struct {
	BaseURL  string
	User     string
	Password string
}

type Client struct {
	Config        Config
	browser       *browser.Browser
	seriousClient *http.Client
	showBugUrl    string
}

func getAuth(config *Config) string {
	pair := []byte(config.User + ":" + config.Password)
	auth := "Basic " + base64.StdEncoding.EncodeToString(pair)
	return auth
}

func getBrowser(config *Config) *browser.Browser {
	browser := surf.NewBrowser()
	browser.AddRequestHeader("Authorization", getAuth(config))

	return browser
}

func getDecentHTTPClient(config *Config) *http.Client {
	tr := http.DefaultClient.Transport
	client := http.Client{Transport: tr}
	rt := WithHeader(tr)
	rt.Set("Authorization", getAuth(config))
	client.Transport = rt
	return &client
}

func New(config Config) (*Client, error) {
	browser := getBrowser(&config)
	seriousClient := getDecentHTTPClient(&config)
	client := &Client{Config: config, browser: browser, seriousClient: seriousClient}
	return client, nil
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func WithHeader(rt http.RoundTripper) withHeader {
	if rt == nil {
		rt = http.DefaultTransport
	}

	return withHeader{Header: make(http.Header), rt: rt}
}

func (h withHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}

	return h.rt.RoundTrip(req)
}

func (c *Client) getShowBugURL(id int, values map[string]string) (string, error) {
	url, err := url.Parse(c.Config.BaseURL)
	if err != nil {
		return "", RequestError{err}
	}

	url.Path = path.Join(url.Path, "show_bug.cgi")

	query := url.Query()
	query.Set("id", fmt.Sprintf("%d", id))
	for k, v := range values {
		query.Set(k, v)
	}
	url.RawQuery = query.Encode()

	return url.String(), nil
}

type User struct {
	Name  string `xml:"name,attr"`
	Email string `xml:",chardata"`
}

type Group struct {
	Id   int    `xml:"id,attr"`
	Name string `xml:",chardata"`
}

type Flag struct {
	Name      string `xml:"name,attr"`
	Id        int    `xml:"id,attr"`
	TypeId    int    `xml:"type_id,attr"`
	Status    string `xml:"status,attr"`
	Setter    string `xml:"setter,attr"`
	Requestee string `xml:"requestee,attr"`
}

type Bug struct {
	Reporter   User `xml:"reporter"`
	AssignedTo User `xml:"assigned_to"`
	QAContact  User `xml:"qa_contact"`

	Groups []Group `xml:"group"`

	BugId              int       `xml:"bug_id"` // 1047068
	CreationTS         time.Time // 2017-07-03 13:29:00 +0000
	ShortDesc          string    `xml:"short_desc"` // L4: test cloud bug
	DeltaTS            time.Time // 2019-03-27 10:45:20 +0000
	ReporterAccessible int       `xml:"reporter_accessible"` // 0
	CCListAccessible   int       `xml:"cclist_accessible"`   // 0
	ClassificationID   int       `xml:"classification_id"`   // 111
	Classification     string    `xml:"classification"`      // foobar Frobnicator Cloud
	Product            string    `xml:"product"`             // foobar Frobnicator Cloud 7
	Component          string    `xml:"component"`           // Frobtool
	Version            string    `xml:"version"`             // Milestone 8
	RepPlatform        string    `xml:"rep_platform"`        // Other
	OpSys              string    `xml:"op_sys"`              // Other
	BugStatus          string    `xml:"bug_status"`          // RESOLVED
	Resolution         string    `xml:"resolution"`          // FIXED
	DupId              int       `xml:"dup_id"`

	BugFileLoc       string `xml:"bug_file_loc"`      //
	StatusWhiteboard string `xml:"status_whiteboard"` // wasL3:48626  zzz
	Keywords         string `xml:"keywords"`          // DSLA_REQUIRED, DSLA_SOLUTION_PROVIDED
	Priority         string `xml:"priority"`          // P5 - None
	BugSeverity      string `xml:"bug_severity"`      // Normal
	TargetMilestone  string `xml:"target_milestone"`  // ---

	EverConfirmed int      `xml:"everconfirmed"`  // 1
	Cc            []string `xml:"cc"`             // user@foobar.com
	EstimatedTime string   `xml:"estimated_time"` // 0.00
	RemainingTime string   `xml:"remaining_time"` // 0.00
	ActualTime    string   `xml:"actual_time"`    // 0.00

	CfFoundby       []string `xml:"cf_foundby"`       // ---
	CfNts_priority  []string `xml:"cf_nts_priority"`  //
	CfBiz_priority  []string `xml:"cf_biz_priority"`  //
	CfBlocker       []string `xml:"cf_blocker"`       // ---
	CfIITDeployment []string `xml:"cf_it_deployment"` // ---
	Token           []string `xml:"token"`

	Votes int `xml:"votes"` // 0

	Flags []Flag `xml:"flag"`

	CommentSortOrder string `xml:"comment_sort_order"` // oldest_to_newest

	Comments    []Comment
	Attachments []Attachment
}

type shadowBug struct {
	Bug
	CreationTS  bzTime             `xml:"creation_ts"` // 2017-07-03 13:29:00 +0000
	DeltaTS     bzTime             `xml:"delta_ts"`    // 2019-03-27 10:45:20 +0000
	Attachments []shadowAttachment `xml:"attachment"`
	Comments    []shadowComment    `xml:"long_desc"`
}

type Attachment struct {
	IsObsolete int       `xml:"isobsolete,attr"`
	IsPatch    int       `xml:"ispatch,attr"`
	IsPrivate  int       `xml:"isprivate,attr"`
	AttachId   int       `xml:"attachid"`
	Date       time.Time // `xml:"date"`
	DeltaTS    time.Time // `xml:"delta_ts"`
	Desc       string    `xml:"desc"`
	Filename   string    `xml:"filename"`
	Type       string    `xml:"type"`
	Size       int       `xml:"size"`
	Attacher   User      `xml:"attacher"`
	Token      string    `xml:"token"`
}

type shadowAttachment struct {
	Attachment
	Date    bzTime `xml:"date"`
	DeltaTS bzTime `xml:"delta_ts"`
}

type Comment struct {
	IsPrivate int  `xml:"is_private,attr"`
	Id        int  `xml:"commentid"`
	Count     int  `xml:"comment_count"`
	Who       User `xml:"who"`
	BugWhen   time.Time
	TheText   string `xml:"thetext"`
}

type shadowComment struct {
	Comment
	BugWhen bzTime `xml:"bug_when"`
}

type Result struct {
	XMLName xml.Name  `xml:"bugzilla"`
	Shadow  shadowBug `xml:"bug"`
}

// A wrapper that represents the time based on the format emitted by
// Bugzilla
type bzTime struct {
	time.Time
}

func (m *bzTime) UnmarshalText(p []byte) error {
	t, err := time.Parse("2006-01-02 15:04:05 +0000", string(p))
	if err != nil {
		return err
	}
	m.Time = t
	return nil
}

func (c *Client) decodeBug(data []byte) (*Bug, error) {
	var result Result
	err := xml.Unmarshal(data, &result)
	if err != nil {
		return nil, ConnectionError{err}
	}

	var bug Bug
	// This is getting annoying:
	bug = result.Shadow.Bug
	bug.CreationTS = result.Shadow.CreationTS.Time
	bug.DeltaTS = result.Shadow.DeltaTS.Time

	for _, shadowAttachment := range result.Shadow.Attachments {
		att := Attachment{}
		att = shadowAttachment.Attachment
		att.Date = shadowAttachment.Date.Time
		att.DeltaTS = shadowAttachment.DeltaTS.Time
		bug.Attachments = append(bug.Attachments, att)
	}
	bug.Comments = make([]Comment, len(result.Shadow.Comments))
	for i, shadowComment := range result.Shadow.Comments {
		bug.Comments[i] = shadowComment.Comment
		bug.Comments[i].BugWhen = shadowComment.BugWhen.Time
	}

	return &bug, nil
}

func (c *Client) PatchBug(source string) string {
	r := regexp.MustCompile("(?s:<flag (.*?)/>)")
	return r.ReplaceAllString(source, "<flag $1></flag>")
}

func (c *Client) GetBug(id int) (*Bug, error) {
	// query.Set("ctype", "xml")
	// query.Set("excludefield", "attachmentdata")
	url, err := c.getShowBugURL(id, map[string]string{"ctype": "xml", "excludefield": "attachmentdata"})
	if err != nil {
		return nil, err
	}

	resp, err := c.seriousClient.Get(url)
	if err != nil {
		return nil, ConnectionError{err}
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	limitedReader := &io.LimitedReader{R: resp.Body, N: 10 * 1024 * 1024}
	body, err := ioutil.ReadAll(limitedReader)
	if err != nil {
		return nil, ConnectionError{err}
	}

	patched := c.PatchBug(string(body))

	bug, err := c.decodeBug([]byte(patched))
	return bug, err
}

type BugzillaError struct{ error }

func (e BugzillaError) Error() string {
	return fmt.Sprintf("Error from Bugzilla: %v", e.error)
}

func (c *Client) inspectBugzillaResponse() (err error) {
	dom := c.browser.Dom()
	html, err := dom.Html()
	if err != nil {
		return BugzillaError{fmt.Errorf("could not fetch the HTML response from bugzilla: %v", err)}
	}

	if strings.Contains(html, "Mid-air collision!") {
		return BugzillaError{fmt.Errorf("Mid-air collision!")}
	}
	if strings.Contains(html, "reason=invalid_token") {
		return BugzillaError{fmt.Errorf("Invalid token! (Ask the developers!)")}
	}
	if !strings.Contains(html, "Changes submitted for") {
		messages := make([]string, 0)
		re := regexp.MustCompile(`[ \s]+`)
		dom.Find("p").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			text = re.ReplaceAllString(text, " ")
			if strings.Contains(text, "Please go back") {
				return
			}
			messages = append(messages, text)
		})
		if len(messages) == 0 {
			return BugzillaError{fmt.Errorf("Unknown error while submitting the form")}
		}
		joined := strings.Join(messages, "; ")
		if strings.Contains(joined, "Match Failed;") {
			return BugzillaError{fmt.Errorf("Bugzilla was unable to make any match at all for one or more of the names and/or email addresses")}
		}
		return BugzillaError{fmt.Errorf("Message: %s", joined)}
	}
	return
}

func (c *Client) clearNeedinfo(form browser.Submittable, all bool) (err error) {
	count := 0
	dom := c.browser.Dom()
	dom.Find("input[id^=needinfo_override_]").Each(func(i int, s *goquery.Selection) {
		if count > 0 && !all {
			err = RequestError{fmt.Errorf("More than one needinfo found")}
			return
		}
		v, _ := s.Attr("id")
		form.Check(v)
		count++
	})
	return
}

func (c *Client) findClearNeedinfoFor(email string) (controlName string, err error) {
	dom := c.browser.Dom()
	found := false
	expr := fmt.Sprintf(`input[name^="requestee-"][value="%s"]`, email)
	dom.Find(expr).Each(func(i int, s *goquery.Selection) {
		fullName := s.AttrOr("name", "")
		id := fullName[len("requestee-"):]
		maybeControl := fmt.Sprintf("needinfo_override_%s", id)
		overrideExpr := fmt.Sprintf("input[name=%s]", maybeControl)
		dom.Find(overrideExpr).Each(func(i int, s *goquery.Selection) {
			found = true
			controlName = maybeControl
		})
	})
	if !found {
		err = BugzillaError{fmt.Errorf("no control found for the email %v", email)}
		return
	}
	return
}

var PriorityMap = map[string]string{
	"P0": "P0 - Crit Sit",
	"P1": "P1 - Urgent",
	"P2": "P2 - High",
	"P3": "P3 - Medium",
	"P4": "P4 - Low",
	"P5": "P5 - None",
}

// Changes to be performed by Update() for a given bug
type Changes struct {
	SetNeedinfo       string
	RemoveNeedinfo    string
	ClearNeedinfo     bool
	ClearAllNeedinfos bool

	AddComment       string
	CommentIsPrivate bool

	SetURL         string
	SetAssignee    string
	SetPriority    string
	SetDescription string
	SetWhiteboard  string
	SetStatus      string
	SetResolution  string
	SetDuplicate   int

	AddCc    string
	RemoveCc string
	CcMyself bool
}

func (c *Client) Update(id int, changes Changes) (err error) {
	url, err := c.getShowBugURL(id, nil)
	if err != nil {
		return
	}
	c.browser.Open(url)
	form, err := c.browser.Form("form[name=changeform]")
	if err != nil {
		return BugzillaError{fmt.Errorf("failed to find the form element in the bug html: %v", err)}
	}
	if changes.SetNeedinfo != "" {
		form.Set("needinfo", "1")
		form.Set("needinfo_role", "other")
		form.Set("needinfo_from", changes.SetNeedinfo)
	}
	if changes.RemoveNeedinfo != "" {
		control := ""
		control, err = c.findClearNeedinfoFor(changes.RemoveNeedinfo)
		if err != nil {
			return
		}
		form.Set(control, "1")
	}
	if changes.ClearNeedinfo {
		err = c.clearNeedinfo(form, changes.ClearAllNeedinfos)
		if err != nil {
			return
		}
	}
	if changes.AddComment != "" {
		form.Set("comment", changes.AddComment)
		if changes.CommentIsPrivate {
			form.Set("comment_is_private", "1")
			form.Set("commentprivacy", "1")
		}
	}
	if changes.SetURL != "" {
		form.Set("bug_file_loc", changes.SetURL)
	}
	if changes.SetAssignee != "" {
		form.Set("assigned_to", changes.SetAssignee)
	}
	if changes.SetDescription != "" {
		form.Set("short_desc", changes.SetDescription)
	}
	if changes.SetPriority != "" {
		prio, ok := PriorityMap[changes.SetPriority]
		if !ok {
			return BugzillaError{fmt.Errorf("invalid priority value: %v", changes.SetPriority)}
		}
		form.Set("priority", prio)
	}
	if changes.AddCc != "" {
		form.Set("newcc", changes.AddCc)
	}
	if changes.RemoveCc != "" {
		form.Set("cc", changes.RemoveCc)
		form.Set("removecc", "1")
	}
	if changes.CcMyself {
		form.Set("addselfcc", "1")
	}
	if changes.SetWhiteboard != "" {
		form.Set("status_whiteboard", changes.SetWhiteboard)
	}
	if changes.SetStatus != "" {
		form.Set("bug_status", changes.SetStatus)
	}
	if changes.SetResolution != "" {
		form.Set("resolution", changes.SetResolution)
	}
	if changes.SetDuplicate != 0 {
		form.Set("dup_id", fmt.Sprintf("%d", changes.SetDuplicate))
	}
	err = form.Submit()
	if err != nil {
		return BugzillaError{fmt.Errorf("failed to send a request to bugzilla: %v", err)}
	}
	err = c.inspectBugzillaResponse()
	return
}
