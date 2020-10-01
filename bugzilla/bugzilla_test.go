package bugzilla_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/bhdn/go-suseapi/bugzilla"
	. "gopkg.in/check.v1"
)

type clientSuite struct {
	bz *bugzilla.Client
}

var _ = Suite(&clientSuite{})

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

func (cs *clientSuite) SetUpTest(c *C) {
}

func makeClient(url string) *bugzilla.Client {
	config := bugzilla.Config{BaseURL: url,
		User: "me", Password: "letmein"}
	bz, _ := bugzilla.New(config)
	return bz
}

func (cs *clientSuite) TestCreateClient(c *C) {
	bz := makeClient("http://foobar.com/")
	c.Assert(bz, NotNil)
}

func (cs *clientSuite) TestPatchBug(c *C) {
	bz := makeClient("")
	c.Assert(string(bz.PatchBug([]byte("<flag foo=bar\n />"))), Equals, "<flag foo=bar\n ></flag>")
}

const bugXml = `
<bugzilla version="4.4.12" urlbase="http://bugzilla.foobar.com/" maintainer="maintainer@foobar.com" exporter="username@foobar.com">

    <bug>
          <bug_id>1047068</bug_id>

          <creation_ts>2017-07-03 13:29:00 +0000</creation_ts>
          <short_desc>L4: test cloud bug</short_desc>
          <delta_ts>2019-03-27 10:45:20 +0000</delta_ts>
          <reporter_accessible>0</reporter_accessible>
          <cclist_accessible>0</cclist_accessible>
          <classification_id>111</classification_id>
          <classification>foobar Frobnicator Cloud</classification>
          <product>foobar Frobnicator Cloud 7</product>
          <component>Frob</component>
          <version>Milestone 8</version>
          <rep_platform>Other</rep_platform>
          <op_sys>Other</op_sys>
          <bug_status>RESOLVED</bug_status>
          <resolution>FIXED</resolution>


          <bug_file_loc></bug_file_loc>
          <status_whiteboard>wasZZ:48626  zzz</status_whiteboard>
          <keywords>FIRST_KEYWORD, SECOND_KEYWORD</keywords>
          <priority>P5 - None</priority>
          <bug_severity>Normal</bug_severity>
          <target_milestone>---</target_milestone>


          <everconfirmed>1</everconfirmed>
          <reporter name="Firstname Lastname">username@foobar.com</reporter>
          <assigned_to name="Firstname Lastname">username@foobar.com</assigned_to>
          <cc>username@foobar.com</cc>

    <cc>anotheremail@gmail.com</cc>
          <estimated_time>0.00</estimated_time>
          <remaining_time>0.00</remaining_time>
          <actual_time>0.00</actual_time>

          <qa_contact name="Firstname Lastname">username@foobar.com</qa_contact>
          <cf_foundby>---</cf_foundby>
          <cf_nts_priority></cf_nts_priority>
          <cf_biz_priority></cf_biz_priority>
          <cf_blocker>---</cf_blocker>
          <cf_marketing_qa_status>---</cf_marketing_qa_status>
          <cf_it_deployment>---</cf_it_deployment>
          <cf_foundby>---</cf_foundby>
          <cf_nts_priority></cf_nts_priority>
          <cf_biz_priority></cf_biz_priority>
          <votes>0</votes>


          <comment_sort_order>oldest_to_newest</comment_sort_order>
          <token>1553684018-xPLcA7NX7EE2dlA_egJKcjTWk4pzA1QCBzbcKXuG3yU</token>

      <flag name="needinfo"
          id="201661"
          type_id="4"
          status="?"
          setter="username@foobar.com"
          requestee="username@foobar.com"
    />
    <flag name="needinfo"
          id="201662"
          type_id="4"
          status="?"
          setter="username@foobar.com"
          requestee="username@foobar.com"
    />
    <flag name="SHIP_STOPPER"
          id="201663"
          type_id="2"
          status="?"
          setter="username@foobar.com"
          requestee="username@foobar.com"
    />

          <group id="10">foobaronly</group>
          <group id="17">foobar Enterprise Partner</group>



          <long_desc isprivate="0">
    <commentid>7315202</commentid>
    <comment_count>0</comment_count>
    <who name="Firstname Lastname">username@foobar.com</who>
    <bug_when>2017-07-03 13:29:15 +0000</bug_when>
    <thetext>This is a test cloud incident.</thetext>
  </long_desc><long_desc isprivate="0">
    <commentid>7986182</commentid>
    <comment_count>57</comment_count>
    <who name="Firstname Lastname">username@foobar.com</who>
    <bug_when>2018-12-20 17:28:07 +0000</bug_when>
    <thetext>(In reply to username@foobar.com from comment 56)
&gt; test file comment with -r

skldjskdj</thetext>
  </long_desc><long_desc isprivate="0">
    <commentid>7986183</commentid>
    <comment_count>58</comment_count>
    <who name="Firstname Lastname">username@foobar.com</who>
    <bug_when>2018-12-20 17:28:42 +0000</bug_when>
    <thetext>comment</thetext>
  </long_desc><long_desc isprivate="0">
    <commentid>7986184</commentid>
    <comment_count>59</comment_count>
    <who name="Firstname Lastname">username@foobar.com</who>
    <bug_when>2018-12-20 17:28:58 +0000</bug_when>
    <thetext>(In reply to username@foobar.com from comment 58)
&gt; comment

comment and reply -- reply takes priority</thetext>
  </long_desc>
 

          <attachment isobsolete="0" ispatch="0" isprivate="0">
            <attachid>766283</attachid>
            <date>2018-04-06 12:48:00 +0000</date>
            <delta_ts>2018-04-06 12:48:24 +0000</delta_ts>
            <desc>description</desc>
            <filename>a.txt</filename>
            <type>text/plain</type>
            <size>2</size>
            <attacher name="Firstname Lastname">username@foobar.com</attacher>

              <token>1553684018-Zr0rbiHmQyVbVum-470APzDGkCc30CRnMiPqvaQBDpY</token>

          </attachment>
          <attachment isobsolete="0" ispatch="0" isprivate="0">
            <attachid>766284</attachid>
            <date>2018-04-06 12:50:00 +0000</date>
            <delta_ts>2018-04-06 12:50:44 +0000</delta_ts>
            <desc>description</desc>
            <filename>a.txt</filename>
            <type>text/plain</type>
            <size>2</size>
            <attacher name="Firstname Lastname">username@foobar.com</attacher>

              <token>1553684018-9kdHLOKDEEOKPpnKhNQTgDt24Elhhv3amg6zp-vGlsI</token>

          </attachment>


    </bug>

</bugzilla>
`

func (cs *clientSuite) TestGetBug(c *C) {
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/show_bug.cgi":
			query := r.URL.Query()
			c.Assert(query.Get("id"), Equals, "1047068")
			io.WriteString(w, bugXml)
		default:
			http.Error(w, "Unimplemented", 500)
			return
		}
	}))
	defer ts0.Close()

	bz := makeClient(ts0.URL)
	bug, err := bz.GetBug(1047068)
	c.Assert(err, IsNil)
	c.Assert(bug, NotNil)
	c.Assert(bug.BugID, Equals, 1047068)
	c.Assert(bug.ShortDesc, Equals, "L4: test cloud bug")
	c.Assert(bug.CreationTS, Equals, time.Date(2017, 7, 3, 13, 29, 0, 0, time.UTC))
	c.Assert(bug.DeltaTS, Equals, time.Date(2019, 3, 27, 10, 45, 20, 0, time.UTC))
	c.Assert(bug.ReporterAccessible, Equals, 0)
	c.Assert(bug.CCListAccessible, Equals, 0)
	c.Assert(bug.Classification, Equals, "foobar Frobnicator Cloud")
	c.Assert(bug.ClassificationID, Equals, 111)
	c.Assert(bug.Product, Equals, "foobar Frobnicator Cloud 7")
	c.Assert(bug.Component, Equals, "Frob")
	c.Assert(bug.Version, Equals, "Milestone 8")
	c.Assert(bug.RepPlatform, Equals, "Other")
	c.Assert(bug.OpSys, Equals, "Other")
	c.Assert(bug.BugStatus, Equals, "RESOLVED")
	c.Assert(bug.Resolution, Equals, "FIXED")
	c.Assert(bug.BugFileLoc, Equals, "")
	c.Assert(bug.StatusWhiteboard, Equals, "wasZZ:48626  zzz")
	c.Assert(bug.Keywords, Equals, "FIRST_KEYWORD, SECOND_KEYWORD")
	c.Assert(bug.Priority, Equals, "P5 - None")
	c.Assert(bug.BugSeverity, Equals, "Normal")
	c.Assert(bug.TargetMilestone, Equals, "---")
	c.Assert(bug.EverConfirmed, Equals, 1)
	c.Assert(bug.Reporter.Name, Equals, "Firstname Lastname")
	c.Assert(bug.Reporter.Email, Equals, "username@foobar.com")
	c.Assert(len(bug.Cc), Equals, 2)
	sort.Strings(bug.Cc)
	c.Assert(bug.Cc[0], Equals, "anotheremail@gmail.com")
	c.Assert(bug.Cc[1], Equals, "username@foobar.com")
	c.Assert(bug.EstimatedTime, Equals, "0.00")
	c.Assert(bug.ActualTime, Equals, "0.00")
	c.Assert(bug.RemainingTime, Equals, "0.00")
	c.Assert(bug.QAContact.Name, Equals, "Firstname Lastname")
	c.Assert(bug.Votes, Equals, 0)
	c.Assert(len(bug.Groups), Equals, 2)
	c.Assert(bug.Groups[0].ID, Equals, 10)
	c.Assert(bug.Groups[0].Name, Equals, "foobaronly")
	c.Assert(bug.Groups[1].ID, Equals, 17)
	c.Assert(bug.Groups[1].Name, Equals, "foobar Enterprise Partner")
	c.Assert(bug.CommentSortOrder, Equals, "oldest_to_newest")
	c.Assert(len(bug.Comments), Equals, 4)
	c.Assert(bug.Comments[0].ID, Equals, 7315202)
	c.Assert(bug.Comments[0].IsPrivate, Equals, 0)
	c.Assert(bug.Comments[0].Count, Equals, 0)
	c.Assert(bug.Comments[0].BugWhen, Equals, time.Date(2017, 07, 03, 13, 29, 15, 0, time.UTC))
	c.Assert(bug.Comments[0].Who.Name, Equals, "Firstname Lastname")
	c.Assert(bug.Comments[0].Who.Email, Equals, "username@foobar.com")
	c.Assert(bug.Comments[0].TheText, Equals, "This is a test cloud incident.")
	c.Assert(bug.Comments[1].ID, Equals, 7986182)
	c.Assert(bug.Comments[1].IsPrivate, Equals, 0)
	c.Assert(bug.Comments[1].Count, Equals, 57)
	c.Assert(bug.Comments[1].Who.Name, Equals, "Firstname Lastname")
	c.Assert(bug.Comments[1].Who.Email, Equals, "username@foobar.com")
	c.Assert(bug.Comments[1].TheText, Equals, `(In reply to username@foobar.com from comment 56)
> test file comment with -r

skldjskdj`)
	c.Assert(bug.Comments[2].ID, Equals, 7986183)
	c.Assert(bug.Comments[2].IsPrivate, Equals, 0)
	c.Assert(bug.Comments[2].Count, Equals, 58)
	c.Assert(bug.Comments[2].Who.Name, Equals, "Firstname Lastname")
	c.Assert(bug.Comments[2].Who.Email, Equals, "username@foobar.com")
	c.Assert(bug.Comments[2].TheText, Equals, "comment")
	c.Assert(len(bug.Flags), Equals, 3)
	c.Assert(bug.Flags[0].Name, Equals, "needinfo")
	c.Assert(bug.Flags[0].ID, Equals, 201661)
	c.Assert(bug.Flags[0].TypeID, Equals, 4)
	c.Assert(bug.Flags[0].Status, Equals, "?")
	c.Assert(bug.Flags[0].Setter, Equals, "username@foobar.com")
	c.Assert(bug.Flags[0].Requestee, Equals, "username@foobar.com")
	c.Assert(bug.Flags[1].Name, Equals, "needinfo")
	c.Assert(bug.Flags[1].ID, Equals, 201662)
	c.Assert(bug.Flags[1].TypeID, Equals, 4)
	c.Assert(bug.Flags[1].Status, Equals, "?")
	c.Assert(bug.Flags[1].Setter, Equals, "username@foobar.com")
	c.Assert(bug.Flags[1].Requestee, Equals, "username@foobar.com")
	c.Assert(bug.Flags[2].Name, Equals, "SHIP_STOPPER")
	c.Assert(bug.Flags[2].ID, Equals, 201663)
	c.Assert(bug.Flags[2].TypeID, Equals, 2)
	c.Assert(bug.Flags[2].Status, Equals, "?")
	c.Assert(bug.Flags[2].Setter, Equals, "username@foobar.com")
	c.Assert(bug.Flags[2].Requestee, Equals, "username@foobar.com")
	c.Assert(len(bug.Attachments), Equals, 2)
	c.Assert(bug.Attachments[0].IsObsolete, Equals, 0)
	c.Assert(bug.Attachments[0].IsPrivate, Equals, 0)
	c.Assert(bug.Attachments[0].IsPatch, Equals, 0)
	c.Assert(bug.Attachments[0].Date, Equals, time.Date(2018, 4, 6, 12, 48, 0, 0, time.UTC))
	c.Assert(bug.Attachments[0].Filename, Equals, "a.txt")
	c.Assert(bug.Attachments[1].IsObsolete, Equals, 0)
	c.Assert(bug.Attachments[1].IsPrivate, Equals, 0)
	c.Assert(bug.Attachments[1].IsPatch, Equals, 0)
	c.Assert(bug.Attachments[1].Date, Equals, time.Date(2018, 4, 6, 12, 50, 0, 0, time.UTC))
	c.Assert(bug.Attachments[1].Filename, Equals, "a.txt")
}

