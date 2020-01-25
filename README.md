# Go Wolfram|Alpha API Adapter

## About this project

This is a library for the purpose of interfacing with the Wolfram|Alpha APIs.
Currently, their **Full Results** and **Spoken Results** APIs are implemented.
Please note that generated answers **DO NOT** automatically include any
attribution remarks; it is **YOUR** job as the software developer using this
library to ensure proper attribution is given for the answers according to
[Wolfram|Alpha's terms of use](https://www.wolframalpha.com/termsofuse).

## Installation

To add this library to your project, simply run the command

```sh
$ go get -u github.com/ethantrithon/wolframalpha
```

and you're ready to use it.

## Usage

There are two main use cases currently, as the library currently only implements
the Full and Spoken Results APIs. Here is an example piece of code that uses
both of them:

<details>
<summary>Code with comments</summary>

```go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ethantrithon/wolframalpha"
)

func main() {
	wolframalpha.APIKey("YOUR-API-KEY-HERE")
	//Note: "API Key" here is the general term,
	//Wolfram|Alpha calls it your "App ID" in the developer portal.

	stdin := bufio.NewScanner(os.Stdin)
	fmt.Print("What would you like to know? ")
	stdin.Scan()
	text := stdin.Text()

	//Send your query to the Spoken Results API.
	//The "standard" functions will return a channel where the result will
	//appear and run in their own goroutine
	answerChannelSpoken := wolframalpha.AskQuestionSpoken(text)
	answerChannel := wolframalpha.AskQuestion(text)

	//Here you could do other things while you wait for the API calls to finish

	spokenResult := <-answerChannelSpoken //string
	fullResult := <-answerChannel         //*wolframalpha.FullResult

	//will be "" if an error occurred somewhere
	fmt.Println(spokenResult)

	//"GetAnswer" is a special function which pulls the most likely answer from an
	//API call result. Currently, the order is:
	//1. Look for dates (19th January 2038)
	//2. Look for numbers, possibly with units (5770 or 5770 Kelvin)
	//3. "Fall back" to the longest answer string available
	//Be careful! If an error occurred during the *sending* of the request, the
	//resulting *wolframalpha.FullResult will be nil!
	answer, _ := fullResult.GetAnswer() //other return value is an error

	fmt.Println(answer)
}
```

</details>

<details>
<summary>Code without comments</summary>

```go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ethantrithon/wolframalpha"
)

func main() {
	wolframalpha.APIKey("YOUR-API-KEY-HERE")

	stdin := bufio.NewScanner(os.Stdin)
	fmt.Print("What would you like to know? ")
	stdin.Scan()
	text := stdin.Text()

	answerChannelSpoken := wolframalpha.AskQuestionSpoken(text)
	answerChannel := wolframalpha.AskQuestion(text)

	spokenResult := <-answerChannelSpoken
	fullResult := <-answerChannel

	fmt.Println(spokenResult)

	answer, _ := fullResult.GetAnswer()

	fmt.Println(answer)
}
```

</details>

Instead of functions like `AskQuestion`, which run in the background, there are
also the synchronous versions (e.g. `AskQuestionSync`) available. There is also
the version which returns not a `*wolframalpha.FullResult`, but instead the raw
`[]byte` containing the JSON data of the API call result (intended to be only
used internally, but it's there in case you want to use it for yourself.)

## License

© 2020, Ethan Trithon\
Mozilla Public License 2.0, MPL
