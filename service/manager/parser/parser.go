package parser

import (
	"fmt"
	"strings"
)

var Aliases = make(map[string][]string)

// Parses an incoming command to see if it meets one of the predefined rulesets for commands, presently:
// Alias
//   - alias <name> <csv of ips / wildcard address>
func ParseIncoming(command string) string {
	commandFields := strings.Fields(command)
	commandPrefix := commandFields[0]

	switch commandPrefix {
	case "alias":
		if len(commandFields) < 3 {
			return fmt.Sprintf(" Invalid command format for alias : Too few arguments [%d]", len(commandFields))
		}
		name := commandFields[1]
		argument := commandFields[2]
		// alias <name> <wildcard>

		Aliases[name] = append(Aliases[name], strings.Split(argument, ",")...)
		return fmt.Sprintf(" Successfully added alias %s -> %s", commandFields[1], commandFields[2])
	default:
		return fmt.Sprintf(" Unknown global command: %s", command)
	}
}

func CheckAddress(address string) []string {
	alias, ok := Aliases[address]

	if ok {
		return alias
	}

	return []string{""}
}
