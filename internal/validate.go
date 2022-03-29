package internal

import (
	"errors"
	"regexp"
)

var (
	RegArn     = regexp.MustCompile("[\u0009\u000A\u000D\u0020-\u007E\u0085\u00A0-\uD7FF\uE000-\uFFFD\u10000-\u10FFFF]+")
	RegSession = regexp.MustCompile(`[\w+=,.@-]*`)
)

func ValidateSessionName(sn string) error {
	// https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.html
	if len(sn) < 2 || len(sn) > 64 {
		return errors.New("SessionName length constraints: minimum length of 2, maximum length of 64")
	}
	if !RegSession.MatchString(sn) {
		return errors.New(`SessionName must match [\w+=,.@-]*`)
	}
	return nil
}

func ValidateArn(sn string) error {
	// https://docs.aws.amazon.com/STS/latest/APIReference/API_AssumeRole.html
	if len(sn) < 20 || len(sn) > 2048 {
		return errors.New("ARN length constraints: minimum length of 20, maximum length of 2048")
	}
	if !RegArn.MatchString(sn) {
		return errors.New(`ARN must match [\u0009\u000A\u000D\u0020-\u007E\u0085\u00A0-\uD7FF\uE000-\uFFFD\u10000-\u10FFFF]+`)
	}
	return nil
}
