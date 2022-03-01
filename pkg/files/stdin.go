package files

import (
	"fmt"
	"io/ioutil"
	"os"
)

var hasStdinBeenRead bool

func ReadStdin() ([]byte, error) {
	if hasStdinBeenRead {
		return nil, fmt.Errorf("Standard input has been already read once (has the '-' argument been used in more than one flags?)")
	}
	hasStdinBeenRead = true
	return ioutil.ReadAll(os.Stdin)
}
