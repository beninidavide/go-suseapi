// Package bugzilla can get bugs, attachments and update them
// Instead of the nice XMLRPC interface, it uses the web interface, in
// order to allow changing flags (AFAIR) not available in the API.
package bugzilla

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/browser"
)

// RequestError happens when building the request
type RequestError struct{ error }

func (e RequestError) Error() string {
	return fmt.Sprintf("cannot build request: %v", e.error)
}

// ConnectionError happens when performing the request
type ConnectionError struct{ error }

func (e ConnectionError) Error() string {
	return fmt.Sprintf("cannot communicate with server: %v", e.error)
}

// Cacher should be anything that takes the name of the object to be cached
// and returns something that can receive writes with the contents and then
// eventually be closed.
type Cacher interface {
	GetWriter(id string) io.WriteCloser
}

// Config sets the parameters needed to set up the client. Cacher can be
// left zeroed.
type Config struct {
	BaseURL  string
	User     string
	Password string
	Cacher   Cacher
}

// Client keeps the state of the client.
type Client struct {
	Config        Config
	browser       *browser.Browser
	seriousClient *http.Client
	cacher        Cacher
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
	rt := useHeader(tr)
	rt.Set("Authorization", getAuth(config))
	client.Transport = rt
	return &client
}

// New prepares a *Client for connecting to the Bugzilla Web interface
func New(config Config) (*Client, error) {
	browser := getBrowser(&config)
	seriousClient := getDecentHTTPClient(&config)
	client := &Client{Config: config, browser: browser, seriousClient: seriousClient, cacher: config.Cacher}
	return client, nil
}

type withHeader struct {
	http.Header
	rt http.RoundTripper
}