func (cs *clientSuite) TestDifferentTZInBugzilla(c *C) {
	xml := strings.Replace(bugXml, "+0000", "-0200", -1)
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/show_bug.cgi":
			query := r.URL.Query()
			c.Assert(query.Get("id"), Equals, "1047068")
			io.WriteString(w, xml)
		default:
			http.Error(w, "Unimplemented", 500)
			return
		}
	}))
	defer ts0.Close()

	bz := makeClient(ts0.URL)
	bug, err := bz.GetBug(1047068)
	c.Assert(err, IsNil)
	c.Assert(bug, NotNil)
	c.Assert(bug.BugID, Equals, 1047068)
	c.Assert(bug.CreationTS, Equals, time.Date(2017, 7, 3, 15, 29, 0, 0, time.UTC))
}

func makeClientWithCache(url string, cacher bugzilla.Cacher) *bugzilla.Client {
	config := bugzilla.Config{BaseURL: url,
		User: "me", Password: "letmein", Cacher: cacher}
	bz, _ := bugzilla.New(config)
	return bz
}

type CacherHelper struct {
	bugzilla.Cacher

	location string
	buf      FakeBuf
	id       string
}

type FakeBuf struct {
	bytes.Buffer
}

func (f *FakeBuf) Close() error {
	return nil
}

func (c *CacherHelper) GetWriter(id string) io.WriteCloser {
	c.id = id
	return &c.buf
}

type cacheChecker struct {
	BugID      int            `json:"bug_id"`
	ShortDesc  string         `json:"short_desc"`
	AssignedTo *bugzilla.User `json:"assigned_to"`
}

func (cs *clientSuite) TestGetBugWithCacher(c *C) {
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/show_bug.cgi":
			query := r.URL.Query()
			c.Assert(query.Get("id"), Equals, "1047068")
			io.WriteString(w, bugXml)
		default:
			http.Error(w, "Unimplemented", 500)
			return
		}
	}))
	defer ts0.Close()

	var cacher CacherHelper
	bz := makeClientWithCache(ts0.URL, &cacher)
	bug, err := bz.GetBug(1047068)
	c.Assert(err, IsNil)
	c.Assert(cacher.id, Equals, "1047068")
	c.Assert(bug.BugID, Equals, 1047068)
	var checker cacheChecker
	err = json.Unmarshal(cacher.buf.Bytes(), &checker)
	c.Assert(err, IsNil)
	c.Check(checker.BugID, Equals, 1047068)
	c.Check(checker.ShortDesc, Equals, "L4: test cloud bug")
	c.Check(checker.AssignedTo.Name, Equals, "Firstname Lastname")
}

