package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// CommandMap maps FTP commands to Parasise handlers.
var CommandMap map[string]func(*Paradise)

func init() {
	CommandMap = make(map[string]func(*Paradise))

	CommandMap["USER"] = (*Paradise).HandleUser
	CommandMap["PASS"] = (*Paradise).HandlePass
	CommandMap["STOR"] = (*Paradise).HandleStore
	CommandMap["APPE"] = (*Paradise).HandleStore
	CommandMap["STAT"] = (*Paradise).HandleStat
	CommandMap["SYST"] = (*Paradise).HandleSyst
	CommandMap["PWD"] = (*Paradise).HandlePwd
	CommandMap["TYPE"] = (*Paradise).HandleType
	CommandMap["PASV"] = (*Paradise).HandlePassive
	CommandMap["EPSV"] = (*Paradise).HandlePassive
	CommandMap["NLST"] = (*Paradise).HandleList
	CommandMap["LIST"] = (*Paradise).HandleList
	CommandMap["QUIT"] = (*Paradise).HandleQuit
	CommandMap["CWD"] = (*Paradise).HandleCwd
	CommandMap["SIZE"] = (*Paradise).HandleSize
	CommandMap["RETR"] = (*Paradise).HandleRetr
}

// Paradise encapsulates an FTP connection and
// associated streams and synchronization structures.
type Paradise struct {
	writer        *bufio.Writer
	reader        *bufio.Reader
	theConnection net.Conn
	passiveConn   *net.TCPConn
	waiter        sync.WaitGroup
	user          string
	homeDir       string
	path          string
	ip            string
	command       string
	param         string
	total         int64
	buffer        []byte
}

// NewParadise is the factory function for Paradise values.
func NewParadise(connection net.Conn) *Paradise {
	p := Paradise{}

	p.writer = bufio.NewWriter(connection)
	p.reader = bufio.NewReader(connection)
	p.path = "/"
	p.theConnection = connection
	p.ip = connection.RemoteAddr().String()
	return &p
}

// HandleCommmands handles FTP commands.
func (self *Paradise) HandleCommands() {
	fmt.Println("Got client on: ", self.ip)
	self.writeMessage(220, "Welcome to Paradise")
	for {
		line, err := self.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				//continue
			}
			break
		}
		command, param := parseLine(line)
		self.command = command
		self.param = param

		fn := CommandMap[command]
		if fn == nil {
			self.writeMessage(550, "not allowed")
		} else {
			fn(self)
		}
	}
}

func (self *Paradise) writeMessage(code int, message string) {
	line := fmt.Sprintf("%d %s\r\n", code, message)
	self.writer.WriteString(line)
	self.writer.Flush()
}

func (self *Paradise) closePassiveConnection() {
	if self.passiveConn != nil {
		self.passiveConn.Close()
	}
}

func parseLine(line string) (string, string) {
	params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}
	return params[0], strings.TrimSpace(params[1])
}