func useHeader(rt http.RoundTripper) withHeader {
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

func (c *Client) getDownloadURL(id int) (string, error) {
	url, err := url.Parse(c.Config.BaseURL)
	if err != nil {
		return "", RequestError{err}
	}

	url.Path = path.Join(url.Path, "attachment.cgi")

	query := url.Query()
	query.Set("id", fmt.Sprintf("%d", id))
	url.RawQuery = query.Encode()

	return url.String(), nil
}

// User represents user as used in assigned_to, comment author and other
// fields (except Cc.)
type User struct {
	Name  string `xml:"name,attr" json:"name"`
	Email string `xml:",chardata" json:"email"`
}

// Group is a group as seen by Bugzilla
type Group struct {
	ID   int    `xml:"id,attr" json:"id"`
	Name string `xml:",chardata" json:"email"`
}

// Flag represents flags such as needinfo
type Flag struct {
	Name      string `xml:"name,attr" json:"name"`
	ID        int    `xml:"id,attr" json:"id"`
	TypeID    int    `xml:"type_id,attr" json:"type_id"`
	Status    string `xml:"status,attr" json:"status"`
	Setter    string `xml:"setter,attr" json:"setter"`
	Requestee string `xml:"requestee,attr" json:"requestee"`
}

// Bug is a Bug in Bugzilla
type Bug struct {
	Reporter   User `xml:"reporter" json:"reporter"`
	AssignedTo User `xml:"assigned_to" json:"assigned_to"`
	QAContact  User `xml:"qa_contact" json:"qa_contact"`

	Groups []Group `xml:"group" json:"group"`

	BugID              int       `xml:"bug_id" json:"bug_id"`                           // 1047068
	CreationTS         time.Time `json:"creation_ts"`                                   // 2017-07-03 13:29:00 +0000
	ShortDesc          string    `xml:"short_desc" json:"short_desc"`                   // L4: test cloud bug
	DeltaTS            time.Time `json:"delta_ts"`                                      // 2019-03-27 10:45:20 +0000
	ReporterAccessible int       `xml:"reporter_accessible" json:"reporter_accessible"` // 0
	CCListAccessible   int       `xml:"cclist_accessible" json:"cclist_accessible"`     // 0
	ClassificationID   int       `xml:"classification_id" json:"classification_id"`     // 111
	Classification     string    `xml:"classification" json:"classification"`           // foobar Frobnicator Cloud
	Product            string    `xml:"product" json:"product"`                         // foobar Frobnicator Cloud 7
	Component          string    `xml:"component" json:"component"`                     // Frobtool
	Version            string    `xml:"version" json:"version"`                         // Milestone 8
	RepPlatform        string    `xml:"rep_platform" json:"rep_platform"`               // Other
	OpSys              string    `xml:"op_sys" json:"op_sys"`                           // Other
	BugStatus          string    `xml:"bug_status" json:"bug_status"`                   // RESOLVED
	Resolution         string    `xml:"resolution" json:"resolution"`                   // FIXED
	DupID              int       `xml:"dup_id" json:"dup_id"`

	BugFileLoc       string `xml:"bug_file_loc" json:"bug_file_loc"`           //
	StatusWhiteboard string `xml:"status_whiteboard" json:"status_whiteboard"` // wasL3:48626  zzz
	Keywords         string `xml:"keywords" json:"keywords"`                   // DSLA_REQUIRED, DSLA_SOLUTION_PROVIDED
	Priority         string `xml:"priority" json:"priority"`                   // P5 - None
	BugSeverity      string `xml:"bug_severity" json:"bug_severity"`           // Normal
	TargetMilestone  string `xml:"target_milestone" json:"target_milestone"`   // ---

	EverConfirmed int      `xml:"everconfirmed" json:"everconfirmed"`   // 1
	Cc            []string `xml:"cc" json:"cc"`                         // user@foobar.com
	EstimatedTime string   `xml:"estimated_time" json:"estimated_time"` // 0.00
	RemainingTime string   `xml:"remaining_time" json:"remaining_time"` // 0.00
	ActualTime    string   `xml:"actual_time" json:"actual_time"`       // 0.00

	CfFoundby       []string `xml:"cf_foundby" json:"cf_foundby"`             // ---
	CfNtsPriority   []string `xml:"cf_nts_priority" json:"cf_nts_priority"`   //
	CfBizPriority   []string `xml:"cf_biz_priority" json:"cf_biz_priority"`   //
	CfBlocker       []string `xml:"cf_blocker" json:"cf_blocker"`             // ---
	CfIITDeployment []string `xml:"cf_it_deployment" json:"cf_it_deployment"` // ---
	Token           []string `xml:"token" json:"token"`

	Votes int `xml:"votes" json:"votes"` // 0

	Flags []Flag `xml:"flag" json:"flag"`

	CommentSortOrder string `xml:"comment_sort_order" json:"comment_sort_order"` // oldest_to_newest

	Comments    []*Comment
	Attachments []*Attachment
}

type shadowBug struct {
	Error string `xml:"error,attr"`

	Bug
	CreationTS  bzTime             `xml:"creation_ts" json:"creation_ts"` // 2017-07-03 13:29:00 +0000
	DeltaTS     bzTime             `xml:"delta_ts" json:"delta_ts"`       // 2019-03-27 10:45:20 +0000
	Attachments []shadowAttachment `xml:"attachment" json:"attachment"`
	Comments    []shadowComment    `xml:"long_desc" json:"long_desc"`
}

// Attachment as provided by the bug information page. This struct has only
// name, size and attachid set when coming from DownloadAttachment, as it's
// extracted from the HTTP headers.
type Attachment struct {
	IsObsolete int       `xml:"isobsolete,attr" json:"isobsolete"`
	IsPatch    int       `xml:"ispatch,attr" json:"ispatch"`
	IsPrivate  int       `xml:"isprivate,attr" json:"isprivate"`
	AttachID   int       `xml:"attachid" json:"attachid"`
	Date       time.Time // `xml:"date" json:"date"`
	DeltaTS    time.Time // `xml:"delta_ts" json:"delta_ts"`
	Desc       string    `xml:"desc" json:"desc"`
	Filename   string    `xml:"filename" json:"filename"`
	Type       string    `xml:"type" json:"type"`
	Size       int       `xml:"size" json:"size"`
	Attacher   User      `xml:"attacher" json:"attacher"`
	Token      string    `xml:"token" json:"token"`
}

type shadowAttachment struct {
	Attachment
	Date    bzTime `xml:"date" json:"date"`
	DeltaTS bzTime `xml:"delta_ts" json:"delta_ts"`
}

// Comment as in bug comments
type Comment struct {
	IsPrivate int  `xml:"isprivate,attr" json:"isprivate"`
	ID        int  `xml:"commentid" json:"commentid"`
	Count     int  `xml:"comment_count" json:"comment_count"`
	Who       User `xml:"who" json:"who"`
	BugWhen   time.Time
	TheText   string `xml:"thetext" json:"thetext"`
}

type shadowComment struct {
	Comment
	BugWhen bzTime `xml:"bug_when" json:"bug_when"`
}

type xmlResult struct {
	XMLName xml.Name  `xml:"bugzilla" json:"bugzilla"`
	Shadow  shadowBug `xml:"bug" json:"bug"`
}

// A wrapper that represents the time based on the format emitted by
// Bugzilla
type bzTime struct {
	time.Time
}

func (m *bzTime) UnmarshalText(p []byte) error {
	t, err := time.Parse("2006-01-02 15:04:05 -0700", string(p))
	if err != nil {
		return err
	}
	m.Time = t.UTC()
	return nil
}

func (c *Client) decodeBug(data []byte) (*Bug, error) {
	var result xmlResult
	err := xml.Unmarshal(data, &result)
	if err != nil {
		return nil, ConnectionError{err}
	}

	if result.Shadow.Error != "" {
		return nil, ConnectionError{fmt.Errorf("code: %s", result.Shadow.Error)}
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
		bug.Attachments = append(bug.Attachments, &att)
	}
	for _, shadowComment := range result.Shadow.Comments {
		comm := Comment{}
		comm = shadowComment.Comment
		comm.BugWhen = shadowComment.BugWhen.Time
		bug.Comments = append(bug.Comments, &comm)
	}

	return &bug, nil
}

func (c *Client) patchBug(source []byte) []byte {
	r := regexp.MustCompile("(?s:<flag (.*?)/>)")
	return r.ReplaceAll(source, []byte("<flag $1></flag>"))
}

func (c *Client) cacheBug(bug *Bug) {
	cacher, ok := c.cacher.(Cacher)
	if !ok {
		return
	}
	b, err := json.Marshal(bug)
	if err == nil {
		writer := cacher.GetWriter(fmt.Sprintf("%d", bug.BugID))
		writer.Write(b)
		writer.Close()
	}
}

// GetBug gets a *Bug from the Bugzilla API (apibuzilla)
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

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return nil, ConnectionError{fmt.Errorf(http.StatusText(resp.StatusCode))}
	}

	limitedReader := &io.LimitedReader{R: resp.Body, N: 10 * 1024 * 1024}
	body, err := ioutil.ReadAll(limitedReader)
	if err != nil {
		return nil, ConnectionError{err}
	}

	patched := c.patchBug(body)

	bug, err := c.decodeBug(patched)
	if err == nil {
		c.cacheBug(bug)
	}

	return bug, err
}