var showBugHtml = `
<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"
                      "http://www.w3.org/TR/html4/loose.dtd">
<html lang="en">
  <head>
    <title>Bug 1047068 &ndash; ZZ: L4: test cloud bug</title>

      <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
  </head>



  <body >



<div id="header">
<div id="banner">
  </div>

<table border="0" cellspacing="0" cellpadding="0" id="titles">
<tr>
    <td id="title">
      <p>Bugzilla &ndash; Bug&nbsp;1047068</p>
    </td>

    <td id="subtitle">
      <p class="subheader">ZZ: L4: test cloud bug</p>
    </td>

    <td id="information">
      <p class="header_addl_info">Last modified: 2019-03-28 11:40:39 UTC</p>
    </td>
</tr>
</table>

<table id="lang_links_container" cellpadding="0" cellspacing="0"
       class="bz_default_hidden"><tr><td>
</td></tr></table>
<ul class="links">
  <li><a href="./">Home</a></li>
  <li><span class="separator">| </span><a href="enter_bug.cgi">New</a></li>
  <li><span class="separator">| </span><a href="describecomponents.cgi">Browse</a></li>
  <li><span class="separator">| </span><a href="query.cgi">Search</a></li>

  <li class="form">
    <span class="separator">| </span>
    <form action="buglist.cgi" method="get"
        onsubmit="if (this.quicksearch.value == '')
                  { alert('Please enter one or more search terms first.');
                    return false; } return true;">
    <input type="hidden" id="no_redirect_top" name="no_redirect" value="0">
    <input class="txt" type="text" id="quicksearch_top" name="quicksearch" 
           title="Quick Search" value="">
    <input class="btn" type="submit" value="Search" 
           id="find_top"></form>
  <a href="page.cgi?id=quicksearch.html" title="Quicksearch Help">[?]</a></li>

  <li><span class="separator">| </span><a href="report.cgi">Reports</a></li>

  <li>
      <span class="separator">| </span></li>

    <li class="dropdown">
      <span class="anchor">user&#64;foobar.com</span>
      <ul>
        <li><a href="request.cgi?requester=user%40foobar.com&amp;requestee=user%40foobar.com&amp;do_union=1&amp;group=type&amp;action=queue">My Requests</a></li>
        <li><a href="userprefs.cgi">Preferences</a></li>
          <li><a href="admin.cgi">Administration</a></li>
          <li><a href="https://bugzilla.foobar.com/AGLogout">Log out</a></li>
      </ul>
    </li>
<li>
        <span class="separator">| </span>
        <a href="docs/en/html/bug_page.html" target="_blank">Help</a>
      </li>

    <li>
      <span class="separator">| </span></li>
</ul>
</div> 

<div id="bugzilla-body">

<div class="navigation">
  
  <i><font color="#777777">First</font></i>
  <i><font color="#777777">Last</font></i>
  <i><font color="#777777">Prev</font></i>
  <i><font color="#777777">Next</font></i>
  &nbsp;&nbsp;
  <i><font color="#777777">This bug is not in your last
    search results.</font></i>
</div>
<script type="text/javascript">
</script>

<form name="changeform" id="changeform" method="post" action="process_bug.cgi">

  <input type="hidden" name="delta_ts" value="2019-03-28 11:40:39">
  <input type="hidden" name="longdesclength" value="61">
  <input type="hidden" name="id" value="1047068">
  <input type="hidden" name="token" value="1554072294-4daOMysnQcPd3R6pCNCx9r4IekT5pNmfeIrujAD-l0U">
<div class="bz_alias_short_desc_container edit_form"><div class="knob-buttons">
      <input type="submit" value="Save Changes" 
             id="commit_top">
    </div>
        <strike>
     <a href="show_bug.cgi?id=1047068"><b>Bug&nbsp;1047068</b></a> -<span id="summary_alias_container" class="bz_default_hidden">
      <span id="short_desc_nonedit_display">ZZ: L4: test cloud bug</span></strike>
        <small class="editme">(<a href="#" id="editme_action">edit</a>)</small>
     </span>  
       
    <div id="summary_alias_input">
      <table id="summary"> 
        <tr>      <th class="field_label "
    id="field_label_alias">

    <label for="alias">

  <a 
      title="A short, unique name assigned to a bug in order to assist with looking it up and referring to it in other places in Bugzilla."
      class="field_help_link"
      href="page.cgi?id=glossary.html#alias"
  >Alias:</a>
</label>
</th>
            <td><input id="alias" name="alias" class="text_input"
              value="" size="20" maxlength="20">
          </td>
        </tr>
        
        <tr><th class="field_label "
    id="field_label_short_desc">

    <label for="short_desc" accesskey="s">

  <a 
      title="The bug summary is a short sentence which succinctly describes what the bug is about."
      class="field_help_link"
      href="page.cgi?id=glossary.html#short_desc"
  >Summary:</a>
</label>
</th>
          <td><input id="short_desc" name="short_desc" class="text_input"
              value="ZZ: L4: test cloud bug" size="80" maxlength="255" spellcheck="true">
          </td>
        </tr>
      </table>
    </div>
  </div>
  <table class="edit_form">
    <tr>
      
      <td id="bz_show_bug_column_1" class="bz_show_bug_column">     
        <table>
          <tr>
    <th class="field_label">
      <a href="page.cgi?id=status_resolution_matrix.html">Status</a>:
    </th>
    <td id="bz_field_status">
      <span id="static_bug_status">RESOLVED
          FIXED
          (<a href="#add_comment" 
              onclick="window.setTimeout(function() { document.getElementByID('bug_status').focus(); }, 10)">edit</a>)
      </span>
    </td>
  </tr>

          <tr>
    <td colspan="2" class="bz_section_spacer"></td>
  </tr>
          <tr><th class="field_label "
    id="field_label_classification">


  <a 
      title="Bugs are categorised into Classifications, Products and Components. classifications is the top-level categorisation."
      class="field_help_link"
      href="page.cgi?id=glossary.html#classification"
  >Classification:</a>

</th>
  <td class="field_value "
      id="field_container_classification" >foobar frob Cloud</td>
      </tr>

    
    
    
    <tr><th class="field_label "
    id="field_label_product">

    <label for="product">

  <a 
      title="Bugs are categorised into Products and Components. Select a Classification to narrow down this list."
      class="field_help_link"
      href="describecomponents.cgi"
  >Product:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_product" >
        <input type="hidden" id="product_dirty">
        <select id="product" 
                name="product" 
                >
            <option value="foobar frob Cloud 6"
                    id="v1468_product"
              >foobar frob Cloud 6</option>
            <option value="foobar frob Cloud 7"
                    id="v1581_product"
              
                selected="selected">foobar frob Cloud 7</option>
            <option value="foobar frob Cloud 8"
                    id="v1771_product"
              >foobar frob Cloud 8</option>
            <option value="foobar frob Cloud 9"
                    id="v1809_product"
              >foobar frob Cloud 9</option>
        </select>
        

</td>
    </tr>

    
    <tr class="bz_default_hidden"><th class="field_label "
    id="field_label_classification">

    <label for="classification">

  <a 
      title="Bugs are categorised into Classifications, Products and Components. classifications is the top-level categorisation."
      class="field_help_link"
      href="page.cgi?id=glossary.html#classification"
  >Classification:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_classification" >
        <input type="hidden" id="classification_dirty">
        <select id="classification" 
                name="classification" 
                >
            <option value="foobar frob Cloud"
                    id="v111_classification"
              
                selected="selected">foobar frob Cloud</option>
            <option value="foobar Tools"
                    id="v40_classification"
              >foobar Tools</option>
        </select>
        

</td>
    </tr>
        
    
    
    <tr><th class="field_label "
    id="field_label_component">

    <label for="component">

  <a 
      title="Components are second-level categories; each belongs to a particular Product. Select a Product to narrow down this list."
      class="field_help_link"
      href="describecomponents.cgi?product=foobar frob Cloud 7"
  >Component:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_component" >
        <input type="hidden" id="component_dirty">
        <select id="component" 
                name="component" 
                >
            <option value="Component"
                    id="v22104_component"
              >Component</option>
        </select>
</td>
    </tr>
    <tr><th class="field_label "
    id="field_label_version">

    <label for="version">

  <a 
      title="The version field defines the version of the software the bug was found in."
      class="field_help_link"
      href="page.cgi?id=glossary.html#version"
  >Version:</a>
</label>
</th>
        <td>
      <input type="hidden" id="version_dirty">
      <select id="version" name="version">
          <option value="GM">GM
          </option>
          <option value="Maintenance Update">Maintenance Update
          </option>
          <option value="Milestone 8" selected>Milestone 8
          </option>
      </select>
  </td>
    </tr>
        
    
        
    <tr><th class="field_label "
    id="field_label_rep_platform">

    <label for="rep_platform" accesskey="h">

  <a 
      title="The hardware platform the bug was observed on. Note: When searching, selecting the option &quot;All&quot; only finds bugs whose value for this field is literally the word &quot;All&quot;."
      class="field_help_link"
      href="page.cgi?id=glossary.html#rep_platform"
  >Hardware:</a>
</label>
</th>
      <td class="field_value"><input type="hidden" id="rep_platform_dirty">
        <select id="rep_platform" 
                name="rep_platform" 
                >
            <option value="All"
                    id="v1_rep_platform"
              >All</option>
            <option value="AXP"
                    id="v2_rep_platform"
              >AXP</option>
            <option value="aarch64"
                    id="v32_rep_platform"
              >aarch64</option>
            <option value="Other"
                    id="v26_rep_platform"
              
                selected="selected">Other</option>
        </select>
        

        <script type="text/javascript">
        <!--
          initHidingOptionsForIE('rep_platform');
          
        //-->
        </script>
       <input type="hidden" id="op_sys_dirty">
        <select id="op_sys" 
                name="op_sys" 
                >
            <option value="All"
                    id="v1_op_sys"
              >All</option>
            <option value="Frobware 5.1"
                    id="v2_op_sys"
              >FrobWare 5.1</option>
        </select>
        

        <script type="text/javascript">
        <!--
          initHidingOptionsForIE('op_sys');
          
        //-->
        </script>
      </td>
    </tr>
          <tr>
    <td colspan="2" class="bz_section_spacer"></td>
  </tr>
          
          <tr>
      <th class="field_label">
        <label for="priority">
          <a href="page.cgi?id=glossary.html#priority">Priority</a></label>:
      </th>
      <td><input type="hidden" id="priority_dirty">
        <select id="priority" 
                name="priority" 
                >
            <option value="P0 - Crit Sit"
                    id="v1_priority"
              >P0 - Crit Sit</option>
            <option value="P1 - Urgent"
                    id="v2_priority"
              >P1 - Urgent</option>
            <option value="P2 - High"
                    id="v3_priority"
              >P2 - High</option>
            <option value="P3 - Medium"
                    id="v4_priority"
              
                selected="selected">P3 - Medium</option>
            <option value="P4 - Low"
                    id="v5_priority"
              >P4 - Low</option>
            <option value="P5 - None"
                    id="v6_priority"
              >P5 - None</option>
        </select>
        

        <script type="text/javascript">
        <!--
          initHidingOptionsForIE('priority');
          
        //-->
        </script>
        <label for="severity">
          <b>Severity</b></label>:
       <input type="hidden" id="bug_severity_dirty">
        <select id="bug_severity" 
                name="bug_severity" 
                >
            <option value="Critical"
                    id="v2_bug_severity"
              >Critical</option>
            <option value="Major"
                    id="v3_bug_severity"
              >Major</option>
            <option value="Normal"
                    id="v4_bug_severity"
              
                selected="selected">Normal</option>
            <option value="Minor"
                    id="v5_bug_severity"
              >Minor</option>
            <option value="Enhancement"
                    id="v6_bug_severity"
              >Enhancement</option>
        </select>
        

        <script type="text/javascript">
        <!--
          initHidingOptionsForIE('bug_severity');
          
        //-->
        </script>
      </td>
    </tr>

      <tr>
        <th class="field_label">
          <label for="target_milestone">
            <a href="page.cgi?id=glossary.html#target_milestone">
            Target&nbsp;Milestone</a></label>:
        </th><td>
      <input type="hidden" id="target_milestone_dirty">
      <select id="target_milestone" name="target_milestone">
          <option value="---" selected>---
          </option>
          <option value="GM">GM
          </option>
          <option value="Maintenance Update">Maintenance Update
          </option>
      </select>
  </td>
      </tr>            
          
          <tr>
      <th class="field_label">
        <a href="page.cgi?id=glossary.html#assigned_to">Assigned To</a>:
      </th>
      <td>
          <div id="bz_assignee_edit_container" class="bz_default_hidden">
            <span><span class="vcard"><a class="email" href="mailto:user&#64;foobar.com" title="Firstname Lastname &lt;user&#64;foobar.com&gt;"> <span class="fn">Firstname Lastname</span></a>
</span>
              (<a href="#" id="bz_assignee_edit_action">edit</a>)
            </span>
          </div>
          <div id="bz_assignee_input"><input
    name="assigned_to"
    value="user&#64;foobar.com" class="bz_userfield"  size="30"  id="assigned_to" 
  >
            <br>
            <input type="checkbox" id="set_default_assignee" name="set_default_assignee" value="1">
            <label id="set_default_assignee_label" for="set_default_assignee">Reset Assignee to default</label>
          </div>
          <script type="text/javascript">
           hideEditableField('bz_assignee_edit_container', 
                             'bz_assignee_input', 
                             'bz_assignee_edit_action', 
                             'assigned_to', 
                             'user\x40foobar.com' );
           hideEditableField('bz_assignee_edit_container',
                             'bz_assignee_input',
                             'bz_assignee_take_action',
                             'assigned_to',
                             'user\x40foobar.com',
                             'user\x40foobar.com' );
           initDefaultCheckbox('assignee');                  
          </script>
      </td>
    </tr>

    <tr><th class="field_label "
    id="field_label_qa_contact">

    <label for="qa_contact" accesskey="q">

  <a 
      title="The person responsible for confirming this bug if it is unconfirmed, and for verifying the fix once the bug has been resolved."
      class="field_help_link"
      href="page.cgi?id=glossary.html#qa_contact"
  >QA Contact:</a>
</label>
</th>
      <td>
          <div id="bz_qa_contact_edit_container" class="bz_default_hidden">
            <span><span class="vcard"><a class="email" href="mailto:user&#64;foobar.com" title="Firstname Lastname &lt;user&#64;foobar.com&gt;"> <span class="fn">Firstname Lastname</span></a>
</span>
              (<a href="#" id="bz_qa_contact_edit_action">edit</a>)
            </span>
          </div>
          <div id="bz_qa_contact_input"><input
    name="qa_contact"
    value="user&#64;foobar.com" class="bz_userfield"  size="30"  id="qa_contact" 
  >
            <br>
            <input type="checkbox" id="set_default_qa_contact" name="set_default_qa_contact" value="1">
            <label for="set_default_qa_contact" id="set_default_qa_contact_label">Reset QA Contact to default</label>
          </div>
          <script type="text/javascript">
            hideEditableField('bz_qa_contact_edit_container', 
                              'bz_qa_contact_input', 
                              'bz_qa_contact_edit_action', 
                              'qa_contact', 
                              'user\x40foobar.com');
            hideEditableField('bz_qa_contact_edit_container', 
                              'bz_qa_contact_input', 
                              'bz_qa_contact_take_action', 
                              'qa_contact', 
                              'user\x40foobar.com',
                              'user\x40foobar.com');
            initDefaultCheckbox('qa_contact');
          </script>
      </td>
    </tr>
    <script type="text/javascript">
      assignToDefaultOnChange(['product', 'component'],
        'cloud-bugs\x40foobar.com',
        'cloud-bugs\x40foobar.com');
    </script>
          <tr>
    <td colspan="2" class="bz_section_spacer"></td>
  </tr>
          <tr><th class="field_label "
    id="field_label_bug_file_loc">

    <label for="bug_file_loc" accesskey="u">

  <a 
      title="Bugs can have a URL associated with them - for example, a pointer to a web site where the problem is seen."
      class="field_help_link"
      href="page.cgi?id=glossary.html#bug_file_loc"
  >URL:</a>
</label>
</th>
    <td>
        <span id="bz_url_edit_container" class="bz_default_hidden"> 
        (<a href="#" id="bz_url_edit_action">edit</a>)</span>
      <span id="bz_url_input_area"><input id="bug_file_loc" name="bug_file_loc" class="text_input"
              value="" size="40">
      </span>
        <script type="text/javascript">
          hideEditableField('bz_url_edit_container', 
                            'bz_url_input_area', 
                            'bz_url_edit_action', 
                            'bug_file_loc', 
                            "");
        </script>
    </td>
  </tr>
  
    <tr><th class="field_label "
    id="field_label_status_whiteboard">

    <label for="status_whiteboard" accesskey="w">

  <a 
      title="Each bug has a free-form single line text entry box for adding tags and status information."
      class="field_help_link"
      href="page.cgi?id=glossary.html#status_whiteboard"
  >Whiteboard:</a>
</label>
</th><td colspan="2">
       <input id="status_whiteboard" name="status_whiteboard" class="text_input"
              value="wasZZ:48626  zzz     openZZ:54027" size="40">  
  </td>
    </tr>
  
    <tr>
      <th class="field_label">
        <label for="keywords" accesskey="k">
          <a href="describekeywords.cgi"><u>K</u>eywords</a></label>:
      </th>
      <td class="field_value" colspan="2"><div id="keywords_container">
         <input type="text" id="keywords" size="40"
                class="text_input" name="keywords"
                value="DSLA_REQUIRED, DSLA_SOLUTION_PROVIDED">
         <div id="keywords_autocomplete"></div>
       </div>
      </td>
    </tr>

    <tr><th class="field_label "
    id="field_label_tag">

    <label for="tag">

  <a 
      title="Unlike Keywords which are global and visible by all users, Tags are personal and can only be viewed and edited by their author. Editing them won't send any notification to other users. Use them to tag and keep track of bugs."
      class="field_help_link"
      href="page.cgi?id=glossary.html#tag"
  >Tags:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_tag" >
       <div id="tag_container">
         <input type="text" id="tag" size="40"
                class="text_input" name="tag"
                value="">
         <div id="tag_autocomplete"></div>
       </div>
       </td>
    </tr>
          <tr>
    <td colspan="2" class="bz_section_spacer"></td>
  </tr>

          
<tr><th class="field_label "
    id="field_label_dependson">


  <a 
      title="The bugs listed here must be resolved before this bug can be resolved."
      class="field_help_link"
      href="page.cgi?id=glossary.html#dependson"
  >Depends on:</a>

</th>

  <td>
    <span id="dependson_input_area">
        <input name="dependson" 
               id="dependson" class="text_input"
               value="">
    </span>
    
      <span id="dependson_edit_container" 
            class="edit_me bz_default_hidden">
        (<a href="#" id="dependson_edit_action">edit</a>)
      </span>
      <script type="text/javascript">
        hideEditableField('dependson_edit_container', 
                          'dependson_input_area', 
                          'dependson_edit_action', 
                          'dependson', 
                          '');
      </script>
  </td>
  </tr>
  
  <tr><th class="field_label "
    id="field_label_blocked">


  <a 
      title="This bug must be resolved before the bugs listed in this field can be resolved."
      class="field_help_link"
      href="page.cgi?id=glossary.html#blocked"
  >Blocks:</a>

</th>

  <td>
    <span id="blocked_input_area">
        <input name="blocked" 
               id="blocked" class="text_input"
               value="">
    </span>
    
      <span id="blocked_edit_container" 
            class="edit_me bz_default_hidden">
        (<a href="#" id="blocked_edit_action">edit</a>)
      </span>
      <script type="text/javascript">
        hideEditableField('blocked_edit_container', 
                          'blocked_input_area', 
                          'blocked_edit_action', 
                          'blocked', 
                          '');
      </script>
  </td>
  </tr>
  
  <tr>
    <th>&nbsp;</th>
  
    <td colspan="2" align="left" id="show_dependency_tree_or_graph">
      Show dependency <a href="showdependencytree.cgi?id=1047068&amp;hide_resolved=1">tree</a>
  
        /&nbsp;<a href="showdependencygraph.cgi?id=1047068">graph</a>
    </td>
  </tr>
          
        </table>
      </td>
      <td>
        <div class="bz_column_spacer">&nbsp;</div>
      </td>
      
      <td id="bz_show_bug_column_2" class="bz_show_bug_column">
        <table cellpadding="3" cellspacing="1">
        <tr>
    <th class="field_label">
      Reported:
    </th>
    <td>2017-07-03 13:29 UTC by <span class="vcard"><a class="email" href="mailto:user&#64;foobar.com" title="Firstname Lastname &lt;user&#64;foobar.com&gt;"> <span class="fn">Firstname Lastname</span></a>
</span>
    </td>
  </tr>
  
  <tr>
    <th class="field_label">
      Modified:
    </th>
    <td>2019-03-28 11:40 UTC 
      (<a href="show_activity.cgi?id=1047068">History</a>)
    </td>
  
  </tr>
         <tr>
      <th class="field_label">
        <label for="newcc" accesskey="a">CC List:</label>
      </th>
      <td>2 
          users
            including you
          <span id="cc_edit_area_showhide_container" class="bz_default_hidden">
            (<a href="#" id="cc_edit_area_showhide">edit</a>)
          </span>
        <div id="cc_edit_area">
          <br>
            <div>
              <div><label for="cc"><b>Add</b></label></div><input
    name="newcc"
    value="" class="bz_userfield"  size="30"  id="newcc" 
  >
            </div>
            <select id="cc" multiple="multiple" size="5"name="cc">
                <option value="user&#64;foobar.com">user&#64;foobar.com</option>
                <option value="deuser&#64;gmail.com">deuser&#64;gmail.com</option>
            </select>
              <br>
              <input type="checkbox" id="removecc" name="removecc">
              <label for="removecc">
                  Remove selected CCs
              </label>
              <br>
        </div>
      </td>
    </tr>
         <tr>
    <td colspan="2" class="bz_section_spacer"></td>
  </tr>
<tr><th class="field_label "
    id="field_label_see_also">

    <label for="see_also">

  <a 
      title="This allows you to refer to bugs in other installations. You can enter a URL to a bug in the 'Add Bug URLs' field to note that that bug is related to this one. You can enter multiple URLs at once by separating them with a comma. You should normally use this field to refer to bugs in other installations. For bugs in this installation, it is better to use the Depends on and Blocks fields."
      class="field_help_link"
      href="page.cgi?id=glossary.html#see_also"
  >See Also:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_see_also" >

         <span id="container_showhide_see_also"
               class="bz_default_hidden">
           (<a href="#" id="showhide_see_also">add</a>)
         </span>
         <div id="container_see_also">
           <label for="see_also">
             <strong>Add Bug URLs:</strong>
           </label><br>
           <input type="text" id="see_also" size="40"
                  class="text_input" name="see_also">
         </div>
         <script type="text/javascript">
             setupEditLink('see_also');
         </script></td>
    </tr> 
         <tr><th class="field_label "
    id="field_label_cf_foundby">

    <label for="cf_foundby">

  <a 
      title="A custom Drop Down field in this installation of Bugzilla."
      class="field_help_link"
      href="page.cgi?id=glossary.html#cf_foundby"
  >Found By:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_cf_foundby"  colspan="2">
        <input type="hidden" id="cf_foundby_dirty">
        <select id="cf_foundby" 
                name="cf_foundby" 
                >
            <option value="---"
                    id="v1_cf_foundby"
              >---</option>
            <option value="Beta-Customer"
                    id="v2_cf_foundby"
              >Frob Tester</option>
        </select>
        

        <script type="text/javascript">
        <!--
          initHidingOptionsForIE('cf_foundby');
          
        //-->
        </script>
</td>
    </tr>
    <tr><th class="field_label "
    id="field_label_cf_nts_priority">

    <label for="cf_nts_priority">

  <a 
      title="A custom Free Text field in this installation of Bugzilla."
      class="field_help_link"
      href="https://innerweb.foobar.com/organizations/technical_services/ntsprojects/defects.html"
  >Services Priority:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_cf_nts_priority"  colspan="2">
        <input id="cf_nts_priority" class="text_input"
               name="cf_nts_priority"
               value="" size="40"
               maxlength="255"></td>
    </tr>
    <tr><th class="field_label "
    id="field_label_cf_biz_priority">

    <label for="cf_biz_priority">

  <a 
      title="A custom Free Text field in this installation of Bugzilla."
      class="field_help_link"
      href="page.cgi?id=glossary.html#cf_biz_priority"
  >Business Priority:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_cf_biz_priority"  colspan="2">
        <input id="cf_biz_priority" class="text_input"
               name="cf_biz_priority"
               value="" size="40"
               maxlength="255"></td>
    </tr>
    <tr><th class="field_label "
    id="field_label_cf_blocker">

    <label for="cf_blocker">

  <a 
      title="A custom Drop Down field in this installation of Bugzilla."
      class="field_help_link"
      href="page.cgi?id=glossary.html#cf_blocker"
  >Blocker:</a>
</label>
</th>
  <td class="field_value "
      id="field_container_cf_blocker"  colspan="2">
        <input type="hidden" id="cf_blocker_dirty">
        <select id="cf_blocker" 
                name="cf_blocker" 
                >
            <option value="---"
                    id="v1_cf_blocker"
              
                selected="selected">---</option>
            <option value="No"
                    id="v2_cf_blocker"
              >No</option>
            <option value="Yes"
                    id="v3_cf_blocker"
              >Yes</option>
        </select>
        

        <script type="text/javascript">
        <!--
          initHidingOptionsForIE('cf_blocker');
          
        //-->
        </script>
</td>
    </tr>
    <tr><th class="field_label  bz_hidden_field"
    id="field_label_cf_marketing_qa_status">

    <label for="cf_marketing_qa_status">

  <a 
      title="A custom Drop Down field in this installation of Bugzilla."
      class="field_help_link"
      href="page.cgi?id=glossary.html#cf_marketing_qa_status"
  >Marketing QA Status:</a>
</label>
</th>
  <td class="field_value  bz_hidden_field"
      id="field_container_cf_marketing_qa_status"  colspan="2">
        <input type="hidden" id="cf_marketing_qa_status_dirty">
        <select id="cf_marketing_qa_status" 
                name="cf_marketing_qa_status" 
                >
            <option value="---"
                    id="v1_cf_marketing_qa_status"
              
                selected="selected">---</option>
            <option value="Dev Stage"
                    id="v2_cf_marketing_qa_status"
              >Dev Stage</option>
        </select>
        
</td>
    </tr>
    <tr><th class="field_label  bz_hidden_field"
    id="field_label_cf_it_deployment">

    <label for="cf_it_deployment">

  <a 
      title="A custom Drop Down field in this installation of Bugzilla."
      class="field_help_link"
      href="page.cgi?id=glossary.html#cf_it_deployment"
  >IT Deployment:</a>
</label>
</th>
  <td class="field_value  bz_hidden_field"
      id="field_container_cf_it_deployment"  colspan="2">
        <input type="hidden" id="cf_it_deployment_dirty">
        <select id="cf_it_deployment" 
                name="cf_it_deployment" 
                >
            <option value="---"
                    id="v1_cf_it_deployment"
              
                selected="selected">---</option>
            <option value="Development/Test"
                    id="v2_cf_it_deployment"
              >Development/Test</option>
            <option value="Staging"
                    id="v3_cf_it_deployment"
              >Staging</option>
            <option value="Production"
                    id="v4_cf_it_deployment"
              >Production</option>
        </select>
        
</td>
    </tr>
         <tr>
    <td colspan="2" class="bz_section_spacer"></td>
  </tr>
         <ul>
    <li><a href="tr_new_case.cgi?product=foobar%20frob%20Cloud%207&bug=1047068">Create test case</a></li>
</ul><ul>
  <li><a href="enter_bug.cgi?cloned_bug_id=1047068">Clone This Bug</a></li>
</ul>
                <tr>
      <th class="field_label">
        <label>Flags:</label>
      </th>
      <td><script src="js/flag.js?1411227336" type="text/javascript"></script>

<table id="flags">

  <tbody>
    <tr>
      <td>
          <span title="Firstname Lastname &lt;user&#64;foobar.com&gt;">user</span>:
      </td>
      <td>
        <label title="Set this flag when the bug is in need of additional information" for="flag-201661">needinfo</label>
      </td>
      <td>
        <input type="hidden" id="flag-201661_dirty">
        <select id="flag-201661" name="flag-201661"
                title="Set this flag when the bug is in need of additional information"
                onchange="toggleRequesteeField(this);"
                class="flag_select flag_type-4">
        
          <option value="X"></option>
            <option value="?" selected>?</option>
            <option value="+" >+</option>
            <option value="-" >-</option>
        </select>
      </td>
        <td>
            <span style="white-space: nowrap;"><input
    name="requestee-201661"
    value="user&#64;foobar.com" class="requestee"  id="requestee-201661" 
  >
            </span>
        </td>
    </tr>
  </tbody><tbody>
    <tr>
      <td>
          <span title="Firstname Lastname &lt;user&#64;foobar.com&gt;">user</span>:
      </td>
      <td>
        <label title="Set this flag when the bug is in need of additional information" for="flag-201662">needinfo</label>
      </td>
      <td>
        <input type="hidden" id="flag-201662_dirty">
        <select id="flag-201662" name="flag-201662"
                title="Set this flag when the bug is in need of additional information"
                onchange="toggleRequesteeField(this);"
                class="flag_select flag_type-4">
        
          <option value="X"></option>
            <option value="?" selected>?</option>
            <option value="+" >+</option>
            <option value="-" >-</option>
        </select>
      </td>
        <td>
            <span style="white-space: nowrap;"><input
    name="requestee-201662"
    value="user&#64;foobar.com" class="requestee"  id="requestee-201662" 
  >
            </span>
        </td>
    </tr>
  </tbody>
<tbody>
    <tr>
      <td>
          <span title="Firstname Lastname &lt;user&#64;foobar.com&gt;">user</span>:
      </td>
      <td>
        <label title="blank, means that the bug has not been considered as being a show stopper
?, means that there is a request to evaluate the bug as a ship stopper
-, means that a request to evaluate a bug as a ship stopper resulted in the bug
NOT being identified as a ship stopper
+, means that a request to evaluate a bug as a ship stopper DID result in the
bug being identified as a ship stopper" for="flag-201663">SHIP_STOPPER</label>
      </td>
      <td>
        <input type="hidden" id="flag-201663_dirty">
        <select id="flag-201663" name="flag-201663"
                title="blank, means that the bug has not been considered as being a show stopper
?, means that there is a request to evaluate the bug as a ship stopper
-, means that a request to evaluate a bug as a ship stopper resulted in the bug
NOT being identified as a ship stopper
+, means that a request to evaluate a bug as a ship stopper DID result in the
bug being identified as a ship stopper"
                onchange="toggleRequesteeField(this);"
                class="flag_select flag_type-2">
        
          <option value="X"></option>
            <option value="?" selected>?</option>
            <option value="+" >+</option>
            <option value="-" >-</option>
        </select>
      </td>
        <td>
            <span style="white-space: nowrap;"><input
    name="requestee-201663"
    value="user&#64;foobar.com" class="requestee"  id="requestee-201663" 
  >
            </span>
        </td>
    </tr>
  </tbody>

<tbody class="bz_flag_type">
    <tr>
      <td>
      </td>
      <td>
        <label title="?: needs review
+: reviewed
-: not used" for="flag_type-3">CCB_Review</label>
      </td>
      <td>
        <input type="hidden" id="flag_type-3_dirty">
        <select id="flag_type-3" name="flag_type-3"
                title="?: needs review
+: reviewed
-: not used"
                onchange="toggleRequesteeField(this);"
                class="flag_select flag_type-3">
        
          <option value="X"></option>
            <option value="?" >?</option>
            <option value="+" >+</option>
            <option value="-" >-</option>
        </select>
      </td>
        <td>
        </td>
    </tr>
  </tbody>

  
      <tbody class="bz_flag_type">
        <tr><td colspan="3"><hr></td></tr>
      </tbody><tbody class="bz_flag_type">
    <tr>
      <td>addl.
      </td>
      <td>
        <label title="Set this flag when the bug is in need of additional information" for="flag_type-4">needinfo</label>
      </td>
      <td>
        <input type="hidden" id="flag_type-4_dirty">
        <select id="flag_type-4" name="flag_type-4"
                title="Set this flag when the bug is in need of additional information"
                onchange="toggleRequesteeField(this);"
                class="flag_select flag_type-4">
        
          <option value="X"></option>
            <option value="?" >?</option>
            <option value="+" >+</option>
            <option value="-" >-</option>
        </select>
      </td>
        <td>
            <span style="white-space: nowrap;"><input
    name="requestee_type-4"
    value="" class="requestee"  id="requestee_type-4" 
  >
            </span>
        </td>
    </tr>
  </tbody>
</table>
          <span id="bz_flags_more_container" class="bz_default_hidden">
            (<a href="#" id="bz_flags_more_action">more flags</a>)
          </span>
      </td>
    </tr>

        </table>
      </td>
    </tr>
    <tr>
      <td colspan="3">
          <hr id="bz_top_half_spacer">
      </td>
    </tr>
  </table>

  <table id="bz_big_form_parts" cellspacing="0" cellpadding="0"><tr>
  <td><table class="bz_time_tracking_table">
    <tr><th class="field_label "
    id="field_label_estimated_time">

    <label for="estimated_time">

  <a 
      title="The amount of time that has been estimated it will take to resolve this bug."
      class="field_help_link"
      href="page.cgi?id=glossary.html#estimated_time"
  >Orig. Est.:</a>
</label>
</th>
      <th>
        Current Est.:
      </th><th class="field_label "
    id="field_label_work_time">

    <label for="work_time">

  <a 
      title="The total amount of time spent on this bug so far."
      class="field_help_link"
      href="page.cgi?id=glossary.html#work_time"
  >Hours Worked:</a>
</label>
</th><th class="field_label "
    id="field_label_remaining_time">

    <label for="remaining_time">

  <a 
      title="The number of hours of work left on this bug, calculated by subtracting the Hours Worked from the Orig. Est.."
      class="field_help_link"
      href="page.cgi?id=glossary.html#remaining_time"
  >Hours Left:</a>
</label>
</th><th class="field_label "
    id="field_label_percentage_complete">

    <label for="percentage_complete">

  <a 
      title="How close to 100% done this bug is, by comparing its Hours Worked to its Orig. Est.."
      class="field_help_link"
      href="page.cgi?id=glossary.html#percentage_complete"
  >%Complete:</a>
</label>
</th>
      <th>
        Gain:
      </th><th class="field_label "
    id="field_label_deadline">

    <label for="deadline">

  <a 
      title="The date that this bug must be resolved by, entered in YYYY-MM-DD format."
      class="field_help_link"
      href="page.cgi?id=glossary.html#deadline"
  >Deadline:</a>
</label>
</th>
    </tr>
    <tr>
      <td>
        <input name="estimated_time" id="estimated_time"
               value="0.0"
               size="6" maxlength="6">
      </td>
      <td>0.0
      </td>
      <td>0.0 +
        <input name="work_time" id="work_time"
               value="0" size="3" maxlength="6"
               onchange="adjustRemainingTime();">
      </td>
      <td>
        <input name="remaining_time" id="remaining_time"
               value="0.0"
               size="6" maxlength="6" onchange="updateRemainingTime();">
      </td>
      <td>0
      </td>
      <td>0.0
      </td>
       <td><input name="deadline" size="20"
             id="deadline"
             value=""
             onchange="updateCalendarFromField(this)">
      <button type="button" class="calendar_button"
              id="button_calendar_deadline"
              onclick="showCalendar('deadline')">
        <span>Calendar</span>
      </button>

      <div id="con_calendar_deadline"></div>

      <script type="text/javascript">
        <!--
          createCalendar('deadline');
        //--></script>
      </td>        
    </tr>
    <tr>
      <td colspan="7" class="bz_summarize_time">
        <a href="summarize_time.cgi?id=1047068&amp;do_depends=1">
        Summarize time (including time for bugs
        blocking this bug)</a>
      </td>
    </tr>
  </table>
<br>
<table id="attachment_table" cellspacing="0" cellpadding="4">
  <tr id="a0">
    <th colspan="2" align="left">
      Attachments
    </th>
  </tr>


      <tr id="a1" class="bz_contenttype_text_plain">
        <td valign="top">
            <a href="attachment.cgi?id=766283"
               title="View the content of the attachment">
          <b>description</b></a>

          <span class="bz_attach_extra_info">
              (2 bytes,
                text/plain)

            <br>
            <a href="#attach_766283"
               title="Go to the comment associated with the attachment">2018-04-06 12:48 UTC</a>,

            <span class="vcard"><a class="email" href="mailto:user&#64;foobar.com" title="Firstname Lastname &lt;user&#64;foobar.com&gt;"> <span class="fn">Firstname Lastname</span></a>
</span>
          </span>
        </td>


        <td valign="top">
          <a href="attachment.cgi?id=766283&amp;action=edit">Details</a>
        </td>
      </tr>
  <tr class="bz_attach_footer">
    <td colspan="2">
        <span class="bz_attach_view_hide">
            <a id="view_all" href="attachment.cgi?bugid=1047068&amp;action=viewall">View All</a>
        </span>
        <a href="attachment.cgi?bugid=1047068&amp;action=enter">Add an attachment</a>
        (proposed patch, testcase, etc.)
    </td>
  </tr>
</table>
<br>

  </td>
  <td><div class="bz_group_visibility_section">



          <div id="bz_restrict_group_visibility_help">
            <b>Only users in all of the selected groups can view this 
              bug:</b>
             <p class="instructions">
               Unchecking all boxes makes this a more user accessible bug.
             </p>
          </div>

        <input type="hidden" name="defined_groups" 
               value="foobaronly">

      <input type="checkbox" value="foobaronly"
             name="groups" id="group_10" checked="checked">
      <label for="group_10">This bug is only open to employees and contractors</label>
      <br>

        <img src="images/padlock.png" /><em> This bug is open to foobar Enterprise Partners</em><br />


      <div id="bz_enable_role_visibility_help">
        <b>Users in the roles selected below can always view 
          this bug:</b>
      </div>
      <div id="bz_enable_role_visibility">
        <div>
            <input type="hidden" name="defined_reporter_accessible" value="1">
          <input type="checkbox" value="1"
                 name="reporter_accessible" id="reporter_accessible">
          <label for="reporter_accessible">Reporter</label>
        </div>
        <div>
            <input type="hidden" name="defined_cclist_accessible" value="1">
          <input type="checkbox" value="1"
                 name="cclist_accessible" id="cclist_accessible">
          <label for="cclist_accessible">CC List</label>
        </div>
        <p class="instructions">
          The assignee
             and QA contact
          can always see a bug, and this section does not
          take effect unless the bug is restricted to at
          least one group.
        </p>
      </div>
  </div>
  </td>
  </tr></table>

  
  <div id="comments">

<!-- This auto-sizes the comments and positions the collapse/expand links 
     to the right. -->
<table class="bz_comment_table" cellpadding="0" cellspacing="0"><tr>
<td>
<div id="c0" class="bz_comment bz_first_comment">

      <div class="bz_first_comment_head">

          <span class="bz_comment_actions">
              [<a class="bz_reply_link" href="#add_comment"
                  onclick="replyToComment('0', '7315202', 'Firstname Lastname'); return false;"
              >reply</a>]
          </span>

          <div class="bz_private_checkbox">
            <input type="hidden" value="1"
                   name="defined_isprivate_7315202">
            <input type="checkbox"
                   name="isprivate_7315202" value="1"
                   id="isprivate_7315202"
                   onClick="updateCommentPrivacy(this, 0)">
            <label for="isprivate_7315202">Private</label>
          </div>

        <span class="bz_comment_number">
          <a 
             href="show_bug.cgi?id=1047068#c0">Description</a>
        </span>

        <span class="bz_comment_user">
          <span class="vcard"><a class="email" href="mailto:user&#64;foobar.com" title="Firstname Lastname &lt;user&#64;foobar.com&gt;"> <span class="fn">Firstname Lastname</span></a>
</span>
        </span>

        <span class="bz_comment_time">
          2017-07-03 13:29:15 UTC
        </span>
      </div>



<pre class="bz_comment_text"  id="comment_text_0">This is a test cloud treta.</pre>
    </div><div id="c1" class="bz_comment bz_private">

      <div class="bz_comment_head">

          <span class="bz_comment_actions">
              [<a class="bz_reply_link" href="#add_comment"
                  onclick="replyToComment('1', '7315205', 'ZZ treta Coordination'); return false;"
              >reply</a>]
            <script type="text/javascript"><!--
              addCollapseLink(1, 'Toggle comment display'); // -->
            </script>
          </span>

          <div class="bz_private_checkbox">
            <input type="hidden" value="1"
                   name="defined_isprivate_7315205">
            <input type="checkbox"
                   name="isprivate_7315205" value="1"
                   id="isprivate_7315205"
                   onClick="updateCommentPrivacy(this, 1)" checked="checked">
            <label for="isprivate_7315205">Private</label>
          </div>

        <span class="bz_comment_number">
          <a 
             href="show_bug.cgi?id=1047068#c1">Comment 1</a>
        </span>

        <span class="bz_comment_user">
          <span class="vcard"><a class="email" href="mailto:supporters&#64;foobar.com" title="ZZ treta Coordination &lt;supporters&#64;foobar.com&gt;"> <span class="fn">ZZ treta Coordination</span></a>
</span>
        </span>

        <span class="bz_comment_time">
          2017-07-03 13:31:23 UTC
        </span>
      </div>



<pre class="bz_comment_text"  id="comment_text_1">ZZ:48626 is now handled by Firstname Lastname.</pre>
    </div><div id="c2" class="bz_comment">

      <div class="bz_comment_head">

          <span class="bz_comment_actions">
              [<a class="bz_reply_link" href="#add_comment"
                  onclick="replyToComment('2', '7323867', 'Firstname Lastname'); return false;"
              >reply</a>]
            <script type="text/javascript"><!--
              addCollapseLink(2, 'Toggle comment display'); // -->
            </script>
          </span>

          <div class="bz_private_checkbox">
            <input type="hidden" value="1"
                   name="defined_isprivate_7323867">
            <input type="checkbox"
                   name="isprivate_7323867" value="1"
                   id="isprivate_7323867"
                   onClick="updateCommentPrivacy(this, 2)">
            <label for="isprivate_7323867">Private</label>
          </div>

        <span class="bz_comment_number">
          <a 
             href="show_bug.cgi?id=1047068#c2">Comment 2</a>
        </span>

        <span class="bz_comment_user">
          <span class="vcard"><a class="email" href="mailto:user&#64;foobar.com" title="Firstname Lastname &lt;user&#64;foobar.com&gt;"> <span class="fn">Firstname Lastname</span></a>
</span>
        </span>

        <span class="bz_comment_time">
          2017-07-11 10:50:17 UTC
        </span>
      </div>



<pre class="bz_comment_text"  id="comment_text_2">This is a multi line comment.

The purpose is to test how far it can go with multiline handling. Word Word Word Word Word Word Word Word Word Word Word Word Word Word. 

Done.</pre>
    </div><div id="c3" class="bz_comment">

      <div class="bz_comment_head">

          <span class="bz_comment_actions">
              [<a class="bz_reply_link" href="#add_comment"
                  onclick="replyToComment('3', '7323889', 'Firstname Lastname'); return false;"
              >reply</a>]
            <script type="text/javascript"><!--
              addCollapseLink(3, 'Toggle comment display'); // -->
            </script>
          </span>

          <div class="bz_private_checkbox">
            <input type="hidden" value="1"
                   name="defined_isprivate_7323889">
            <input type="checkbox"
                   name="isprivate_7323889" value="1"
                   id="isprivate_7323889"
                   onClick="updateCommentPrivacy(this, 3)">
            <label for="isprivate_7323889">Private</label>
          </div>

        <span class="bz_comment_number">
          <a 
             href="show_bug.cgi?id=1047068#c3">Comment 3</a>
        </span>

        <span class="bz_comment_user">
          <span class="vcard"><a class="email" href="mailto:user&#64;foobar.com" title="Firstname Lastname &lt;user&#64;foobar.com&gt;"> <span class="fn">Firstname Lastname</span></a>
</span>
        </span>

        <span class="bz_comment_time">
          2017-07-11 11:00:00 UTC
        </span>
      </div>


<pre class="bz_comment_text"  id="comment_text_59">(In reply to <a href="mailto:user&#64;foobar.com">user&#64;foobar.com</a> from <a href="show_bug.cgi?id=1047068#c58">comment 58</a>)
<span class="quote">&gt; comment</span >

comment and reply -- reply takes priority</pre>
    </div><div id="c60" class="bz_comment bz_private">

      <div class="bz_comment_head">

          <span class="bz_comment_actions">
              [<a class="bz_reply_link" href="#add_comment"
                  onclick="replyToComment('60', '8092390', 'ZZ treta Coordination'); return false;"
              >reply</a>]
            <script type="text/javascript"><!--
              addCollapseLink(60, 'Toggle comment display'); // -->
            </script>
          </span>

          <div class="bz_private_checkbox">
            <input type="hidden" value="1"
                   name="defined_isprivate_8092390">
            <input type="checkbox"
                   name="isprivate_8092390" value="1"
                   id="isprivate_8092390"
                   onClick="updateCommentPrivacy(this, 60)" checked="checked">
            <label for="isprivate_8092390">Private</label>
          </div>

        <span class="bz_comment_number">
          <a 
             href="show_bug.cgi?id=1047068#c60">Comment 60</a>
        </span>

        <span class="bz_comment_user">
          <span class="vcard"><a class="email" href="mailto:supporters&#64;foobar.com" title="ZZ treta Coordination &lt;supporters&#64;foobar.com&gt;"> <span class="fn">ZZ treta Coordination</span></a>
</span>
        </span>

        <span class="bz_comment_time">
          2019-03-28 10:23:44 UTC
        </span>
      </div>



<pre class="bz_comment_text"  id="comment_text_60">ZZ:54027 is now handled by Firstname Lastname.</pre>
    </div>


  

</td>
<td>
    <ul class="bz_collapse_expand_comments">
      <li><a href="#" onclick="toggle_all_comments('collapse'); 
                               return false;">Collapse All Comments</a></li>
      <li><a href="#" onclick="toggle_all_comments('expand');
                               return false;">Expand All Comments</a></li>
        <li class="bz_add_comment"><a href="#" 
            onclick="return goto_add_comments('bug_status_bottom');">
            Add Comment</a></li>                               
    </ul>
</td>
</tr></table>
  </div>

    <hr><div id="add_comment" class="bz_section_additional_comments">
      <label for="comment" accesskey="c"><b>Additional 
        <u>C</u>omments</b></label>:

        <input type="checkbox" name="comment_is_private" value="1"
               id="newcommentprivacy"
               onClick="updateCommentTagControl(this, 'comment')">
        <label for="newcommentprivacy">
          Make comment private
        </label>

      <!-- This table keeps the submit button aligned with the box. -->
      <table><tr><td><textarea name="comment" id="comment"
            rows="10"
          cols="80"
            onFocus="this.rows=25"></textarea>
            <script>
               updateCommentTagControl(document.getElementByID('newcommentprivacy'), 'comment');
            </script><div id="needinfo_container">
      
      <script>
        var summary_container = document.getElementByID('static_bug_status');
        summary_container.appendChild(document.createTextNode('[NEEDINFO]'));
      </script>
    <table>
      <tr>
          
          <td align="center">
            <input type="checkbox" id="needinfo_override_201661"
                   name="needinfo_override_201661" value="1">
          </td>
          <td>
            <label for="needinfo_override_201661">
              Clear the needinfo request for
              <em>user&#64;foobar.com</em>.
            </label>
          </td>
      </tr>
      <tr>
<!--EXTRANEEDINFO          
          <td align="center">
            <input type="checkbox" id="needinfo_override_201662"
                   name="needinfo_override_201662" value="1">
          </td>
          <td>
            <label for="needinfo_override_201662">
              Clear the needinfo request for
              <em>user&#64;foobar.com</em>.
            </label>
          </td>
EXTRANEEDINFO-->
      </tr>
      <tr>
        <td align="center">
          <input type="checkbox" name="needinfo" value="1" id="needinfo" onchange="needinfo_focus()">
        </td>
        <td>
          <label for="needinfo">Need more information from</label>
          <select name="needinfo_role" id="needinfo_role" onchange="needinfo_role_changed()">
            <option value="other">other</option>
            <option value="reporter">reporter</option>
            <option value="assigned_to">assignee</option>
              <option value="qa_contact">qa contact</option>
            <option value="">anyone</option>
          </select>
          <span id="needinfo_from_container"><input
    name="needinfo_from"
    value="" onchange="needinfo_other_changed()"  title="Enter one or more comma separated users to request more information from"  size="30"  id="needinfo_from" 
  >
          </span>
          <span id="needinfo_role_identity"></span>
        </td>
      </tr>
    </table>
  </div>
        <br><div class="knob-buttons">
      <input type="submit" value="Save Changes" 
             id="commit">
    </div>

        <table id="bug_status_bottom"
               class="status" cellspacing="0" cellpadding="0">
          <tr>
            <th class="field_label">
              <a href="page.cgi?id=status_resolution_matrix.html">Status</a>:
            </th>
            <td><div id="status"><input type="hidden" id="bug_status_dirty">
        <select id="bug_status" 
                name="bug_status" 
                >
            <option value="REOPENED"
                    id="v5_bug_status"
              >REOPENED</option>
            <option value="RESOLVED"
                    id="v6_bug_status"
              
                selected="selected">RESOLVED</option>
            <option value="VERIFIED"
                    id="v7_bug_status"
              >VERIFIED</option>
        </select>
        

        <script type="text/javascript">
        <!--
          initHidingOptionsForIE('bug_status');
          
        //-->
        </script>

    <noscript><br>resolved&nbsp;as&nbsp;</noscript>

  <span id="resolution_settings"><input type="hidden" id="resolution_dirty">
        <select id="resolution" 
                name="resolution" 
                >
            <option value="FIXED"
                    id="v2_resolution"
              
                selected="selected">FIXED</option>
            <option value="INVALID"
                    id="v3_resolution"
              >INVALID</option>
            <option value="WONTFIX"
                    id="v4_resolution"
              >WONTFIX</option>
            <option value="NORESPONSE"
                    id="v10_resolution"
              >NORESPONSE</option>
            <option value="UPSTREAM"
                    id="v12_resolution"
              >UPSTREAM</option>
            <option value="FEATURE"
                    id="v11_resolution"
              >FEATURE</option>
            <option value="DUPLICATE"
                    id="v7_resolution"
              >DUPLICATE</option>
            <option value="WORKSFORME"
                    id="v8_resolution"
              >WORKSFORME</option>
            <option value="MOVED"
                    id="v9_resolution"
              >MOVED</option>
        </select>
        

        <script type="text/javascript">
        <!--
          initHidingOptionsForIE('resolution');
          
        //-->
        </script>
  </span>

    <noscript><br> duplicate</noscript>
    <span id="duplicate_settings">of
      <span id="dup_id_container" class="bz_default_hidden">bug 
        (<a href="#" id="dup_id_edit_action">edit</a>)
      </span
      ><input id="dup_id" name="dup_id" size="6"
              value="">
    </span>
    <div id="dup_id_discoverable" class="bz_default_hidden">
      <a href="#" id="dup_id_discoverable_action">Mark as Duplicate</a>
    </div>
</div>

            </td>
          </tr>
        </table>
      </td></tr></table>

    
  </div>        

</form>

<hr>
<ul class="related_actions">
    <li><a href="show_bug.cgi?format=multiple&amp;id=1047068">Format For Printing</a></li>
    <li>&nbsp;-&nbsp;<a href="show_bug.cgi?ctype=xml&amp;id=1047068">XML</a></li>
    <li>&nbsp;-&nbsp;<a href="enter_bug.cgi?cloned_bug_id=1047068">Clone This Bug</a></li>
    
    <li>&nbsp;-&nbsp;<a href="#">Top of page </a></li>
    </ul>        


<div class="navigation">
  
  <i><font color="#777777">First</font></i>
  <i><font color="#777777">Last</font></i>
  <i><font color="#777777">Prev</font></i>
  <i><font color="#777777">Next</font></i>
  &nbsp;&nbsp;
  <i><font color="#777777">This bug is not in your last
    search results.</font></i>
</div>

<br>
</div>



<div id="footer">
  <div class="intro"></div>




<ul id="useful-links">
  <li id="links-actions"><ul class="links">
  <li><a href="./">Home</a></li>
  <li><span class="separator">| </span><a href="enter_bug.cgi">New</a></li>
  <li><span class="separator">| </span><a href="describecomponents.cgi">Browse</a></li>
  <li><span class="separator">| </span><a href="query.cgi">Search</a></li>

  <li class="form">
    <span class="separator">| </span>
    <form action="buglist.cgi" method="get"
        onsubmit="if (this.quicksearch.value == '')
                  { alert('Please enter one or more search terms first.');
                    return false; } return true;">
    <input type="hidden" id="no_redirect_bottom" name="no_redirect" value="0">
    <script type="text/javascript">
      if (history && history.replaceState) {
        var no_redirect = document.getElementByID("no_redirect_bottom");
        no_redirect.value = 1;
      }
    </script>
    <input class="txt" type="text" id="quicksearch_bottom" name="quicksearch" 
           title="Quick Search" value="">
    <input class="btn" type="submit" value="Search" 
           id="find_bottom"></form>
  <a href="page.cgi?id=quicksearch.html" title="Quicksearch Help">[?]</a></li>

  <li><span class="separator">| </span><a href="report.cgi">Reports</a></li>

  <li>
      <span class="separator">| </span></li>


<li>
        <span class="separator">| </span>
        <a href="docs/en/html/bug_page.html" target="_blank">Help</a>
      </li>

    <li>
      <span class="separator">| </span></li>
</ul>
  </li>

  
    

  <div class="outro"></div>
</div>


</body>
</html>
`

