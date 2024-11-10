package parser

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/user"
	"strings"
)

var (
	Directory string
)

// Get outbound IP
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// Get hostname
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	u, err := user.Current()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	} else {
		// Check if we need to set our home directory (initialize)
		if Directory == "" {
			Directory = u.HomeDir
		}
		return fmt.Sprintf("[green]%s@%s[-]:[blue]%s[-]$", u.Username, hostname, Directory)
	}
}

// Move down a directory
func moveDownDirectory() (bool, string) {
	index := strings.LastIndex(Directory, "/")

	if index != -1 && Directory != "/" {
		Directory = Directory[:index]
	}

	return true, Directory
}

// Checks whether or not a directory exists
func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func moveToDirectoryDefaultCase(arguments string) (bool, string) {
	if arguments[0] == '/' {
		if directoryExists(arguments) {
			Directory = arguments
			return true, fmt.Sprintf("echo New Directory: %s", Directory)
		}
	} else {
		if directoryExists(Directory + "/" + arguments) {
			Directory = Directory + "/" + arguments
			return true, fmt.Sprintf("echo New Directory: %s", Directory)
		}
	}
	return false, fmt.Sprintf("Directory did not exist: %s", arguments)
}

// Parses a command input
func ParseCommand(message string) (bool, string) {
	commandFields := strings.Fields(message)
	_, arguments, found := strings.Cut(message, commandFields[0])
	arguments = strings.TrimSpace(arguments)

	if found {
		// When we run cd, we want to return the new hostname since directory changed.
		switch strings.TrimSpace(string(commandFields[0])) {
		case "cd":
			switch arguments {
			case ".":
				return false, GetHostname()
			case "..":
				moveDownDirectory()
				return false, GetHostname()
			default:
				moveToDirectoryDefaultCase(arguments)
				return false, GetHostname()
			}
		case "CARP-ip":
			return false, GetOutboundIP()
		case "CARP-hostname":
			return false, GetHostname()
		default:
			return true, fmt.Sprintf("cd %s;%s", Directory, message)
		}
	}

	return false, "Unable to find command prefix in command line."
}

// func main() {
// 	success, dir := ParseCommand("cd /home/dtbm")
// 	fmt.Println("Success:", success, "New Directory:", dir)

// 	success, dir = ParseCommand("cd ..")
// 	fmt.Println("Success:", success, "New Directory:", dir)

// 	success, dir = ParseCommand("cd dtbm")
// 	fmt.Println("Success:", success, "New Directory:", dir)

// 	success, dir = ParseCommand("cd /home/dtbm/github")
// 	fmt.Println("Success:", success, "New Directory:", dir)
// }
