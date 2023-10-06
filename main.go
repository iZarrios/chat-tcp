package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/iZarrios/chat-tcp/tcpchat"
)

type (
	errMsg error
)

type model struct {
	sub      chan struct{} // where we'll receive activity notifications
	Conn     net.Conn
	viewport viewport.Model
	messages []string
	textarea textarea.Model
	err      error
}

func (m *model) ReadMessagesFromConnection() {
	buffer := make([]byte, 2048) // Adjust the buffer size as needed
	model := m
	conn := m.Conn
	n, err := conn.Read(buffer)
	if err != nil || n <= 0 {
		// Handle the error (e.g., connection closed)
		return
	}

	// Process and display the received message
	message := string(buffer[:n])

	model.messages = append(model.messages, message)
	model.viewport.SetContent(strings.Join(model.messages, "\n"))
	model.textarea.Reset()
	model.viewport.GotoBottom()
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(100, 20)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		// Handle the error
		panic(err)
	}
	conn.Write([]byte("/join hi\n"))

	return model{
		Conn:     conn,
		textarea: ta,
		messages: []string{},
		viewport: vp,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	// m.ReadMessagesFromConnection()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			m.Conn.Write([]byte("/quit\n"))
			m.Conn.Close()
			return m, tea.Quit
		case tea.KeyEnter:

			msg := m.textarea.Value()
			m.Conn.Write([]byte("/msg " + msg + "\n"))

			m.messages = append(m.messages, "You: "+msg)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		case tea.KeyCtrlA:
			msg := m.textarea.Value()
			m.Conn.Write([]byte("/join " + msg + "\n"))

			m.messages = append(m.messages, "You: "+msg)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		case tea.KeyCtrlU:
			m.textarea.Reset()
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func main() {
	argc := len(os.Args)
	if argc < 2 {
		// server
		s := tcpchat.NewServer(
			tcpchat.WithAddress(":3000"),
			tcpchat.WithCmdBufferSize(100),
		)
		msg := fmt.Sprintf("Listening on %v with buffer size of %v and trying to watch room '%v' \n", s.ListenAddress[1:], cap(s.Cmds), s.RoomWatcher)
		fmt.Print(msg)

		err := s.Start()
		if err != nil {
			log.Fatal("Could not start server ", err)
		}
	} else {
		//client
		model := initialModel()
		// go readMessagesFromConnection(model.Conn, &model)
		p := tea.NewProgram(model)

		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}

	}

}
