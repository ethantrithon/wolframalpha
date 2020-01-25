//This Source Code Form is subject to the terms of the Mozilla Public
//License, v. 2.0. If a copy of the MPL was not distributed with this
//file, You can obtain one at http://mozilla.org/MPL/2.0/.

package wolframalpha

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

//APIKey will set the api key used for all queries to k. Intended to only be
//used once
func APIKey(k string) {
	if apikey != "" {
		return
	}
	if k == "" {
		return
	}
	apikey = k
}

//KeepParens sets internal variables to influence how the Get.*Answer functions
//work.
//If parens are ignored (=false) [default], strings from wolframalpha will have
//parentheses removed (e.g. "foo (bar)" will be treated as "foo")
//If parens are kept (=true), strings from wolframalpha are treated and
//processed "as is"
func KeepParens(k bool) {
	keepParens = k
}

//AskQuestionSpokenSync will send the query q to the wolframalpha spoken results
//API and return the result. Will error if the connection fails
func AskQuestionSpokenSync(q string) (r string, e error) {
	resp, e := http.Get(fmt.Sprintf(spokenResultsURL, url.QueryEscape(q), apikey))

	if e != nil {
		return "", e
	}

	defer resp.Body.Close()

	bytes, e := ioutil.ReadAll(resp.Body)
	return string(bytes), e
}

//AskQuestionSpoken will send the answer to query q on the returned channel.
//Will close channel after sending response.
//Will send empty string if the connection fails.
func AskQuestionSpoken(q string) <-chan string {
	r := make(chan string)
	go func() {
		a, e := AskQuestionSpokenSync(q)
		if e != nil {
			r <- ""
			close(r)
			return
		}
		r <- a
		close(r)
	}()
	return r
}

//AskQuestionJSONSync will return a []byte containing the full result in JSON
//format. Will error if the connection fails or the JSON result is somehow
//malformed
func AskQuestionJSONSync(q string) (j []byte, e error) {
	resp, e := http.Get(fmt.Sprintf(fullResultsURL, url.QueryEscape(q), apikey))
	if e != nil {
		return nil, e
	}

	defer resp.Body.Close()

	j, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		return nil, e
	}
	if !json.Valid(j) {
		return nil, errInvalidJSON
	}

	return
}

//AskQuestionJSON will send a []byte containing the full result in JSON format
//to the returned channel. Will close the channel after sending response.
//Will send a nil slice if the connection fails or the JSON result is somehow
//malformed
func AskQuestionJSON(q string) <-chan []byte {
	j := make(chan []byte)
	go func() {
		jsonRes, e := AskQuestionJSONSync(q)
		if e != nil {
			j <- nil
			close(j)
			return
		}
		j <- jsonRes
		close(j)
	}()
	return j
}

//AskQuestionSync will return a *fullResult containing the answer to query q as
//returned from wolframalpha
//Will error if the connection fails or the underlying JSON was malformed
func AskQuestionSync(q string) (r *FullResult, e error) {
	jsonRes, e := AskQuestionJSONSync(q)
	if e != nil {
		return nil, e
	}
	return DecodeJSON(jsonRes)
}

//AskQuestion will send a *fullResult containing the answer to query q on the
//returned channel. Will close channel after sending response.
//Will send a nil result if the connection fails or the underlying JSON was
//malformed
func AskQuestion(q string) <-chan *FullResult {
	r := make(chan *FullResult)
	go func() {
		result, e := AskQuestionSync(q)
		if e != nil {
			r <- nil
			close(r)
			return
		}

		r <- result
		close(r)
	}()
	return r
}

//──────────────────────────────────────────────────────────────────────────────

//DecodeJSON will unmarshal a []byte containing json into a fullResult. Will
//error if the argument is not unmarshalable into a fullResult
func DecodeJSON(j []byte) (r *FullResult, e error) {
	r = &FullResult{}
	e = json.Unmarshal(j, r)
	return
}

//DecodeJSONString will unmarshal a string containing json into a *FullResult.
//Will error if the argument is not unmarshalable into a *FullResult
func DecodeJSONString(j string) (r *FullResult, e error) {
	r = &FullResult{}
	e = json.Unmarshal([]byte(j), r)
	return
}

//RemoveInputInterpretation deletes the pod containing the input interpretation
//from the list of pods in place and returns the *FullResult back
func (f *FullResult) RemoveInputInterpretation() *FullResult {
	f.QueryResult.RemoveInputInterpretation()
	return f
}

//RemoveInputInterpretation deletes the pod containing the input interpretation
//from the list of pods in place and returns the *QueryResult back
func (q *QueryResult) RemoveInputInterpretation() *QueryResult {
	//The input interpretation should be at position 0, so let's assume that it is
	if q.Pods[0].Title == "Input interpretation" {
		q.Pods = q.Pods[1:]
	} else {
		//our assumption failed => doing it the long way
		//walking backwards here to avoid potential out-of-range panics
		for i := len(q.Pods) - 1; i >= 0; i-- {
			if q.Pods[i].Title == "Input interpretation" {
				q.Pods = append(q.Pods[:i], q.Pods[i+1:]...)
			}
		}
	}

	return q
}

//GetAnswer will return either the numerical answer (value and unit combined) -
//if present - or if not, the longest answer instead
//If either fails, GetAnswer will return the error and no other values
func (f *FullResult) GetAnswer() (s string, e error) {
	return f.QueryResult.GetAnswer()
}