// ErrBugzilla is an error from Bugzilla
type ErrBugzilla struct{ error }

func (e ErrBugzilla) Error() string {
	return fmt.Sprintf("Error from Bugzilla: %v", e.error)
}

func (c *Client) inspectBugzillaResponse() (err error) {
	dom := c.browser.Dom()
	html, err := dom.Html()
	if err != nil {
		return ErrBugzilla{fmt.Errorf("could not fetch the HTML response from bugzilla: %v", err)}
	}

	if strings.Contains(html, "Mid-air collision!") {
		return ErrBugzilla{fmt.Errorf("mid-air collision")}
	}
	if strings.Contains(html, "reason=invalid_token") {
		return ErrBugzilla{fmt.Errorf("invalid token! (Ask the developers!)")}
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
			return ErrBugzilla{fmt.Errorf("Unknown error while submitting the form")}
		}
		joined := strings.Join(messages, "; ")
		if strings.Contains(joined, "Match Failed;") {
			return ErrBugzilla{fmt.Errorf("Bugzilla was unable to make any match at all for one or more of the names and/or email addresses")}
		}
		return ErrBugzilla{fmt.Errorf("Message: %s", joined)}
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
		err = ErrBugzilla{fmt.Errorf("no control found for the email %v", email)}
		return
	}
	return
}

