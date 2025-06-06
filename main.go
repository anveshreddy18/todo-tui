package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	taskBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(0, 1).
			Margin(0, 0, 1, 0)

	selectedTaskBoxStyle = taskBoxStyle.
				BorderForeground(lipgloss.Color("205")).
				Background(lipgloss.Color("236"))
)

// @anveshreddy18 -- in the last make it iota to make it easy to extend
type mode string

const (
	normalMode    mode = "normal"   // this mode is for viewing the dasboard for the regular addition of tasks
	completedMode mode = "done"     // this is for viewing the dashboard for the tasks that are done
	additionMode  mode = "addition" // mode for adding the task, a new input bar opens where the task can be added.
	editingMode   mode = "edit"     // the editing mode is used to identify & differentiate between similar actions performed in different modes.
)

type Task struct {
	name string
	id   int
}

type TaskList struct {
	list   []Task
	cursor int
}

type errMsg error

type model struct {
	spinner           spinner.Model
	quitting          bool
	err               error
	currentMode       mode
	taskListPerMode   map[mode]*TaskList
	taskID            int // monotonously increasing id for the tasks, every task will then have a unique ID
	additionTextInput textinput.Model
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel() model {
	// initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	taskListMap := make(map[mode]*TaskList)
	// Fill some values in the normalMode
	taskListMap[normalMode] = &TaskList{
		list: []Task{
			{
				id:   2,
				name: "two",
			},
			{
				id:   3,
				name: "three",
			},
			{
				id:   4,
				name: "four",
			},
			{
				id:   5,
				name: "five",
			},
		},
		cursor: 1,
	}
	// Fill some values in the completedMode
	taskListMap[completedMode] = &TaskList{
		list: []Task{
			{
				id:   0,
				name: "completed-zero",
			},
			{
				id:   1,
				name: "completed-one",
			},
		},
		cursor: 1,
	}
	// text input model for new tasks
	addNewTaskTextModel := textinput.New()
	addNewTaskTextModel.Placeholder = "your task goes here"
	addNewTaskTextModel.Prompt = ""
	addNewTaskTextModel.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // light grey
	addNewTaskTextModel.Width = 24                                                               // Set width to fit the placeholder
	addNewTaskTextModel.Focus()
	return model{
		spinner:           s,
		currentMode:       normalMode,
		taskListPerMode:   taskListMap,
		taskID:            len(taskListMap[normalMode].list) + len(taskListMap[completedMode].list),
		additionTextInput: addNewTaskTextModel,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		// leaving scope for adding more initialisations
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle additionMode text input
	if m.currentMode == additionMode {
		var cmd tea.Cmd
		m.additionTextInput, cmd = m.additionTextInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == tea.KeyEnter.String() {
			taskName := m.additionTextInput.Value()
			m.additionTextInput.Reset()
			if strings.TrimSpace(taskName) == "" {
				// If input is empty, just return to normal mode without adding a task
				m.currentMode = normalMode
				return m, nil
			}
			taskID := m.taskID
			m.taskID++
			newTask := Task{taskName, taskID}
			m.taskListPerMode[normalMode].list = append(m.taskListPerMode[normalMode].list, newTask)
			m.currentMode = normalMode
			return m, nil
		}
		return m, cmd
	}
	if m.currentMode == editingMode {
		var cmd tea.Cmd
		m.additionTextInput, cmd = m.additionTextInput.Update(msg)
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == tea.KeyEnter.String() {
			modifiedTaskName := m.additionTextInput.Value()
			// if the modifiedTaskName is empty, that means the task should be deleted
			if modifiedTaskName == "" {
				m.removeTaskFromCurrentList(normalMode)
			} else {
				taskList := m.taskListPerMode[normalMode]
				cursorInd := taskList.cursor
				taskList.list[cursorInd].name = modifiedTaskName
			}
			m.currentMode = normalMode
			return m, nil
		}
		return m, cmd
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		switch msg.String() {
		case "s":
			// Toggle between normal and completed modes.
			switch m.currentMode {
			case normalMode:
				m.currentMode = completedMode
			case completedMode:
				m.currentMode = normalMode
			}
		case tea.KeyUp.String(), tea.KeyLeft.String():
			if m.taskListPerMode[m.currentMode].cursor > 0 {
				m.taskListPerMode[m.currentMode].cursor--
			}
		case tea.KeyDown.String(), tea.KeyRight.String():
			taskList := m.taskListPerMode[m.currentMode]
			if taskList.cursor < len(taskList.list)-1 {
				taskList.cursor++
			}
		case tea.KeyEnter.String():
			// meaning the task is done, only if we're in the normal mode. If we're in completedMode, it's a no-op
			switch m.currentMode {
			case normalMode:
				// remove the task from the normal mode and add it to the completed mode
				curTask := m.removeTaskFromCurrentList(normalMode)

				// Only add if curTask is not empty
				if curTask.name != "" {
					completedList := m.taskListPerMode[completedMode]
					completedList.list = append(completedList.list, curTask)
				}
			case editingMode:
				// this should edit the task in the normalmode task list
			}
		case "d":
			// whenever d is pressed on a task, it should be removed from that list, either in normalMode or completedMode.
			m.removeTaskFromCurrentList(m.currentMode)
		case "a":
			// whenever 'a' is pressed, it should open up a screen which asks for the task name and when entered, it should add the task to normalMode tasklist
			m.currentMode = additionMode
			return m, textinput.Blink
		case "e":
			// editing is only applicable in the normal mode
			if m.currentMode != normalMode {
				return m, nil
			}
			// Use normalMode to get the current task, not editingMode
			curTaskList := m.taskListPerMode[normalMode]
			cursorInd := curTaskList.cursor
			if len(curTaskList.list) == 0 {
				return m, nil
			}
			taskName := curTaskList.list[cursorInd].name
			m.additionTextInput.SetValue(taskName)
			m.currentMode = editingMode
			return m, textinput.Blink
		}
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	// early returns
	if m.err != nil {
		return m.err.Error()
	}
	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf("\n%s\n", m.spinner.View()))
	if m.quitting {
		return b.String() + "\n"
	}
	// actual processing logic
	b.WriteString(fmt.Sprintf("%s Mode\n\n", m.currentMode))
	switch m.currentMode {
	case normalMode, completedMode:
		// Show the task list
		taskList := m.taskListPerMode[m.currentMode]
		for i, task := range taskList.list {
			taskStr := fmt.Sprintf("#%d %s", task.id, task.name)
			var box string
			if i == taskList.cursor {
				box = selectedTaskBoxStyle.Render(taskStr)
			} else {
				box = taskBoxStyle.Render(taskStr)
			}
			b.WriteString(box + "\n")
		}
	case additionMode:
		// Show the text input box which accepts keyboard input
		b.WriteString(fmt.Sprintf("Name the task: %s\n", m.additionTextInput.View()))
	case editingMode:
		b.WriteString(fmt.Sprintf("Edit the task: %s\n", m.additionTextInput.View()))
	}

	return b.String()
}

func (m model) removeTaskFromCurrentList(mode mode) Task {
	// modify the list corresponding to the given mode
	currentList := m.taskListPerMode[mode]
	taskIND := currentList.cursor
	if len(currentList.list) == 0 {
		return Task{}
	}
	curTask := currentList.list[taskIND]
	newCurrentList := []Task{}
	for i, task := range currentList.list {
		if i != taskIND {
			newCurrentList = append(newCurrentList, task)
		}
	}
	currentList.list = newCurrentList

	// adjusting the cursor in the given mode
	if m.taskListPerMode[mode].cursor >= len(m.taskListPerMode[mode].list) {
		m.taskListPerMode[mode].cursor = len(m.taskListPerMode[mode].list) - 1
	}
	if m.taskListPerMode[mode].cursor < 0 {
		m.taskListPerMode[mode].cursor = 0
	}
	return curTask
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
