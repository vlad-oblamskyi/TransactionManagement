package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strings"
	b64 "encoding/base64"
	"regexp"
	"encoding/json"
	//"github.com/hyperledger/fabric/core/util"
	//"math/big"
	//"time"
)

const KVS_HANLDER_KEY = "KVS_HANDLER_KEY"
const MT199_TEMPLATE = "{1:F01[[SENDER]]XXXX0000000000}{2:I199[[RECEIVER]]XXXXN}{4:\r\n" +
	               ":20:HL-[[TX-ID]]\r\n" +
	               ":79:[[COMMENT]]\r\n" +
	               "-}";

type TransactionManagement struct {
}

type Details struct {
	InputMessage  string  `json:"inputMessage"`
	OutputMessage string  `json:"outputMessage"`
}

type AccountKey struct {
	HolderBIC string  `json:"holderBic"`
	OwnerBIC  string  `json:"ownerBic"`
	Currency  string  `json:"currency"`
	Type      string  `json:"type"`
}

type AccountValue struct {
	Amount    string  `json:"amount"`
	Currency  string  `json:"currency"`
	Type      string  `json:"type"`
	Date      string  `json:"date"`
	Number    string  `json:"number"`
}

type Organization struct {
	BIC string
	Account string
}

type Transaction struct {
	TransactionId string            `json:"transactionId"`
	Sender Organization             `json:"sender"`
	Receiver Organization           `json:"receiver"`
	SenderAccountKey AccountKey     `json:"senderAccountKey"`
	ReceiverAccountKey AccountKey   `json:"receiverAccountKey"`
	Fee string                      `json:"fee"`
	Amount string                   `json:"amount"`
	TransactionDetails Details      `json:"details"`
	Status string                   `json:"status"`
	Comment string                  `json:"comment"`
	Time string                     `json:"time"`
}

type PermissionAccountKey struct {
	Type         string  `json:"type"`
	Holder       string  `json:"holder"`
	Owner        string  `json:"owner"`
	Currency     string  `json:"currency"`
	AccountType  string  `json:"accountType"`
}

type Permission struct {
	Key      PermissionAccountKey  `json:"accountKey"`
	Access   string  `json:"access"`
}

type UserDetails struct {
	Password     string  `json:"password"`
	Permissions  []Permission  `json:"permissions"`
}


func (t *TransactionManagement) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. KVS chaincode id is expected");
	}
	stub.PutState(KVS_HANLDER_KEY, []byte(args[0]))

	return nil, nil
}

