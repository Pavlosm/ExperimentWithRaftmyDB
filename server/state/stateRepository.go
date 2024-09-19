package state

import (
	"bufio"
	"fmt"
	"log/slog"
	"myDb/server/cfg"
	"os"
	"strconv"
	"strings"
)

type StateRepository struct {
	fileName string
}

type StateRepositoryHandler interface {
	WriteEntries(logs ...Log) error
	ReadEntries() ([]Log, error)
}

func InitStateRepository(cfg cfg.ServerConfig) *StateRepository {

	f := "logs_" + cfg.Me.MyNameString() + ".txt"

	return &StateRepository{
		fileName: f,
	}
}

func (s *StateRepository) WriteEntries(logs ...Log) error {

	f, err := os.OpenFile(s.fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		slog.Error("Error opening file", "error", err)
		return err
	}

	defer f.Close()

	for _, v := range logs {
		s := fmt.Sprintf("%v|%v|%v\n", v.Term, v.Index, v.Command)

		_, err = f.WriteString(s)
		if err != nil {
			slog.Error("Error writing logs to file", "error", err)
			return err
		}
	}

	return nil
}

func (s *StateRepository) ReadEntries() ([]Log, error) {
	file, err := os.OpenFile(s.fileName, os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logs []Log
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) == 3 {
			t, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				slog.Error("Error parsing term", "error", err)
				return nil, err
			}
			i, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				slog.Error("Error parsing index", "error", err)
				return nil, err
			}
			log := Log{
				Term:    t,
				Index:   i,
				Command: parts[2],
			}
			logs = append(logs, log)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}
