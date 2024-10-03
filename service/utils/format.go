package utils

// Imports
import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// Colors
var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

// GetIPFromRequest extracts the IP address from the request
func GetIPFromRequest(request *http.Request) string {
    remoteAddr := request.RemoteAddr

    ip, _, err := net.SplitHostPort(remoteAddr)
    if err != nil {
        return ""
    }

    return ip
}

// WrapErrorCheck checks whether or not ErrorState has a value, and logs it, otherwise logs SuccessState.
func WrapErrorCheck(request *http.Request, ErrorState error, SuccessState string) {
	Address := GetIPFromRequest(request)
	if ErrorState != nil {
		fmt.Println(fmt.Sprintf("[%s! %s%s] - %s%s%s - %s: %s", Red, time.Now().Format("15:04:05"), Reset, Cyan, Address, Reset, SuccessState, ErrorState.Error()))
	} else {
		fmt.Println(fmt.Sprintf("[%s+ %s%s] - %s%s%s - %s", Green, time.Now().Format("15:04:05"), Reset, Cyan, Address, Reset, SuccessState))
	}
}

// WrapErrorCheck checks whether or not ErrorState has a value, and logs it, otherwise logs SuccessState.
func LogMessage(ipAddress string, Message string) {
	fmt.Println(fmt.Sprintf("[%s- %s%s] - %s%s%s - %s", Yellow, time.Now().Format("15:04:05"), Reset, Cyan, ipAddress, Reset, Message))
}

