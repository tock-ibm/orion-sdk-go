package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.ibm.com/blockchaindb/sdk/examples/cars/commands"
	"github.ibm.com/blockchaindb/server/pkg/logger"
	"gopkg.in/alecthomas/kingpin.v2"
)

const demoDirEnvar = "CARS_DEMO_DIR"

func main() {
	kingpin.Version("0.0.1")

	c := &logger.Config{
		Level:         "info",
		OutputPath:    []string{"stdout"},
		ErrOutputPath: []string{"stderr"},
		Encoding:      "console",
		Name:          "bcdb-client",
	}
	lg, err := logger.New(c)

	output, exit, err := executeForArgs(os.Args[1:], lg)
	if err != nil {
		kingpin.Fatalf("parsing arguments: %s. Try --help", err)
	}
	fmt.Println(output)
	os.Exit(exit)
}

func executeForArgs(args []string, lg *logger.SugarLogger) (output string, exit int, err error) {
	//
	// command line flags
	//
	app := kingpin.New("cars", "Car registry demo")
	demoDir := app.Flag("demo-dir",
		fmt.Sprintf("Path to the folder that will contain all the material for the demo. If missing, taken from envar: %s", demoDirEnvar)).
		Short('d').
		Envar(demoDirEnvar).
		Required().
		String()

	generate := app.Command("generate", "Generate crypto material for all roles: admin, dmv, dealer, alice, bob; and the BCDB server.")

	init := app.Command("init", "Initialize the server, load it with users, create databases.")
	replica := init.Flag("server", "URI of blockchain DB replica, http://host:port, to connect to").Short('s').Required().URL()

	mintRequest := app.Command("mint-request", "Issue a request to mint a car by a dealer.")
	mrUserID := mintRequest.Flag("user", "dealer user ID").Short('u').Required().String()
	mrCarRegistry := mintRequest.Flag("car", "car registration plate").Short('c').Required().String()

	mintApprove := app.Command("mint-approve", "Approve a request to mint a car, create car record.")
	maUserID := mintApprove.Flag("user", "DMV user ID").Short('u').Required().String()
	maRequestRecKey := mintApprove.Flag("mint-request-key", "mint-request record key").Short('k').Required().String()

	transferTo := app.Command("transfer-to", "A seller issues an contract to sell a car")
	ttUserID := transferTo.Flag("user", "seller user ID").Short('u').Required().String()
	ttBuyerID := transferTo.Flag("buyer", "buyer user ID").Short('b').Required().String()
	ttCar := transferTo.Flag("car", "car registration plate").Short('c').Required().String()

	transferReceive := app.Command("transfer-receive", "A buyer agrees to the contract to sell a car")
	trUserID := transferReceive.Flag("user", "buyer user ID").Short('u').Required().String()
	trCar := transferReceive.Flag("car", "car registration plate").Short('c').Required().String()
	trTrsToRecordKey := transferReceive.Flag("transfer-to-key", "transfer-to record key").Short('k').Required().String()
	command := kingpin.MustParse(app.Parse(args))

	//
	// call the underlying implementations
	//
	var resp *http.Response

	switch command {
	case generate.FullCommand():
		err := commands.Generate(*demoDir)
		if err != nil {
			return "", 1, err
		}
		return "Generated demo materials to: " + *demoDir, 0, nil

	case init.FullCommand():
		err := commands.Init(*demoDir, *replica, lg)
		if err != nil {
			return "", 1, err
		}
		return "Initialized server from: " + *demoDir, 0, nil

	case mintRequest.FullCommand():
		out, err := commands.MintRequest(*demoDir, *mrUserID, *mrCarRegistry, lg)
		if err != nil {
			return "", 1, err
		}

		return fmt.Sprintf("Issued mint request:\n%s\n", out), 0, nil

	case mintApprove.FullCommand():
		out, err := commands.MintApprove(*demoDir, *maUserID, *maRequestRecKey, lg)
		if err != nil {
			return "", 1, err
		}

		return fmt.Sprintf("Approved mint request:\n%s\n", out), 0, nil

	case transferTo.FullCommand():
		out, err := commands.TransferTo(*demoDir, *ttUserID, *ttBuyerID, *ttCar, lg)
		if err != nil {
			return "", 1, err
		}

		return fmt.Sprintf("Issued transfer-to:\n%s\n", out), 0, nil

	case transferReceive.FullCommand():
		out, err := commands.TransferReceive(*demoDir, *trUserID, *trCar, *trTrsToRecordKey, lg)

		if err != nil {
			return "", 1, err
		}

		return fmt.Sprintf("Issued transfer-receive:\n%s\n", out), 0, nil
	}

	if err != nil {
		return errorOutput(err), 1, nil
	}

	bodyBytes, err := readBodyBytes(resp.Body)
	if err != nil {
		return errorOutput(err), 1, nil
	}

	return responseOutput(resp.StatusCode, bodyBytes), 0, nil
}

func responseOutput(statusCode int, responseBody []byte) string {
	status := fmt.Sprintf("Status: %d", statusCode)

	var buffer bytes.Buffer
	json.Indent(&buffer, responseBody, "", "\t")
	response := fmt.Sprintf("%s", buffer.Bytes())

	output := fmt.Sprintf("%s\n%s", status, response)

	return output
}

func readBodyBytes(body io.ReadCloser) ([]byte, error) {
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("reading http response body: %s", err)
	}
	body.Close()

	return bodyBytes, nil
}

func errorOutput(err error) string {
	return fmt.Sprintf("Error: %s\n", err)
}