func (t *TransactionManagement) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	case "transfer":
		if len(args) != 2 {
			return nil, errors.New("Incorrect number of arguments. 2 parameters are expected:  authToken, MT message");
		}
		token := args[0]
		state, _ := stub.GetState(KVS_HANLDER_KEY)
		mapId := string(state);
		byteMtMessage, _ := b64.StdEncoding.DecodeString(string(args[1]))
		mtMessage := string(byteMtMessage)

		// Parse MT Message
		senderAccountKey := AccountKey {
			HolderBIC: getIntermediaryBIC(mtMessage),
			OwnerBIC: getSender(mtMessage),
			Currency: getTransferCurrency(mtMessage),
			Type: "nostro",
		}

		receiverAccountKey := AccountKey {
			HolderBIC: getIntermediaryBIC(mtMessage),
			OwnerBIC: getReceiver(mtMessage),
			Currency: getTransferCurrency(mtMessage),
			Type: "vostro",
		}

		senderOrganization := Organization {
			BIC: getSender(mtMessage),
			Account: getCredAccount(mtMessage),
		}

		test, _ := json.Marshal(senderAccountKey)
		test, _ = json.Marshal(receiverAccountKey)
		test, _ = json.Marshal(senderOrganization)


		strings.Replace(token, "","",-1)
		strings.Replace(mapId, "","",-1)

		return nil, errors.New("RESULT: " + string(test));

		//receiverOrganization := Organization {
		//	BIC: getIntermediaryBIC(mtMessage),
		//	Account: getBenAccount(mtMessage),
		//}

		//
		//transaction := &Transaction {
		//	TransactionId: stub.GetTxID(),
		//	Sender: senderOrganization,
		//	Receiver:  receiverOrganization,
		//	SenderAccountKey: senderAccountKey,
		//	ReceiverAccountKey: receiverAccountKey,
		//	Fee: getTransferFee(mtMessage),
		//	Amount: getTransferAmount(mtMessage),
		//	TransactionDetails: Details { InputMessage: mtMessage },
		//}
		//
		//test, _ := json.Marshal(transaction)
		//return nil, errors.New("RESULT: " + string(test));
		//
		//// Validate transaction
		//transaction.Status = "Success"
		//
		//var account AccountValue
		//jsonSenderAccountKey, _ := json.Marshal(transaction.SenderAccountKey)
		//queryArgs := util.ToChaincodeArgs("function", string(jsonSenderAccountKey))
		//queryResult, _ := stub.QueryChaincode(mapId, queryArgs)
		//if err := json.Unmarshal(queryResult, &account); err != nil {
		//	transaction.Status = "Failure"
		//	transaction.Comment = "Unable to get sender account from the KVS"
		//}
		//
		//jsonUserKey, _ := b64.StdEncoding.DecodeString(token)
		//queryArgs = util.ToChaincodeArgs("function", string(jsonUserKey))
		//queryResult, _ = stub.QueryChaincode(mapId, queryArgs)
		//var userDetails UserDetails
		//if err := json.Unmarshal(queryResult, &userDetails); err != nil {
		//	transaction.Status = "Failure"
		//	transaction.Comment = "Unable to get user by the token"
		//}
		//var allowTransfer bool
		//for i := 0; i < len(userDetails.Permissions); i ++ {
		//	if userDetails.Permissions[i].Key.Currency == transaction.SenderAccountKey.Currency &&
		//	   userDetails.Permissions[i].Key.Holder == transaction.SenderAccountKey.HolderBIC &&
		//	   userDetails.Permissions[i].Key.Owner == transaction.SenderAccountKey.OwnerBIC &&
		//	   userDetails.Permissions[i].Key.AccountType == transaction.SenderAccountKey.Type {
		//		allowTransfer = true
		//	}
		//}
		//if !allowTransfer {
		//	transaction.Status = "Failure"
		//	transaction.Comment = "User doesn't have the permission for the action"
		//}
		//currentAmount := new(big.Rat)
		//currentAmount.SetString(account.Amount)
		//fee := new(big.Rat)
		//fee.SetString(transaction.Fee)
		//transferableAmount := new(big.Rat)
		//transferableAmount.SetString(transaction.Amount)
		//subtotal := new(big.Rat).Sub(currentAmount, transferableAmount)
		//subtotal.Sub(subtotal, fee)
		//if (subtotal.Sign() == -1) {
		//	transaction.Status = "Failure"
		//	transaction.Comment = "Unable to transfer the requested amount"
		//}
		//
		//// Prepare output message
		//newAmount := new(big.Rat).Sub(transferableAmount, fee)
		//var outputMessage string
		//if (transaction.Status == "Success") {
		//	outputMessage := mtMessage
		//	outputMessage = strings.Replace(outputMessage, getReceiver(mtMessage), getIntermediaryBIC(mtMessage), -1)
		//	outputMessage = strings.Replace(outputMessage, getSender(mtMessage), getReceiver(mtMessage), -1)
		//	outputMessage = strings.Replace(outputMessage, ":57A:" + getIntermediaryBIC(mtMessage), ":52A:" + getSender(mtMessage), -1)
		//	outputMessage = strings.Replace(outputMessage, strings.Replace(transaction.Amount, ".", ",", -1), strings.Replace(newAmount.String(), ".", ",", -1), -1)
		//	outputMessage = strings.Replace(outputMessage, ":71G:" + transaction.SenderAccountKey.Currency + strings.Replace(transaction.Fee, ".", ",", -1) , "", -1)
		//} else {
		//	outputMessage := MT199_TEMPLATE
		//	outputMessage = strings.Replace(outputMessage, "[[SENDER]]", getReceiver(mtMessage), -1)
		//	outputMessage = strings.Replace(outputMessage, "[[RECEIVER]]", getSender(mtMessage), -1)
		//	outputMessage = strings.Replace(outputMessage, "[[TX-ID]]", transaction.TransactionId, -1)
		//	outputMessage = strings.Replace(outputMessage, "[[COMMENT]]", transaction.Comment, -1)
		//}
		//transaction.TransactionDetails.OutputMessage = outputMessage
		//transaction.Time = time.Now().UTC().Format(time.RFC3339)
		//
		//if (transaction.Status == "Success") {
		//	var sender AccountValue
		//	jsonSenderAccountKey, _ = json.Marshal(transaction.SenderAccountKey)
		//	senderQueryArgs := util.ToChaincodeArgs("function", string(jsonSenderAccountKey))
		//	queryResult, _ = stub.QueryChaincode(mapId, senderQueryArgs)
		//	if err := json.Unmarshal(queryResult, &sender); err != nil {
		//		panic(err)
		//	}
		//	var receiver AccountValue
		//	jsonReceiverAccountKey, _ := json.Marshal(transaction.ReceiverAccountKey)
		//	receiverQueryArgs := util.ToChaincodeArgs("function", string(jsonReceiverAccountKey))
		//	queryResult, _ = stub.QueryChaincode(mapId, receiverQueryArgs)
		//	if err := json.Unmarshal(queryResult, &receiver); err != nil {
		//		panic(err)
		//	}
		//	currentAmount := new(big.Rat)
		//	currentAmount.SetString(sender.Amount)
		//	fee := new(big.Rat)
		//	fee.SetString(transaction.Fee)
		//	transferableAmount := new(big.Rat)
		//	transferableAmount.SetString(transaction.Amount)
		//	receiverAmount := new(big.Rat)
		//	receiverAmount.SetString(receiver.Amount)
		//
		//	currentAmount.Sub(currentAmount, transferableAmount).Sub(currentAmount, fee)
		//	receiverAmount.Add(receiverAmount, transferableAmount)
		//
		//	sender.Amount = currentAmount.String()
		//	receiver.Amount = receiverAmount.String()
		//
		//	jsonSenderAccountKey, _ = json.Marshal(transaction.SenderAccountKey)
		//	jsonSender, _ := json.Marshal(sender)
		//	senderInvokeArgs := util.ToChaincodeArgs("put", string(jsonSenderAccountKey), string(jsonSender))
		//	stub.InvokeChaincode(mapId, senderInvokeArgs)
		//
		//	jsonReceiverAccountKey, _ = json.Marshal(transaction.ReceiverAccountKey)
		//	jsonReceiver, _ := json.Marshal(receiver)
		//	receiverInvokeArgs := util.ToChaincodeArgs("put", string(jsonReceiverAccountKey), string(jsonReceiver))
		//	stub.InvokeChaincode(mapId, receiverInvokeArgs)
		//}
		//
		//jsonTransaction, _ := json.Marshal(transaction)
		//invokeArgs := util.ToChaincodeArgs("put", transaction.TransactionId, string(jsonTransaction))
		//stub.InvokeChaincode(mapId, invokeArgs)

		return nil, nil
	default:
		return nil, errors.New("Unsupported operation")
	}
}