var changesSubmitted = `
<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd"><html lang="en"><head>
    <title>Bug 1047068
    processed</title>
      <meta http-equiv="Content-Type" content="text/html; charset=UTF-8"/>
  </head>
  <body>
<dl>
  <dt>Changes submitted for <a class="bz_bug_link
          bz_status_RESOLVED  bz_closed" title="RESOLVED FIXED - L3: L4: test cloud bug" href="show_bug.cgi?id=1047068">bug 1047068</a></dt>
  <dd><dl><dt>Email sent to:</dt>
  <dd>
      no one
  </dd>
<dt>Excluding:</dt>
  <dd>
        <codeuser@gmail.com</code>,
        <code>user@foobar.com</code>
  </dd>
</dl>
</body></html>
`

func (cs *clientSuite) TestUpdateNeedinfo(c *C) {
	queries := make(chan url.Values, 10)
	showBug := make(chan string, 10)
	processBug := make(chan string, 10)
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/show_bug.cgi":
			io.WriteString(w, <-showBug)
		case "/process_bug.cgi":
			r.ParseForm()
			query := r.Form
			queries <- query
			io.WriteString(w, <-processBug)
		default:
			http.Error(w, "Unimplemented", 500)
			return
		}
	}))
	defer ts0.Close()
	bz := makeClient(ts0.URL)

	email := "user@foobar.com"
	changes := bugzilla.Changes{SetNeedinfo: email}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err := bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query := <-queries
	c.Assert(query.Get("needinfo"), Equals, "1")
	c.Assert(query.Get("needinfo_role"), Equals, "other")
	c.Assert(query.Get("needinfo_from"), Equals, email)

	comment := "this is a comment"
	changes = bugzilla.Changes{AddComment: comment}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("comment"), Equals, comment)
	c.Assert(query.Get("comment_is_private"), Not(Equals), "1")
	c.Assert(query.Get("commentprivacy"), Not(Equals), "1")

	comment = "this is a comment2"
	changes = bugzilla.Changes{AddComment: comment, CommentIsPrivate: true}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("comment"), Equals, comment)
	c.Assert(query.Get("comment_is_private"), Equals, "1")
	c.Assert(query.Get("commentprivacy"), Equals, "1")

	changes = bugzilla.Changes{ClearNeedinfo: true}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("needinfo_override_201661"), Equals, "1")

	changes = bugzilla.Changes{ClearNeedinfo: true, ClearAllNeedinfos: false}
	showBug <- strings.Replace(strings.Replace(showBugHtml, "<!--EXTRANEEDINFO", "", -1), "EXTRANEEDINFO-->", "", -1)
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, ".*More than one needinfo found.*")

	changes = bugzilla.Changes{ClearNeedinfo: true, ClearAllNeedinfos: true}
	showBug <- strings.Replace(strings.Replace(showBugHtml, "<!--EXTRANEEDINFO", "", -1), "EXTRANEEDINFO-->", "", -1)
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("needinfo_override_201661"), Equals, "1")
	c.Assert(query.Get("needinfo_override_201662"), Equals, "1")

	// special case: muiltiple needinfos for the same user
	// the last one found is picked
	email = "user@foobar.com"
	changes = bugzilla.Changes{RemoveNeedinfo: email}
	showBug <- strings.Replace(strings.Replace(showBugHtml, "<!--EXTRANEEDINFO", "", -1), "EXTRANEEDINFO-->", "", -1)
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("needinfo_override_201661"), Equals, "")
	c.Assert(query.Get("needinfo_override_201662"), Equals, "1")

	email = "user@foobar.com"
	changes = bugzilla.Changes{RemoveNeedinfo: email}
	html := strings.Replace(strings.Replace(showBugHtml, "<!--EXTRANEEDINFO", "", -1), "EXTRANEEDINFO-->", "", -1)
	html = strings.Replace(html, `value="user&#64;foobar.com" class="requestee"  id="requestee-201662"`,
		`value="user2&#64;foobar.com" class="requestee"  id="requestee-201662"`, -1)
	showBug <- html
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("needinfo_override_201661"), Equals, "1")
	c.Assert(query.Get("needinfo_override_201662"), Equals, "")

	url := "http://foobar.com/1/2"
	changes = bugzilla.Changes{SetURL: url}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("bug_file_loc"), Equals, url)

	assignee := "user3@foobar.com"
	changes = bugzilla.Changes{SetAssignee: assignee}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("assigned_to"), Equals, assignee)

	priority := "P0"
	changes = bugzilla.Changes{SetPriority: priority}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("priority"), Equals, "P0 - Crit Sit")

	priority = "wrong"
	changes = bugzilla.Changes{SetPriority: priority}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, ".*invalid priority value.*")

	cc := "newuser@foobar.com"
	changes = bugzilla.Changes{AddCc: cc}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("newcc"), Equals, cc)

	cc = "newuser@foobar.com, another@foobar.com"
	changes = bugzilla.Changes{AddCc: cc}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("newcc"), Equals, cc)

	changes = bugzilla.Changes{CcMyself: true}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("addselfcc"), Equals, "1")

	remove := "user@foobar.com"
	changes = bugzilla.Changes{RemoveCc: remove}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	removed, ok := query["cc"]
	c.Assert(ok, Equals, true)
	c.Assert(removed, DeepEquals, []string{remove})
	c.Assert(query.Get("removecc"), Equals, "1")

	title := "This is the bug title"
	changes = bugzilla.Changes{SetDescription: title}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("short_desc"), Equals, title)

	whiteboard := "This is the whiteboard message"
	changes = bugzilla.Changes{SetWhiteboard: whiteboard}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("status_whiteboard"), Equals, whiteboard)

	status := "REOPENED"
	changes = bugzilla.Changes{SetStatus: status}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("bug_status"), Equals, status)

	resolution := "INVALID"
	changes = bugzilla.Changes{SetResolution: resolution}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("resolution"), Equals, resolution)

	duplicate := 123456
	changes = bugzilla.Changes{SetDuplicate: duplicate}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
	query = <-queries
	c.Assert(query.Get("dup_id"), Equals, fmt.Sprintf("%d", duplicate))

	delta := time.Date(2019, 01, 01, 01, 02, 03, 0, time.UTC)
	changes = bugzilla.Changes{AddComment: "Some comment", DeltaTS: delta, CheckDeltaTS: true}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, ".*collision.*")

	delta = time.Date(2019, 03, 28, 11, 40, 39, 0, time.UTC)
	changes = bugzilla.Changes{AddComment: "Some comment", DeltaTS: delta, CheckDeltaTS: true}
	showBug <- showBugHtml
	processBug <- changesSubmitted
	err = bz.Update(101234, changes)
	c.Assert(err, IsNil)
}

