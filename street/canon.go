package street

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var apartmentNumberRegex = regexp.MustCompile(`^(\d+)(.*)$`)

func Canon(address string) string {
	lines := strings.SplitN(address, ",", 2)
	if len(lines) != 2 {
		return address
	}

	firstLine := lines[0]
	remainingLines := lines[1]

	apartmentNumberString := ""
	if index := strings.Index(firstLine, "#"); index >= 0 {
		apartmentLine := strings.TrimSpace(firstLine[index+1:])
		firstLine = strings.TrimSpace(firstLine[:index])

		matches := apartmentNumberRegex.FindStringSubmatch(apartmentLine)
		if len(matches) == 3 {
			numberPart := matches[1]
			suffixPart := matches[2]

			number, err := strconv.ParseInt(numberPart, 10, 64)
			if err == nil {
				apartmentNumberString = fmt.Sprintf("%09d", number)
			}
			apartmentNumberString += suffixPart
		}
	}

	parts := strings.SplitN(firstLine, " ", 2)
	addressNumberString := parts[0]
	streetInfo := parts[1]
	{
		addressNumber, err := strconv.ParseInt(addressNumberString, 10, 64)
		if err == nil {
			addressNumberString = fmt.Sprintf("%09d", addressNumber)
			if addressNumber%2 == 0 {
				addressNumberString = "e" + addressNumberString
			} else {
				addressNumberString = "o" + addressNumberString
			}
		}
	}
	return streetInfo + " " + addressNumberString + "#" + apartmentNumberString + "," + remainingLines
}
