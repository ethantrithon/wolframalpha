//This Source Code Form is subject to the terms of the Mozilla Public
//License, v. 2.0. If a copy of the MPL was not distributed with this
//file, You can obtain one at http://mozilla.org/MPL/2.0/.

package wolframalpha

import (
	"errors"
	"regexp"
)

const (
	numericalAnswerProbability = 0.1
	dateAnswerProbability      = 0.3
	numberRegex                = `^-?\d*\.?\d+(×10\^-?\d+)?$`
	finderRegex                = `-?\d*\.?\d+(×10\^-?\d+)?`
	fullResultsURL             = "https://api.wolframalpha.com/v2/query?input=%s&format=plaintext&output=JSON&appid=%s"
	spokenResultsURL           = "https://api.wolframalpha.com/v1/spoken?i=%s&appid=%s"
)

var (
	errInvalidJSON    = errors.New("received JSON was invalid. Should never happen")
	errIsProbablyDate = errors.New("numerical answer is likely to be a date")
	errNoLikely       = errors.New("no answer above numerical likeliness threshold")
	errNoMatch        = errors.New("no number found in likely numerical answer. Should never happen")
	errNoPods         = errors.New("no pods in result, check for errors in the result or the presence of 'DidYouMean's")

	apikey = ""

	keepParens = false

	numericalRgx = regexp.MustCompile(numberRegex)
	numberFinder = regexp.MustCompile(finderRegex)
	parenRemover = regexp.MustCompile(`\(.*\)`).ReplaceAllString
)

//FullResult is returned by a HTTP.GET to the wolframalpha api.
//You should not instantiate this yourself.
type FullResult struct {
	QueryResult *QueryResult `json:"queryresult"`
}

//QueryResult contains all the actual information about a result
//You should not instantiate this yourself.
type QueryResult struct {
	Pods          Pods          `json:"pods"`
	Didyoumeans   DidYouMeans   `json:"didyoumeans"`
	Datatypes     string        `json:"datatypes"`
	Timedout      string        `json:"timedout"`
	Timedoutpods  string        `json:"timedoutpods"`
	Recalculate   string        `json:"recalculate"`
	ID            string        `json:"id"`
	Host          string        `json:"host"`
	Server        string        `json:"server"`
	Related       string        `json:"related"`
	Version       string        `json:"version"`
	Parseidserver string        `json:"parseidserver"`
	Error         *ErrorUnion   `json:"error"`
	Numpods       int           `json:"numpods"`
	Timing        float64       `json:"timing"`
	Parsetiming   float64       `json:"parsetiming"`
	Sources       *SourcesUnion `json:"sources"`
	Tips          *Tips         `json:"tips"`
	Warnings      *Warnings     `json:"warnings"`
	Assumptions   *Assumptions  `json:"assumptions"`
	Success       bool          `json:"success"`
	Parsetimedout bool          `json:"parsetimedout"`
}

//ErrorUnion holds either a *bool value or a *QError (e.g. when a wrong api key
//is used).
//You should not instantiate this yourself.
type ErrorUnion struct {
	Bool  *bool
	Error *QError
}

//QError describes what went wrong in generating a QueryResult from wolframalpha
//You should not instantiate this yourself.
type QError struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

//Pods is a slice of individual pod elements
//You should not instantiate this yourself.
type Pods []Pod

//Pod is one block of general information in the QueryResult. Specifics (what
//you want) are most likely in subpods
//You should not instantiate this yourself.
type Pod struct {
	SubPods         SubPods          `json:"subpods"`
	States          []State          `json:"states"`
	Title           string           `json:"title"`
	Scanner         string           `json:"scanner"`
	ID              string           `json:"id"`
	Position        int              `json:"position"`
	NumSubPods      int              `json:"numsubpods"`
	ExpressionTypes *ExpressionTypes `json:"expressiontypes"`
	Infos           *PodInfos        `json:"infos"`
	Error           bool             `json:"error"`
	Primary         bool             `json:"primary"`
}

//SubPods is a slice of individual SubPod elements
//You should not instantiate this yourself.
type SubPods []SubPod

