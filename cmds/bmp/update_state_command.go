package bmp

import (
	"errors"

	clients "github.com/cloudfoundry-community/bosh-softlayer-tools/clients"
	cmds "github.com/cloudfoundry-community/bosh-softlayer-tools/cmds"
	common "github.com/cloudfoundry-community/bosh-softlayer-tools/common"
	fakes "github.com/cloudfoundry-community/bosh-softlayer-tools/common/fakes"
)

type updateStateCommand struct {
	options cmds.Options

	ui      common.UI
	printer common.Printer

	bmpClient clients.BmpClient
}

func NewUpdateStateCommand(options cmds.Options, bmpClient clients.BmpClient) updateStateCommand {
	consoleUi := common.NewConsoleUi(options.Verbose)

	return updateStateCommand{
		options: options,

		ui:        consoleUi,
		printer:   common.NewDefaultPrinter(consoleUi, options.Verbose),
		bmpClient: bmpClient,
	}
}

func NewFakeUpdateStateCommand(options cmds.Options, bmpClient clients.BmpClient, userInput string) updateStateCommand {
	FakeUi := fakes.NewFakeUi()
	FakeUi.UserInput = userInput

	return updateStateCommand{
		options:   options,
		ui:        FakeUi,
		printer:   common.NewDefaultPrinter(FakeUi, options.Verbose),
		bmpClient: bmpClient,
	}
}

func (cmd updateStateCommand) Name() string {
	return "update-state"
}

func (cmd updateStateCommand) Description() string {
	return `Update the server state (\"bm.state.new\", \"bm.state.using\", \"bm.state.loading\", \"bm.state.failed\", \"bm.state.deleted\")`
}

func (cmd updateStateCommand) Usage() string {
	return "bmp update-state --server <server-id> --state <state>"
}

func (cmd updateStateCommand) Options() cmds.Options {
	return cmd.options
}

func (cmd updateStateCommand) Validate() (bool, error) {
	cmd.printer.Printf("Validating %s command: options: %#v", cmd.Name(), cmd.options)

	if cmd.options.Server == "" {
		return false, errors.New("cannot have empty server ID")
	}

	if cmd.options.State == "" {
		return false, errors.New("cannot have empty state")
	} else {
		if !cmd.isValidState(cmd.options.State) {
			return false, errors.New("the server state is incorrect!")
		}
	}

	return true, nil
}

func (cmd updateStateCommand) Execute(args []string) (int, error) {
	cmd.printer.Printf("Executing %s command: args: %#v, options: %#v", cmd.Name(), args, cmd.options)

	cmd.ui.PrintlnInfo("WARNING: Be careful updating the state of a server, as it might break your deployment!")
	if cmd.isConfirmed() {
		updateStateResponse, err := cmd.bmpClient.UpdateState(cmd.options.Server, cmd.options.State)
		if err != nil {
			return 1, err
		}

		if updateStateResponse.Status != 200 {
			return updateStateResponse.Status, nil
		}

		cmd.ui.PrintlnInfo("Update Successful!")

		return 0, nil
	} else {
		return 1, errors.New("Update Cancelled!")
	}
}

// Private Methods

func (cmd updateStateCommand) isValidState(state string) bool {
	switch state {
	case "bm.state.new":
		return true
	case "bm.state.using":
		return true
	case "bm.state.loading":
		return true
	case "bm.state.failed":
		return true
	case "bm.state.deleted":
		return true
	}

	return false
}

func (cmd updateStateCommand) isConfirmed() bool {
	var userInput string
	cmd.ui.PrintfInfo("Continue to update? (type 'yes' or 'y' to continue)")
	_, err := cmd.ui.Scanln(&userInput)
	if err != nil {
		return false
	}

	if cmd.isYes(userInput) {
		return true
	}

	return false
}

func (cmd updateStateCommand) isYes(userInput string) bool {
	switch userInput {
	case "y":
		return true
	case "Y":
		return true
	case "yes":
		return true
	case "Yes":
		return true
	case "YES":
		return true
	}

	return false
}