//GetAnswer will return either the numerical answer (value and unit combined) -
//if present - or if not, the longest answer instead.
//If either fails, GetAnswer will return the error and no other values
func (q *QueryResult) GetAnswer() (s string, e error) {
	num, unit, err := q.GetNumericalAnswer()
	s = fmt.Sprintf("%s %s", num, unit)

	//predictable and "safe" error. none of our answers were numerical, or we
	//found a date, so we return the longest one (or error)
	noNumberFound := err == errNoLikely || err == errIsProbablyDate
	if noNumberFound {
		//overwrite error here as well, because if it works, then we put <nil> back
		//in error so the check below won't cause issues
		s, err = q.GetLongestAnswer()
	}

	if err != nil {
		return "", err
	}

	return
}

//GetLongestAnswer will return the longest answer contained in any subpod in the
//result.
//Will error if there are no pods
func (f *FullResult) GetLongestAnswer() (s string, e error) {
	return f.QueryResult.GetLongestAnswer()
}

//GetLongestAnswer will return the longest answer contained in any subpod in the
//result.
//Will error if there are no pods
func (q *QueryResult) GetLongestAnswer() (s string, e error) {
	if q.Numpods == 0 {
		return "", errNoPods
	}

	for _, p := range q.Pods {
		for _, sp := range p.SubPods {
			txt := removeParens(sp.PlainText)
			if len(txt) > len(s) {
				s = txt
				continue
			}
		}
	}

	return
}

//GetNumericalAnswer will return the value v and unit u (if present) as strings
//in the first matching subpod. A subpod matches if its {.PlainText} field
//contains >= 10% numbers (i.e. 10% of the "words" (split by whitespace) are
//numerical)
//Will error if no numbers are found or there are no pods in the result
func (f *FullResult) GetNumericalAnswer() (v string, u string, e error) {
	return f.QueryResult.GetNumericalAnswer()
}

//GetNumericalAnswer will return the value v and unit u (if present) as strings
//in the first matching subpod. A subpod matches if its {.PlainText} field
//contains >= 10% numbers (i.e. 10% of the "words" (split by whitespace) are
//numerical)
//Prefers date answers over numbers. If a date is found, `errIsProbablyDate`
//will be returned
//Will error if no numbers are found or there are no pods in the result
func (q *QueryResult) GetNumericalAnswer() (v string, u string, e error) {
	if q.Numpods == 0 {
		return "", "", errNoPods
	}

	for _, p := range q.Pods {
		for _, sp := range p.SubPods {
			//don't check date answers for numerical-ness
			if isLongDateAnswer(sp.PlainText) {
				return "", "", errIsProbablyDate
			}

			found, v, u, e := analyzeSubPodForNumericalAnswer(sp)
			if found {
				return v, u, e
			}
		}
	}

	//errNoLikely means none of the answers contained enough numbers to be marked
	//a numerical answer
	return "", "", errNoLikely
}

//───Helpers────────────────────────────────────────────────────────────────────

//checks if a subpod contains a numerical answer, i.e. more than
//{numericalAnswerProbability}% (10%) of words match the number regex.
func analyzeSubPodForNumericalAnswer(sp SubPod) (found bool, v string, u string, e error) {
	words := strings.Split(removeParens(sp.PlainText), " ")
	numOfWords := len(words)
	countNumbers := 0

	//count words in the text
	for _, w := range words {
		if isNumber(w) {
			countNumbers++
		}
	}

	if float32(countNumbers)/float32(numOfWords) < numericalAnswerProbability {
		//not enough words are numbers, return not found (probably continue with the
		//next subpod
		return false, "", "", nil
	}

	matches := numberFinder.FindStringIndex(sp.PlainText)

	if matches == nil {
		return true, "", "", errNoMatch
	}

	idx := 0
	for i, word := range words {
		if isNumber(word) {
			idx = i
			break
		}
	}

	v = words[idx]

	//find unit
	if idx+1 < len(words) {
		u = words[idx+1]
	}

	return true, v, u, e
}

//if keepParens is set, the returned string will have text in parens removed.
//E.g. foo (bar) -> foo
func removeParens(s string) string {
	if keepParens {
		return s
	}
	return parenRemover(s, "")
}

func isNumber(s string) bool {
	return numericalRgx.MatchString(s)
}

//true if [words]... contains word
func has(word string, words ...string) bool {
	for _, x := range words {
		if word == x {
			return true
		}
	}
	return false
}

//true if {dateAnswerProbability}% (30%) of words in a string are "date words",
//i.e. weekdays or months (written out)
func isLongDateAnswer(s string) bool {
	words := strings.Split(
		//remove special characters first, then split by words
		regexp.MustCompile(`[-.,!?]`).ReplaceAllString(s, ""),
		" ")

	if len(words) == 0 {
		return false
	}

	weekdays := []string{
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
		"Sunday",
	}
	months := []string{
		"January",
		"February",
		"March",
		"April",
		"May",
		"June",
		"July",
		"August",
		"September",
		"October",
		"November",
		"December",
	}
	datewords := 0

	for _, word := range words {
		if has(word, weekdays...) || has(word, months...) {
			datewords++
		}
	}

	return float32(datewords)/float32(len(words)) > dateAnswerProbability
}

//ForEach will apply the given function f to every pod in the slice
func (ps Pods) ForEach(f func(Pod)) {
	for _, p := range ps {
		f(p)
	}
}

//ForEach will apply the given function f to every subpod in the slice
func (sps SubPods) ForEach(f func(SubPod)) {
	for _, sp := range sps {
		f(sp)
	}
}
