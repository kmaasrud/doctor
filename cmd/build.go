package cmd

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kmaasrud/doctor/core"
	"github.com/kmaasrud/doctor/lua"
	"github.com/kmaasrud/doctor/msg"
	"github.com/kmaasrud/doctor/utils"
)

type WarningError struct {
	Stderr string
}

func (e *WarningError) Error() string {
	return e.Stderr
}

type FatalError struct {
	Stderr string
}

func (e *FatalError) Error() string {
	return e.Stderr
}

func Build() error {
	// Check for dependencies
	err := CheckPath("pandoc")
	if err != nil {
		return errors.New("Build failed. " + err.Error())
	}

	// Find root
	rootPath, err := utils.FindDoctorRoot()
	if err != nil {
		return errors.New("Build failed. " + err.Error())
	}

	// Initialize the command
	cmdArgs := []string{"-s", "-o", filepath.Join(rootPath, "main.pdf")}

	// Add resource paths
	resourcePaths := strings.Join([]string{rootPath, filepath.Join(rootPath, "assets"), filepath.Join(rootPath, "secs")}, utils.ResourceSep)
	cmdArgs = append(cmdArgs, "--resource-path="+resourcePaths)

	// Add Pandoc options from config. TODO: Clean this up a bit
	msg.Info("Applying configuration from doctor.toml...")
	conf, err := core.ConfigFromFile(filepath.Join(rootPath, "doctor.toml"))
	if err != nil {
		return err
	}
	jsonFilename := filepath.Join(rootPath, ".metadata.json")
	err = conf.WritePandocJson(jsonFilename)
	if err != nil {
		return err
	}
	cmdArgs = append(cmdArgs, "--metadata-file="+jsonFilename)

	// Specify PDF engine and add options for specific engines
	err = CheckPath(conf.Build.Engine)
	if err != nil {
		return errors.New("Build failed. " + err.Error())
	}
	cmdArgs = append(cmdArgs, fmt.Sprintf("--pdf-engine=%s", conf.Build.Engine))
	if conf.Build.Engine == "tectonic" {
        // Tectonic chatters a lot. Make it a bit more silent
		cmdArgs = append(cmdArgs, "--pdf-engine-opt=-c=minimal")
	}

	// Find source files
	msg.Info("Looking for source files...")
	secs, err := utils.FindSections(rootPath)
	if err != nil {
		return err
	}
	cmdArgs = append(cmdArgs, core.PathsFromSections(secs)...)
	msg.Info(fmt.Sprintf("Found %d source files!", len(secs)))

	// If references.bib exists, run with citeproc and add bibliography
	if _, err := os.Stat(filepath.Join(rootPath, "assets", "references.bib")); err == nil {
		msg.Info("Running with citeproc. Bibliography: " + filepath.Join("assets", "references.bib"))
		cmdArgs = append(cmdArgs, "-C", "--bibliography=references.bib")
	}

	// Make sure all temporary files are cleaned up after function is run
	defer cleanUp(rootPath, &conf)

	// Temporarily write any Lua filters to file and add them to command
	if conf.Build.LuaFilters {
		msg.Info("Adding Lua filters...")
		for filename, filter := range lua.Filters {
			err := os.WriteFile(filepath.Join(rootPath, filename), filter, 0644)
			if err != nil {
				return errors.New("Could not create Lua file. " + err.Error())
			}
			cmdArgs = append(cmdArgs, "-L", filename)
		}
	}

	// Execute command
	done := make(chan struct{})
	go msg.Do("Building document with Pandoc", done)
	err = runPandocWith(cmdArgs)
	msg.CloseDo(done)

	// Handle errors
	if err != nil {
		var warnStr, errStr string
		switch thisErr := err.(type) {
		case *FatalError:
			_, errStr = msg.CleanStderrMsg(thisErr.Stderr)
			return errors.New("Doctor exited with errors. They are as follows:\n\n" + errStr)
		case *WarningError:
			warnStr, _ = msg.CleanStderrMsg(thisErr.Stderr)
			msg.Success("Document built.")
			return errors.New("Doctor exited with warnings. They are as follows:\n\n" + warnStr)
		default:
			return errors.New("Could not run command. " + err.Error())
		}
	}
	msg.Success("Document built.")
	return nil
}

func runPandocWith(cmdArgs []string) error {
	var stderr bytes.Buffer
	cmd := exec.Command("pandoc", cmdArgs...)
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Fatal error
	if err != nil {
		return &FatalError{string(stderr.Bytes())}
	}
	// Non-fatal, but stderr is not empty, so it includes warnings
	if stderr := string(stderr.Bytes()); len(stderr) != 0 {
		return &WarningError{string(stderr)}
	}
	return nil
}

func cleanUp(rootPath string, conf *core.Config) {
	msg.Info("Cleaning up temporary files...")
	if conf.Build.LuaFilters {
		for filename := range lua.Filters {
			err := os.Remove(filepath.Join(rootPath, filename))
			if err != nil {
				msg.Error("Failed to remove Lua filter " + filename + ". " + err.Error())
			}
		}
	}

	err := os.Remove(filepath.Join(rootPath, ".metadata.json"))
	if err != nil {
		msg.Error("Failed to remove JSON metadata file. " + err.Error())
	}
}
