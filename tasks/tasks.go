// Package tasks provides needed types and operations on asana tasks data
package tasks

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Data struct {
	Data []Task `json:"data"`
}

// NewData creates a Data object
func NewData(r io.Reader) (*Data, error) {
	d := &Data{}

	err := d.FromJSON(r)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// FromJSON decodes json file into a Data object
func (data *Data) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(data)
}

// Task provides asana task data
type Task struct {
	GID         string       `json:"gid"`
	Assignee    Assignee     `json:"assignee"`
	Name        string       `json:"name"`
	Notes       string       `json:"notes"`
	CompletedAt string       `json:"completed_at"`
	Parent      Parent       `json:"parent"`
	Memberships []Membership `json:"memberships"`
	Subtasks    []Task       `json:"subtasks"`
}

func (t *Task) String() string {
	keyword := "TODO"
	includeCompleted := os.Getenv("INCLUDE_COMPLETED") == "true"
	assignee := os.Getenv("ASSIGNEE")

	if t.Assignee == (Assignee{}) || t.Assignee.Name != assignee {
		return ""
	}

	if t.CompletedAt != "" {
		if !includeCompleted {
			return ""
		}

		keyword = "DONE"
	}

	var taskBuilder strings.Builder
	taskBuilder.WriteString(
		fmt.Sprintf("** %s %s", keyword, strings.TrimSpace(t.Name)),
	)
	if t.Notes != "" {
		taskBuilder.WriteString(fmt.Sprintf("\n%s", strings.TrimSpace(t.Notes)))
		if len(t.Subtasks) == 0 && t.Parent == (Parent{}) {
			taskBuilder.WriteRune('\n')
		}
	}

	for _, st := range t.Subtasks {
		stString := st.String()
		if stString != "" {
			taskBuilder.WriteString(fmt.Sprintf("\n*%s", st.String()))
		}
	}
	if len(t.Subtasks) == 0 && t.Parent != (Parent{}) {
		taskBuilder.WriteRune('\n')
	}

	return taskBuilder.String()
}

// Assignee provides task assignee data
type Assignee struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

// Parent provides task parent data
type Parent struct {
	GID string `json:"gid"`
}

// Membership provides task memebership data
type Membership struct {
	Section *Section `json:"section"`
}

// Section provides task section data
type Section struct {
	GID   string `json:"gid"`
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}

func (s *Section) String() string {
	var sectionBuilder strings.Builder

	sectionBuilder.WriteString(fmt.Sprintf("* %s", s.Name))

	for _, t := range s.Tasks {
		tString := t.String()
		if tString != "" {
			sectionBuilder.WriteString(fmt.Sprintf("\n%s", tString))
		}
	}

	return sectionBuilder.String()
}

// Sections provides list of sections data
type Sections struct {
	kv map[string]*Section
	l  []string
}

// Get returns the section for a GID
func (ss *Sections) Get(k string) (*Section, bool) {
	v, ok := ss.kv[k]
	if !ok {
		return nil, false
	}

	return v, true
}

// Set sets (or replaces) the section for a GID
func (ss *Sections) Set(k string, v *Section) *Section {
	ss.kv[k] = v
	ss.l = append(ss.l, k)
	return v
}

// NewData creates a Sections object
func NewSections(tasks []Task) Sections {
	sections := Sections{}
	includeCompleted := os.Getenv("INCLUDE_COMPLETED") == "true"

	for _, task := range tasks {
		if includeCompleted || !includeCompleted && task.CompletedAt == "" {
			taskSection := task.Memberships[0].Section
			_, ok := sections.Get(taskSection.GID)
			if !ok {
				sections.Set(taskSection.GID, taskSection)
			}

			section, _ := sections.Get(taskSection.GID)
			section.Tasks = append(section.Tasks, task)
		}
	}

	return sections
}

func (ss Sections) String() string {
	var sectionsBuilder strings.Builder

	for _, gid := range ss.l {
		s, _ := ss.Get(gid)

		sectionsBuilder.WriteString(s.String())
		sectionsBuilder.WriteRune('\n')
	}

	return sectionsBuilder.String()
}
