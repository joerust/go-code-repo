/*
Copyright IBM Corp 2016 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
		 http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
	"strings"
    "encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// ReferralPartnerChaincodeBroker implementation stores and updates referral information on the blockchain
type ReferralPartnerChaincodeBroker struct {
	RetailChaincode string
	CommercialChaincode string
	BankingChaincode string
}

type CustomerReferral struct {
	ReferralId string `json:"referralId"`
    CustomerName string `json:"customerName"`
	ContactNumber string `json:"contactNumber"`
	CustomerId string `json:"customerId"`
	EmployeeId string `json:"employeeId"`
	Departments []string `json:"departments"`
    CreateDate int64 `json:"createDate"`
	Status string `json:"status"`
	Mortgage Mortgage `json:"mortgage"`
}

type Mortgage struct {
	MortgageNumber string `json:"mortgageNumber"`
    MortgageType string `json:"mortgageType"`
	ReferralId string `json:"referralId"`
	Rate string `json:"rate"`
	Amount string `json:"amount"`
}

func main() {
	err := shim.Start(new(ReferralPartnerChaincodeBroker))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}	
}

func BytesToString(b []byte) string {
    bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
    sh := reflect.StringHeader{bh.Data, bh.Len}
    return *(*string)(unsafe.Pointer(&sh))
}

// Init resets all the things
func (t *ReferralPartnerChaincodeBroker) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	// Initialize the partner names
	t.RetailChaincode = args[0]
	t.CommercialChaincode = args[1]
	t.BankingChaincode = args[2]
	
	return nil, nil
}

// Invoke is our entry point to invoke a chaincode function
func (t *ReferralPartnerChaincodeBroker) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "createReferral" {
		return t.createReferral(stub, args)
	} else if function == "updateReferralStatus" {
		return t.updateReferralStatus(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation")
}

// Query is our entry point for queries
func (t *ReferralPartnerChaincodeBroker) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	} else if function == "searchByStatus" {
		return t.searchByStatus(args, stub)
	} else if function == "searchByPartner" {
		return t.searchByPartner(stub, args)
	}
	
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query")
}

func unmarshallBytes(valAsBytes []byte) (error, CustomerReferral) {
	var err error
	var referral CustomerReferral

	err = json.Unmarshal(valAsBytes, &referral)

	fmt.Println("JSON Unmarshalled")	
	if err != nil {
		fmt.Println(err.Error())
	}
	
	return err, referral
}

func (t *ReferralPartnerChaincodeBroker) marshallReferral(referral CustomerReferral) (error, []byte) {
	fmt.Println("Marshalling JSON to bytes")
	valAsbytes, err := json.Marshal(referral)
	
	if err != nil {
		fmt.Println("Marshalling JSON to bytes failed")
		return err, nil
	}
	
	return nil, valAsbytes
}

func (t *ReferralPartnerChaincodeBroker) getReferralPartner(stub *shim.ChaincodeStub, args []string) (error, string) {
	var err error
	var valAsBytes []byte
	var keyValue string
	
	valAsBytes, err = stub.InvokeChaincode(t.RetailChaincode, "Query", args)

	if valAsBytes != nil && err == nil{
		keyValue = BytesToString(valAsBytes)
		if strings.Contains(keyValue, "Did not find entry") == false {
			return nil, t.RetailChaincode
		}
	}
	
	valAsBytes, err = stub.InvokeChaincode(t.CommercialChaincode, "Query", args)
	
	if valAsBytes != nil && err == nil {
		keyValue = BytesToString(valAsBytes)
		if strings.Contains(keyValue, "Did not find entry") == false {
			return nil, t.CommercialChaincode
		}
	}
	
	valAsBytes, err = stub.InvokeChaincode(t.BankingChaincode, "Query", args)
	
	if valAsBytes != nil && err == nil {
		keyValue = BytesToString(valAsBytes)
		if strings.Contains(keyValue, "Did not find entry") == false {
			return nil, t.BankingChaincode
		}
	}
		
	if err != nil {
		fmt.Println(err.Error())
	}
	
	return err, ""
}

// updateReferral - invoke function to updateReferral key/value pair
func (t *ReferralPartnerChaincodeBroker) updateReferralStatus(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var value string
	var valAsbytes []byte
	var err error
	
	fmt.Println("running updateReferral()")
	
	err, value = t.getReferralPartner(stub, args)
	
	if value == t.RetailChaincode {
		valAsbytes, err = stub.InvokeChaincode(t.RetailChaincode, "Invoke", args)
	} else if value == t.CommercialChaincode {
		valAsbytes, err = stub.InvokeChaincode(t.CommercialChaincode, "Invoke", args)
	} else if value == t.BankingChaincode {
		valAsbytes, err = stub.InvokeChaincode(t.BankingChaincode, "Invoke", args)
	}
	
	return valAsbytes, err
}

// createReferral - invoke function to write key/value pair
func (t *ReferralPartnerChaincodeBroker) createReferral(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var referralData string
	var err error
	var valAsbytes []byte
	
	fmt.Println("running createReferral()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2 parameters, name of the key and value to set")
	}

	referralData = args[1]
	
	var referral CustomerReferral

	err = json.Unmarshal([] byte(referralData), &referral)
	
	if err != nil {
		return []byte("Count not unmarshall payload data: " + referralData + " for the ledger"), err
	}
	
	// Create a ledger record that indexes the referral id by the partner
	for i := range referral.Departments {
	    if referral.Departments[i] == "RETAIL" {
			fmt.Println("Running Retail create chaincode for payload: " + referralData)
			valAsbytes, err = stub.InvokeChaincode(t.RetailChaincode, "Invoke", args)
			
		} else if referral.Departments[i] == "COMMERCIAL" {
			valAsbytes, err = stub.InvokeChaincode(t.CommercialChaincode, "Invoke", args)
		} else if referral.Departments[i] == "BANKING" {
			valAsbytes, err = stub.InvokeChaincode(t.BankingChaincode, "Invoke", args)
		}
		
		if err != nil {
			return []byte("Count not index the bytes by department from the value: " + referralData + " on the ledger"), err
		}
	}
	
	if err != nil {
				fmt.Println("Error Running Retail create chaincode for payload: " + referralData)
				fmt.Println(err.Error())
			}
		
	return valAsbytes, nil
}

func (t *ReferralPartnerChaincodeBroker) processCommaDelimitedReferrals(delimitedReferrals string, stub *shim.ChaincodeStub) ([]byte, error) {
	commaDelimitedReferrals := strings.Split(delimitedReferrals, ",")

	referralResultSet := "["
	
	for i := range commaDelimitedReferrals {
		valAsbytes, err := stub.GetState(commaDelimitedReferrals[i])
		
		if err != nil {
			return nil, err
		}
		
		if i == 0 {
			referralResultSet = referralResultSet + BytesToString(valAsbytes)
		} else {
			referralResultSet = referralResultSet + "," + BytesToString(valAsbytes)
		}
	}
	
	referralResultSet += "]"
	return []byte(referralResultSet), nil
}

func (t *ReferralPartnerChaincodeBroker) searchByPartner(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	valAsbytes, err := stub.InvokeChaincode(args[0], "Query", args)
	
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
	
	return valAsbytes, nil
}

func (t *ReferralPartnerChaincodeBroker) searchByStatus(args []string, stub *shim.ChaincodeStub) ([]byte, error) {
	var matchingRetailReferrals, matchingCommercialReferrals, matchingBankingReferrals, searchResults []CustomerReferral
	var valAsbytes []byte
	var err error
	
	
	valAsbytes, err = stub.InvokeChaincode(t.RetailChaincode, "Invoke", args)
	err = json.Unmarshal(valAsbytes, &matchingRetailReferrals)
	
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
	
	valAsbytes, err = stub.InvokeChaincode(t.CommercialChaincode, "Invoke", args)
	err = json.Unmarshal(valAsbytes, &matchingCommercialReferrals)
	
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
	
	valAsbytes, err = stub.InvokeChaincode(t.BankingChaincode, "Invoke", args)
	err = json.Unmarshal(valAsbytes, &matchingBankingReferrals)
	
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
	
	searchResults = append(searchResults, matchingRetailReferrals...)
	
	for i := range matchingCommercialReferrals {
		matchFound := false
		for j := range searchResults {
		    if matchingCommercialReferrals[i].ReferralId == searchResults[j].ReferralId  {
				matchFound = true
			}
		}
		
		if matchFound == false {
			searchResults = append(searchResults, matchingCommercialReferrals[i])
		}
	}

	for i := range matchingBankingReferrals {
		matchFound := false
		for j := range searchResults {
		    if matchingBankingReferrals[i].ReferralId == searchResults[j].ReferralId  {
				matchFound = true
			}
		}
		
		if matchFound == false {
			searchResults = append(searchResults, matchingBankingReferrals[i])
		}
	}
		
	if(err != nil) {
		return nil, err
	}
	
	return valAsbytes, nil
}


// read - query function to read key/value pair
func (t *ReferralPartnerChaincodeBroker) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error
	var value string
	var valAsbytes []byte
	
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}
	
	err, value = t.getReferralPartner(stub, args)

	if value == t.RetailChaincode {
		valAsbytes, err = stub.InvokeChaincode(t.RetailChaincode, "Query", args)
	} else if value == t.CommercialChaincode {
		valAsbytes, err = stub.InvokeChaincode(t.CommercialChaincode, "Query", args)
	} else if value == t.BankingChaincode {
		valAsbytes, err = stub.InvokeChaincode(t.BankingChaincode, "Query", args)
	}
	
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return []byte(jsonResp), err
	}
	
	if valAsbytes == nil {
		return []byte("Did not find entry for key: " + key), nil
	}
	
	return valAsbytes, nil
}