func (cs *clientSuite) TestUnauthorized(c *C) {
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer ts0.Close()
	bz := makeClient(ts0.URL)
	bug, err := bz.GetBug(1047068)
	c.Assert(bug, IsNil)
	c.Assert(err, ErrorMatches, ".*Unauthorized*")
}

var sampleHtmlError = `
<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"
                      "http://www.w3.org/TR/html4/loose.dtd">
<html lang="en">
  <head>
    <title>Invalid Username Or Password</title>

      <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">

  <body onload=""
        class="bugzilla-suse-com yui-skin-sam">
</body>
</html>
`

// TestGetBugNotPermitted triggers the code path when a request is made to
// the wrong endpoint of Bugzilla (one that uses another type of auth)
func (cs *clientSuite) TestGetBugExpectedBugzilla(c *C) {
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, sampleHtmlError, http.StatusOK)
	}))
	defer ts0.Close()
	bz := makeClient(ts0.URL)
	bug, err := bz.GetBug(1047068)
	c.Assert(bug, IsNil)
	c.Assert(err, ErrorMatches, ".*URL or credentials might be incorrect.*")
}

var sampleError = `
<?xml version="1.0" encoding="UTF-8" standalone="yes" ?>
<!DOCTYPE bugzilla SYSTEM "http://bugzilla.foobar.com/page.cgi?id=bugzilla.dtd">

<bugzilla version="4.4.12"
          urlbase="http://bugzilla.foobar.com/"

          maintainer="novbugzilla-dev@forge.provo.foobar.com"
>

    <bug error="NotPermitted">
      <bug_id>1047068</bug_id>
    </bug>

</bugzilla>
`

