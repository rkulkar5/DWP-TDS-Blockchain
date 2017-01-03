/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"errors"
	"fmt"
	"strconv"
	"encoding/json"
	"crypto/rand"
	"io"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

var marbleIndexStr = "_marbleindex"				//name for the key/value that will store a list of all known marbles
var openTradesStr = "_opentrades"				//name for the key/value that will store all open trades

type Tax struct{
	UUID string `json:"uuid"`
	Pan string `json:"Pan"`					//the fieldtags are needed to keep case from bouncing around
	Date string `json:"Date"`
	IncomeSource string `json:"Source"`
	Income string `json:"Income"`
	//Tax string `json:"Tax"`   
}

type AllTaxes struct{
	AllPanTransactions []Tax `json:"all_taxes"`
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var Aval int
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	// Initialize the chaincode
	Aval, err = strconv.Atoi(args[0])
	if err != nil {
		return nil, errors.New("Expecting integer value for asset holding")
	}

	// Write the state to the ledger
	err = stub.PutState("abc", []byte(strconv.Itoa(Aval)))				//making a test var "abc", I find it handy to read/write to it right away to test the network
	if err != nil {
		return nil, err
	}
	
	var empty []string
	jsonAsBytes, _ := json.Marshal(empty)								//marshal an emtpy array of strings to clear the index
	err = stub.PutState(marbleIndexStr, jsonAsBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ============================================================================================================================
// Run - Our entry point for Invocations - [LEGACY] obc-peer 4/25/2016
// ============================================================================================================================
func (t *SimpleChaincode) Run(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	
	} else if function == "init_tax" {									//create a new marble
		return t.init_tax(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation")
}

// ============================================================================================================================
// Query - Our entry point for Queries
// ============================================================================================================================
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	} else if function == "readAllByPan" {													//read a variable
		return t.readAllByPan(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var name, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	name = args[0]
	valAsbytes, err := stub.GetState(name)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + name + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}




// ============================================================================================================================
// Init Marble - create a new marble, store into chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) init_tax(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0       1       2     3           4
	// "pan", "date", "source", "income" , tax
	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 3")
	}

	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}
	
	//if len(args[4]) <= 0 {
	//	return nil, errors.New("5th argument must be a non-empty string")
	//}
	
	uuid, err := newUUID()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Printf(" The uuid is: %s\n", uuid)
	
	
	
	str := `{"uuid": "`+uuid +`", "Pan": "` + args[0] + `", "Date": "` + args[1] + `", "Source": "` + args[2] + `", "Income": "` + args[3] + `"}`
	err = stub.PutState(uuid, []byte(str))
	
	//store marble with id as key
	if err != nil {
		return nil, err
	}
	
	
	//get the marble index
	marblesAsBytes, err := stub.GetState(marbleIndexStr)
	if err != nil {
		return nil, errors.New("Failed to get marble index")
	}
	var marbleIndex []string
	json.Unmarshal(marblesAsBytes, &marbleIndex)								//un stringify it aka JSON.parse()
	
	//append
	marbleIndex = append(marbleIndex, uuid)								//add marble name to index list
	fmt.Println("! marble index: ", marbleIndex)
	jsonAsBytes, _ := json.Marshal(marbleIndex)
	err = stub.PutState(marbleIndexStr, jsonAsBytes)						//store name of marble

	fmt.Println("- end init tax")
	return nil, nil
}



// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}


// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) readAllByPan(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var jsonResp string
	var err error
	var allTaxes AllTaxes
	var res Tax

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	uuids, err := stub.GetState(marbleIndexStr)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
        
	var uuidList []string
	json.Unmarshal(uuids, &uuidList)
	for i:= range uuidList{
		
		
		panDetails, err := stub.GetState(uuidList[i])						//grab this marble
		if err != nil {
			return nil, nil
		}
		res := Tax{}
		json.Unmarshal(panDetails, &res)
		
		allTaxes.AllPanTransactions[i] = res;
	}
	
	jsonAsBytes, _ := json.Marshal(res)
	
	return jsonAsBytes, nil													//send it onward
}