//SubPod contains the actual information part (i.e. the "answer" to a query).
//You'll most likely just care about the {.Plaintext} of it
//You should not instantiate this yourself.
type SubPod struct {
	Title        string        `json:"title"`
	PlainText    string        `json:"plaintext"`
	MicroSources *MicroSources `json:"microsources,omitempty"`
	ImageSource  string        `json:"imagesource,omitempty"`
	DataSources  *DataSources  `json:"datasources,omitempty"`
	Infos        *SubPodInfos  `json:"infos,omitempty"`
	Primary      bool          `json:"primary,omitempty"`
}

//MicroSources briefly describe where a bit of information came from
//You should not instantiate this yourself.
type MicroSources struct {
	Microsource *SubSource `json:"microsource"`
}

//SubSource is used for various descriptors in SubPods. Contains either a
//single *string or a []string, but not both
//You should not instantiate this yourself.
type SubSource struct {
	String      *string
	StringArray []string
}

//DataSources briefly describe the category of a result
//You should not instantiate this yourself.
type DataSources struct {
	Datasource *SubSource `json:"datasource"`
}

//SubPodInfos contain additional information about the subpod
//You should not instantiate this yourself.
type SubPodInfos struct {
	Links Source `json:"links"`
}

//ExpressionTypes contains either a single *ExpressionType or a
//[]ExpressionType, but not both
//You should not instantiate this yourself.
type ExpressionTypes struct {
	ExpressionType      *ExpressionType
	ExpressionTypeArray []ExpressionType
}

//ExpressionType signals what type of expression the answer is
//You should not instantiate this yourself.
type ExpressionType struct {
	Name string `json:"name"`
}

//State describes one state in a pod
//You should not instantiate this yourself.
type State struct {
	Name       string `json:"name"`
	Input      string `json:"input"`
	Stepbystep bool   `json:"stepbystep,omitempty"`
}

//PodInfos contains general information about a pod
//You should not instantiate this yourself.
type PodInfos struct {
	Units []Unit `json:"units"`
	Text  string `json:"text,omitempty"`
	Links []Link `json:"links"`
}

//Unit describes how an answer is measured, in short and long forms (usually
//mathematical symbol and written out name)
//You should not instantiate this yourself.
type Unit struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}

//Link is an explanation on wolframalpha of an answer
//You should not instantiate this yourself.
type Link struct {
	URL   string `json:"url"`
	Text  string `json:"text"`
	Title string `json:"title"`
}

//SourcesUnion contains either a single *Source or a []Source, but not both
//You should not instantiate this yourself.
type SourcesUnion struct {
	Source      *Source
	SourceArray []Source
}

//Source contains the origins of a piece of data in the response. You should use
//this for attribution/giving credit (if required) in addition to crediting
//wolframalpha themselves
//You should not instantiate this yourself.
type Source struct {
	URL  string `json:"url"`
	Text string `json:"text"`
}

//Tips include tips about how to improve your query if it wasn't understood to
//produce a reasonable result (e.g. "Check your spelling, and use English")
//You should not instantiate this yourself.
type Tips struct {
	Text string `json:"text"`
}

//DidYouMeans are a slice of individual DidYouMean elements
//You should not instantiate this yourself.
type DidYouMeans []DidYouMean

//DidYouMean includes information about possible variants you could have meant
//in your query. Useful for disambiguation purposes
//You should not instantiate this yourself.
type DidYouMean struct {
	Score string `json:"score"`
	Level string `json:"level"`
	Val   string `json:"val"`
}

//Warnings include information about automatic corrections wolframalpha has made
//(e.g. "Freddy Mercury" -> "Freddie Mercury"
//You should not instantiate this yourself.
type Warnings struct {
	Word       string `json:"word"`
	Suggestion string `json:"suggestion"`
	Text       string `json:"text"`
}

//Assumptions include information about how wolframalpha interpreted your query
//if not sure enough to say for certain
//You should not instantiate this yourself.
type Assumptions struct {
	Type     string           `json:"type"`
	Word     string           `json:"word"`
	Template string           `json:"template"`
	Count    int              `json:"count"`
	Values   AssumptionValues `json:"values"`
}

//AssumptionValues is a slice of individual AssumptionValue elements
//You should not instantiate this yourself.
type AssumptionValues []AssumptionValue

//AssumptionValue describes how a particular part of the input was understood
//You should not instantiate this yourself.
type AssumptionValue struct {
	Name  string `json:"name"`
	Word  string `json:"word"`
	Desc  string `json:"desc"`
	Input string `json:"input"`
}
