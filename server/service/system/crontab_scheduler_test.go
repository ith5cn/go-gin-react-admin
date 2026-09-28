package system

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	systemModel "server/model/system"
	systemRequest "server/model/system/request"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/robfig/cron/v3"
)

func TestSchedulerPollPreservesUnchangedEntriesAndSurvivesReadFailure(t *testing.T) {
	_, mock := userTestDB(t)
	s := &crontabScheduler{cron: cron.New(cron.WithParser(crontabParser)), entries: map[uint]cron.EntryID{}}
	rows := func() *sqlmock.Rows {
		return sqlmock.NewRows([]string{"id", "rule", "task_style", "target", "status"}).AddRow(1, "*/5 * * * *", 1, "system/clean-logs", 1)
	}
	mock.ExpectQuery("SELECT .*ai_tool_crontab").WithArgs(1).WillReturnRows(rows())
	if err := s.reloadLocked(); err != nil {
		t.Fatal(err)
	}
	id := s.entries[1]
	if id == 0 {
		t.Fatal("task not registered")
	}
	mock.ExpectQuery("SELECT .*ai_tool_crontab").WithArgs(1).WillReturnRows(rows())
	if err := s.reloadLocked(); err != nil {
		t.Fatal(err)
	}
	if s.entries[1] != id {
		t.Fatal("unchanged polling reset the schedule")
	}
	mock.ExpectQuery("SELECT .*ai_tool_crontab").WithArgs(1).WillReturnError(errors.New("database unavailable"))
	if err := s.reloadLocked(); err == nil {
		t.Fatal("read error lost")
	}
	if s.entries[1] != id {
		t.Fatal("read failure removed live task")
	}
	mock.ExpectQuery("SELECT .*ai_tool_crontab").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id"}))
	if err := s.reloadLocked(); err != nil {
		t.Fatal(err)
	}
	if len(s.entries) != 0 {
		t.Fatal("disabled task still registered")
	}
}

func TestCrontabInternalTasksSortedAndMetadata(t *testing.T) {
	const (
		firstName  = "test/catalog-a"
		secondName = "test/catalog-z"
	)
	RegisterCrontabTaskWithMetadata(secondName, "Z task", "z hint", func(string) error { return nil })
	RegisterCrontabTaskWithMetadata(firstName, "A task", "a hint", func(string) error { return nil })
	t.Cleanup(func() {
		crontabTaskRegistryMu.Lock()
		delete(crontabTaskRegistry, firstName)
		delete(crontabTaskRegistry, secondName)
		crontabTaskRegistryMu.Unlock()
	})

	options := CrontabInternalTasks()
	values := make([]string, 0, len(options))
	var found *CrontabInternalTaskOption
	for index := range options {
		values = append(values, options[index].Value)
		if options[index].Value == firstName {
			option := options[index]
			found = &option
		}
	}
	if !sort.StringsAreSorted(values) {
		t.Fatalf("internal task options are not sorted: %v", values)
	}
	if found == nil || found.Label != "A task" || found.ParameterHint != "a hint" {
		t.Fatalf("registered metadata missing from catalog: %#v", found)
	}
}

func TestValidateCrontabTarget(t *testing.T) {
	tests := []struct {
		name   string
		style  *int16
		target *string
		want   error
	}{
		{name: "registered internal task", style: int16Pointer(1), target: stringPointer("system/clean-logs")},
		{name: "unknown internal task", style: int16Pointer(1), target: stringPointer("system/missing"), want: ErrCrontabTaskUnknown},
		{name: "valid http URL", style: int16Pointer(2), target: stringPointer("https://example.com/jobs/run")},
		{name: "valid http URL with port", style: int16Pointer(2), target: stringPointer("http://localhost:8080/run")},
		{name: "URL without host", style: int16Pointer(2), target: stringPointer("https:///run"), want: ErrCrontabURLInvalid},
		{name: "unsupported URL scheme", style: int16Pointer(2), target: stringPointer("ftp://example.com/run"), want: ErrCrontabURLInvalid},
		{name: "empty target", style: int16Pointer(1), target: stringPointer("  "), want: ErrCrontabTargetRequired},
		{name: "missing style", target: stringPointer("system/clean-logs"), want: ErrCrontabStyleInvalid},
		{name: "invalid style", style: int16Pointer(3), target: stringPointer("system/clean-logs"), want: ErrCrontabStyleInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCrontabTarget(systemModel.AIToolCrontab{
				TaskStyle: test.style,
				Target:    test.target,
			})
			if err != test.want {
				t.Fatalf("validateCrontabTarget() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestRegisteredCrontabTaskRuns(t *testing.T) {
	const name = "test/run"
	var got string
	RegisterCrontabTask(name, func(parameter string) error {
		got = parameter
		return nil
	})
	t.Cleanup(func() {
		crontabTaskRegistryMu.Lock()
		delete(crontabTaskRegistry, name)
		crontabTaskRegistryMu.Unlock()
	})

	if err := runCrontabTarget(systemModel.AIToolCrontab{
		TaskStyle: int16Pointer(1),
		Target:    stringPointer(name),
		Parameter: stringPointer("payload"),
	}); err != nil {
		t.Fatalf("runCrontabTarget() error = %v", err)
	}
	if got != "payload" {
		t.Fatalf("registered task received %q, want payload", got)
	}
}

func TestMergeCrontabPayloadBeforeValidation(t *testing.T) {
	current := systemModel.AIToolCrontab{
		Target:    stringPointer("system/clean-logs"),
		TaskStyle: int16Pointer(1),
	}
	unchanged := mergeCrontabPayload(current, systemRequest.CrontabPayload{})
	if !reflect.DeepEqual(unchanged, current) {
		t.Fatalf("partial update did not preserve existing target fields: %#v", unchanged)
	}

	target := "https://example.com/run"
	style := systemRequest.FlexInt16(2)
	merged := mergeCrontabPayload(current, systemRequest.CrontabPayload{
		Target:    &target,
		TaskStyle: &style,
	})
	if merged.Target == nil || *merged.Target != target || merged.TaskStyle == nil || *merged.TaskStyle != 2 {
		t.Fatalf("partial update was not merged before validation: %#v", merged)
	}
	if err := validateCrontabTarget(merged); err != nil {
		t.Fatalf("merged task should be valid: %v", err)
	}
}

func int16Pointer(value int16) *int16 {
	return &value
}

func stringPointer(value string) *string {
	return &value
}