// PriorityMap maps short priority names to the longer ones, as provided by
// the Web Interface
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

	// DeltaTS should have the timestamp of the last change
	DeltaTS      time.Time
	CheckDeltaTS bool
}

func getDeltaTS(form browser.Submittable) (t *time.Time, err error) {
	raw, err := form.Value("delta_ts")
	if err != nil {
		return nil, ErrBugzilla{fmt.Errorf("can't detect mid-air collision without delta_ts in the bug form: %v", err)}
	}
	raw += " +0000" // this is a workaround against a bad delta_ts sent by the web interface
	var delta bzTime
	err = delta.UnmarshalText([]byte(raw))
	if err != nil {
		return nil, ErrBugzilla{fmt.Errorf("failed to parse delta_ts to prevent mid-air collision")}
	}
	t = &delta.Time
	return
}

func (c *Client) checkDeltaTS(changes *Changes, form browser.Submittable) error {
	if changes.CheckDeltaTS {
		delta, err := getDeltaTS(form)
		if err != nil {
			return err
		}

		if !delta.Equal(changes.DeltaTS) {
			return ErrBugzilla{fmt.Errorf("likely mid-air collision: the bug has been updated at %v", delta)}
		}
	}
	return nil
}

// Update changes a bug with the attribute to be modified provided by
// Changes
func (c *Client) Update(id int, changes Changes) (err error) {
	url, err := c.getShowBugURL(id, nil)
	if err != nil {
		return
	}
	c.browser.Open(url)
	form, err := c.browser.Form("form[name=changeform]")
	if err != nil {
		return ErrBugzilla{fmt.Errorf("failed to find the form element in the bug html: %v", err)}
	}
	if err = c.checkDeltaTS(&changes, form); err != nil {
		return err
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
			return ErrBugzilla{fmt.Errorf("invalid priority value: %v", changes.SetPriority)}
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

	// surf fails to parse cclist_accessible and reporter_accessible
	// https://github.com/headzoo/surf/issues/109
	form.Remove("defined_cclist_accessible")
	form.Remove("defined_reporter_accessible")
	form.Remove("defined_group")

	err = form.Submit()
	if err != nil {
		return ErrBugzilla{fmt.Errorf("failed to send a request to bugzilla: %v", err)}
	}
	err = c.inspectBugzillaResponse()
	return
}

func getAttachmentFromHeaders(id int, header http.Header) (*Attachment, error) {
	rawType := header.Get("Content-Disposition")
	_, info, err := mime.ParseMediaType(rawType)
	if err != nil {
		return nil, ConnectionError{fmt.Errorf("failed to parse Content-Disposition: %v", err)}
	}

	size, err := strconv.Atoi(header.Get("Content-Length"))
	if err != nil {
		return nil, ConnectionError{fmt.Errorf("bad Content-Length in response")}
	}

	name, _ := info["filename"]
	att := &Attachment{AttachID: id, Filename: name, Size: size}

	return att, nil
}

// DownloadAttachment an attachment for download
// Returns an Attachment with only the Size and Filename filled, a reader
// and error.
func (c *Client) DownloadAttachment(id int) (*Attachment, io.ReadCloser, error) {
	url, err := c.getDownloadURL(id)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.seriousClient.Get(url)
	if err != nil {
		return nil, nil, ConnectionError{err}
	}

	att, err := getAttachmentFromHeaders(id, resp.Header)
	if err != nil {
		return nil, nil, err
	}

	return att, resp.Body, nil
}

// GetBug gets a *Bug from a JSON blob
func (c *Client) GetBugFromJSON(source io.Reader) (*Bug, error) {
	var bug Bug
	decoder := json.NewDecoder(source)
	err := decoder.Decode(&bug)
	if err != nil {
		return nil, err
	}
	return &bug, nil
}