func (t *TransactionManagement) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	switch function {
	default:
		return nil, errors.New("Unsupported operation")
	}
}

func getBlock(mtMessage string, blockNumber int) string {
	if blockNumber != 4 {
		r, _ := regexp.Compile("{" + string(blockNumber) + ":(.*?)}");
		block := r.FindString(mtMessage)
		block = strings.Replace(block, "{" + string(blockNumber) + ":", "", -1)
		block = strings.Replace(block, "}", "", -1)

		return block
	} else {
		r, _ := regexp.Compile("{4:(\\s*.*)+-}");
		block := r.FindString(mtMessage)
		block = strings.Replace(block, "{4:", "", -1)
		block = strings.Replace(block, "-}", "", -1)

		return block
	}
}

func getTag(block4 string, tagName string) string {
	r, _ := regexp.Compile(":" + tagName + ":(\\s*.*)+?(\\s:|$)");
	tag := r.FindString(block4)
	tag = strings.Replace(tag, ":" + tagName + ":", "", -1)
	tag = strings.Replace(tag, "\r", "", -1)
	tag = strings.Replace(tag, "\n", "", -1)
	tag = strings.TrimSuffix(tag, ":")

	return tag
}

func getReceiver(mtMessage string) string {
	block2 := getBlock(mtMessage, 2)
	if block2 != "" {
		if (len(block2) == 17 || len(block2) == 21) {
			return block2[4:12]
		} else if len(block2) == 47 {
			return block2[14:22]
		}
	}
	return ""
}

func getSender(mtMessage string) string {
	block1 := getBlock(mtMessage, 1)
	if block1 != "" {
		return block1[3:11]
	}
	return ""
}

func getTransferAmount(mtMessage string) string {
	block4 := getBlock(mtMessage, 4)
	if block4 != "" {
		tag := getTag(block4, "32A")
		amount := tag[9:]
		amount = strings.Replace(amount, ",", ".", -1)
		return amount;
	}
	return ""
}

func getTransferCurrency(mtMessage string) string {
	block4 := getBlock(mtMessage, 4)
	if block4 != "" {
		tag := getTag(block4, "32A")
		currency := tag[6:9]
		return currency;
	}
	return ""
}

func getTransferFee(mtMessage string) string {
	block4 := getBlock(mtMessage, 4)
	if block4 != "" {
		tag := getTag(block4, "71G")
		amount := tag[3:]
		amount = strings.Replace(amount, ",", ".", -1)
		return amount;
	}
	return ""
}

func getCredAccount(mtMessage string) string {
	block4 := getBlock(mtMessage, 4)
	if block4 != "" {
		tag := getTag(block4, "50K")
		return nil, errors.New("RESULT: " + tag);
		account := tag[1:strings.Index(tag, "\n")]
		return account;
	}
	return ""
}

func getBenAccount(mtMessage string) string {
	block4 := getBlock(mtMessage, 4)
	if block4 != "" {
		tag := getTag(block4, "59A")
		account := tag[1:strings.Index(tag, "\n")]
		return account;
	}
	return ""
}

func getIntermediaryBIC(mtMessage string) string {
	block4 := getBlock(mtMessage, 4)
	if block4 != "" {
		tag := getTag(block4, "57A")
		return tag;
	}
	return ""
}

func main() {
	err := shim.Start(new(TransactionManagement))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}