// TestGetBugNotPermitted triggers the code path when a request is made to
// the wrong endpoint of Bugzilla (one that uses another type of auth)
func (cs *clientSuite) TestGetBugNotPermitted(c *C) {
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, sampleError, http.StatusOK)
	}))
	defer ts0.Close()
	bz := makeClient(ts0.URL)
	bug, err := bz.GetBug(1047068)
	c.Assert(bug, IsNil)
	c.Assert(err, ErrorMatches, ".*NotPermitted*")
}

var sampleJSON = `

{
   "version" : "GMC",
   "short_desc" : "Bug short_desc here",
   "cf_nts_priority" : [
      "",
      ""
   ],
   "bug_severity" : "Major",
   "status_whiteboard" : "openTreta:1234",
   "cclist_accessible" : 1,
   "remaining_time" : "0.00",
   "Comments" : [
      {
         "thetext" : "Some comment",
         "comment_count" : 0,
         "who" : {
            "email" : "username@foobar.com",
            "name" : "Firstname Lastname"
         },
         "commentid" : 8082652,
         "BugWhen" : "2019-03-20T19:48:42Z",
         "isprivate" : 0
      },
      {
         "comment_count" : 1,
         "thetext" : "Second comment",
         "who" : {
            "email" : "user@foobar.com",
            "name" : "Firstname Lastname"
         },
         "isprivate" : 1,
         "BugWhen" : "2019-03-20T20:04:54Z",
         "commentid" : 8082662
      },
      {
         "comment_count" : 2,
         "thetext" : "Third comment.",
         "who" : {
            "name" : "Firstname Lastname",
            "email" : "user@foobar.com"
         },
         "isprivate" : 0,
         "BugWhen" : "2019-03-20T20:05:48Z",
         "commentid" : 8082664
      }
   ],
   "actual_time" : "0.00",
   "creation_ts" : "2019-03-20T19:48:00Z",
   "estimated_time" : "0.00",
   "cc" : [
      "lfirstname@foobar.com",
      "user@foobar.com"
   ],
   "component" : "Core",
   "op_sys" : "FROB 12",
   "rep_platform" : "x86-64",
   "cf_it_deployment" : [
      "---"
   ],
   "delta_ts" : "2019-04-13T13:08:58Z",
   "bug_id" : 1129974,
   "assigned_to" : {
      "email" : "user@foobar.com",
      "name" : "Firstname Lastname"
   },
   "resolution" : "",
   "classification_id" : 27,
   "everconfirmed" : 1,
   "reporter_accessible" : 1,
   "target_milestone" : "---",
   "comment_sort_order" : "oldest_to_newest",
   "cf_blocker" : [
      "---"
   ],
   "cf_foundby" : [
      "---",
      "---"
   ],
   "bug_status" : "IN_PROGRESS",
   "group" : [
      {
         "id" : 10,
         "email" : "foobaronly"
      },
      {
         "id" : 17,
         "email" : "Foo Bar Enterprise"
      }
   ],
   "reporter" : {
      "email" : "user@foobar.com",
      "name" : "Firstname Lastname"
   },
   "classification" : "Frobnicator Plus",
   "cf_biz_priority" : [
      "",
      ""
   ],
   "dup_id" : 0,
   "priority" : "P2 - High",
   "product" : "Frobnicator Plus",
   "token" : [
      "1234555-i2XdlR-90bqTkLrFRr8nC6YbAG1xDNlSktCEp9r3Ux4"
   ],
   "votes" : 0,
   "Attachments" : [
      {
         "filename" : "dump.tbz",
         "attacher" : {
            "name" : "Firstname Lastname",
            "email" : "user@foobar.com"
         },
         "isobsolete" : 0,
         "delta_ts" : "2019-03-20T20:07:16Z",
         "isprivate" : 0,
         "desc" : "the dump",
         "ispatch" : 0,
         "Date" : "2019-03-20T20:07:00Z",
         "type" : "application/x-bzip",
         "token" : "123455555-Y0xsvTLjEFWr3pKrKjnPCsjNupLWQ1SGK7utZkCFqeU",
         "size" : 6014076,
         "attachid" : 800750
      }
   ],
   "qa_contact" : {
      "name" : "E-mail List",
      "email" : "qa-contact@foobar.com"
   },
   "keywords" : "DSLA_REQUIRED",
   "bug_file_loc" : "",
   "flag" : [
      {
         "id" : 201207,
         "status" : "?",
         "type_id" : 4,
         "name" : "needinfo",
         "requestee" : "user@foobar.com",
         "setter" : "supporter@foobar.com"
      }
   ]
}
`

func (cs *clientSuite) TestGetFromJSON(c *C) {
	bz := makeClient("https://bugzilla.anythingworkshere.com")
	bug, err := bz.GetBugFromJSON(strings.NewReader(sampleJSON))
	c.Assert(err, IsNil)
	c.Assert(bug, NotNil)
	c.Check(bug.ShortDesc, Equals, "Bug short_desc here")
	c.Check(bug.AssignedTo.Name, Equals, "Firstname Lastname")
	c.Check(bug.BugSeverity, Equals, "Major")
	c.Check(bug.StatusWhiteboard, Equals, "openTreta:1234")
	c.Check(len(bug.Comments), Equals, 3)
	c.Check(bug.Comments[0].TheText, Equals, "Some comment")
	c.Check(bug.Comments[0].BugWhen, Equals, time.Date(2019, 03, 20, 19, 48, 42, 0, time.UTC))
	c.Check(bug.CreationTS, Equals, time.Date(2019, 03, 20, 19, 48, 0, 0, time.UTC))
	c.Check(bug.Cc, DeepEquals, []string{"lfirstname@foobar.com", "user@foobar.com"})
}

func (cs *clientSuite) TestUpdateConnectionClosed(c *C) {
	var ts0 *httptest.Server
	ts0 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts0.CloseClientConnections()
	}))
	defer ts0.Close()
	bz := makeClient(ts0.URL)
	changes := bugzilla.Changes{AddComment: "Some comment"}
	err := bz.Update(101234, changes)
	c.Assert(err, ErrorMatches, ".*failed to get the update form.*")
}
