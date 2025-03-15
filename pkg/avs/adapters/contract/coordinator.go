package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters/txmgr"

	"github.com/Layr-Labs/eigenlayer-cli/pkg/avs/adapters"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Coordinator struct {
	repo   adapters.Repository
	logger logging.Logger
	spec   Specification
	config adapters.Configuration
	client *ethclient.Client
	cache  map[string]interface{}
	TxMgr  *txmgr.TxManager
	dryRun bool
}

type CoordinatorConfiguration struct {
	EthRPCURL string `json:"eth_rpc_url" validate:"required"`
}

type PrivateKeyStoreConfiguration struct {
	PrivateKeyStorePath     string `json:"private_key_store_path"     validate:"required"`
	PrivateKeyStorePassword string `json:"private_key_store_password"`
}

func NewCoordinator(
	repo adapters.Repository,
	logger logging.Logger,
	spec Specification,
	config adapters.Configuration,
	dryRun bool,
) (*Coordinator, error) {
	cfg := &CoordinatorConfiguration{}
	err := config.Unmarshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	client, err := ethclient.Dial(cfg.EthRPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the Ethereum client: %v", err)
	}

	cache := map[string]interface{}{}
	txMgr, err := txmgr.NewTxManager(spec.BaseSpecification, config, logger, dryRun)
	if err != nil {
		return nil, fmt.Errorf("failed to create tx manager: %v", err)
	}
	coordinator := Coordinator{repo, logger, spec, config, client, cache, txMgr, dryRun}

	if err := coordinator.initialize(); err != nil {
		return nil, err
	}

	return &coordinator, nil
}

func (c Coordinator) Client() (*ethclient.Client, error) {
	if c.client == nil {
		url, err := c.load("config:eth_rpc_url", "string", true)
		if err != nil {
			return nil, err
		}

		client, err := ethclient.Dial(url.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to the eth rpc service: %v", err)
		}

		c.client = client
	}

	return c.client, nil
}

func (c Coordinator) initialize() error {
	// Load the cache with values from specification
	spec, err := c.repo.LoadResource(c.spec.Name, "avs.json")
	if err != nil {
		return err
	}

	values := map[string]interface{}{}
	if err := json.Unmarshal(spec, &values); err != nil {
		return err
	}

	for k, v := range values {
		c.store("spec:"+k, v)
	}

	// Load the cache with values from configuration
	for k, v := range c.config.GetAll() {
		c.store("config:"+k, v)
	}

	return nil
}

func (c Coordinator) store(key string, value interface{}) {
	unboxed, ok := value.(map[string]interface{})
	if ok {
		for k, v := range unboxed {
			c.store(key+"."+k, v)
		}
	} else {
		c.logger.Debug(fmt.Sprintf("Cache: store %s = %+v", key, value))
		c.cache[key] = value
	}
}

func (c Coordinator) load(key string, kind string, required bool) (interface{}, error) {
	c.logger.Debug(fmt.Sprintf("Cache: load %s (%s)", key, kind))

	var err error
	value, exist := c.cache[key]
	if !exist {
		transform := ""
		tokens := strings.Split(key, "->")
		if len(tokens) == 2 {
			key = tokens[0]
			transform = tokens[1]
		} else if len(tokens) != 1 {
			return nil, fmt.Errorf("invalid key format: %s", key)
		}

		if strings.HasPrefix(key, "call:") {
			payload := key[len("call:"):]
			value, err = c.Call(payload, nil)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(key, "config:") {
			payload := key[len("config:"):]
			value, err = c.config.Prompt(payload, required, false)
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(key, "const:") {
			payload := key[len("const:"):]
			value = payload
		} else if strings.HasPrefix(key, "func:") {
			payload := key[len("func:"):]
			tokens := strings.FieldsFunc(payload, func(r rune) bool {
				return r == '(' || r == ',' || r == ')'
			})

			for i := range tokens {
				tokens[i] = strings.TrimSpace(tokens[i])
			}

			function := tokens[0]

			params := make(map[string]string)
			for _, token := range tokens[1:] {
				tokens := strings.Split(token, "=")
				if len(tokens) != 2 {
					return nil, fmt.Errorf("malformed function parameter: %s", token)
				}

				params[strings.TrimSpace(tokens[0])] = strings.TrimSpace(tokens[1])
			}

			c.logger.Debug(fmt.Sprintf("Executing function [Function=%s, Parameters=%+v]", function, params))
			switch function {
			case "array_uint8":
				result, err := array_uint8(&c, params)
				if err != nil {
					return nil, err
				}

				value = result

			case "bls_curve_sign":
				result, err := bls_curve_sign(&c, params)
				if err != nil {
					return nil, err
				}

				value = result

			case "bls_sign":
				result, err := bls_sign(&c, params)
				if err != nil {
					return nil, err
				}

				value = result

			case "chain_id":
				result, err := chain_id(&c, params)
				if err != nil {
					return nil, err
				}

				value = result

			case "ecdsa_public_key":
				result, err := ecdsa_public_key(&c, params)
				if err != nil {
					return nil, err
				}

				value = result

			case "bls_public_key":
				result, err := bls_public_key(&c, params)
				if err != nil {
					return nil, err
				}

				value = result

			case "ecdsa_sign":
				result, err := ecdsa_sign(&c, params)
				if err != nil {
					return nil, err
				}

				value = *result

			case "expiry":
				result, err := expiry(&c, params)
				if err != nil {
					return nil, err
				}

				value = result

			case "salt":
				result, err := salt(&c, params)
				if err != nil {
					return nil, err
				}

				value = result

			default:
				return nil, fmt.Errorf("invalid function: %s [%s]", key, function)
			}
		} else if strings.HasPrefix(key, "passwd:") {
			payload := key[len("passwd:"):]
			value, err = c.config.Prompt(payload, required, true)
			if err != nil {
				return nil, err
			}

			c.store(key, value)
		}

		if transform != "" {
			result, err := c.transform(value, transform)
			if err != nil {
				return nil, err
			}

			value = result
		}
	}

	if value != nil {
		switch kind {
		case "address":
			switch vt := value.(type) {
			case string:
				_ = vt
				if common.IsHexAddress(value.(string)) {
					value = common.HexToAddress(value.(string))
				}
			}

		default:
			result, err := convert(value, kind)
			if err != nil {
				return nil, err
			}

			value = result
		}
	}

	c.logger.Debug(fmt.Sprintf("Cache: load %s (%s) = %+v (%T)", key, kind, value, value))

	if required && value == nil {
		return nil, fmt.Errorf("required value missing: %s", key)
	}

	return value, nil
}

func (c Coordinator) transform(value interface{}, transform string) (interface{}, error) {
	var fields []reflect.StructField
	var values []interface{}

	tokens := strings.FieldsFunc(transform, func(r rune) bool {
		return r == '(' || r == ',' || r == ')'
	})

	for i := range tokens {
		tokens[i] = strings.TrimSpace(tokens[i])
	}

	switch tokens[0] {
	case "[]byte":
		var result []byte
		reflected := reflect.ValueOf(value)
		for i := 0; i < reflected.NumField(); i++ {
			result = append(result, []byte(fmt.Sprintf("%x", reflected.Field(i)))...)
		}

		return result, nil

	case "struct":
		source, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("struct transformation not allowed on type: %T", source)
		}

		for _, entry := range tokens[1:] {
			tokens := strings.Split(entry, ":")
			if len(tokens) != 2 {
				return nil, fmt.Errorf("invalid transformation: %s", entry)
			}

			value, ok := source[tokens[1]]
			if !ok {
				return nil, fmt.Errorf("invalid transformation: %s", entry)
			}

			fields = append(fields, reflect.StructField{
				Name: tokens[0],
				Type: reflect.TypeOf(value),
			})

			values = append(values, value)
		}

		result := reflect.New(reflect.StructOf(fields)).Elem()
		for i, v := range fields {
			result.FieldByName(v.Name).Set(reflect.ValueOf(values[i]))
		}

		return result.Addr().Interface(), nil

	default:
		return nil, fmt.Errorf("invalid transformation: %s", tokens[0])
	}
}

func (c Coordinator) mapParameters(parameters []string, inputs []abi.Argument) ([]interface{}, error) {
	values := []interface{}{}
	for i, parameter := range parameters {
		value, err := c.load(parameter, inputs[i].Type.String(), true)
		if err != nil {
			return nil, err
		}

		c.logger.Debug(fmt.Sprintf(
			"Parameter mapped [Key=%s Name=%s, Value=%+v, Type=%T]",
			parameter,
			inputs[i].Name,
			value,
			value,
		))

		values = append(values, value)
	}

	return values, nil
}

func (c Coordinator) transactAndWaitForMinted(
	contractAddr string,
	contract *bind.BoundContract,
	contractArgs []abi.Argument,
	functionName string,
	functionParameters []string,
) error {
	// Prepare input parameters
	contractFunctionParameters, err := c.mapParameters(
		functionParameters,
		contractArgs,
	)
	if err != nil {
		return fmt.Errorf("failed to map parameters for contract function %s: %w", functionName, err)
	}

	c.logger.Debug(
		fmt.Sprintf("Invoking contract transaction [Address=%s, Function=%s]", contractAddr, functionName),
	)
	for i := 0; i < len(contractFunctionParameters); i++ {
		c.logger.Debug(fmt.Sprintf(
			"Parameter [Name=%s, Value=%+v, Type=%T]",
			contractArgs[i].Name,
			contractFunctionParameters[i],
			contractFunctionParameters[i],
		))
	}
	tx, receipt, err := c.TxMgr.CallAndWaitForReceipt(
		context.Background(),
		func(txOpts *bind.TransactOpts) (*types.Transaction, error) {
			return contract.Transact(txOpts, functionName, contractFunctionParameters...)
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create transaction for %s of contract %s: %w", functionName, contractAddr, err)
	}

	if c.dryRun {
		c.logger.Debug(fmt.Sprintf("Transaction created: %s", tx.Hash().Hex()))
	} else {
		c.logger.Debug(fmt.Sprintf("Transaction created: %s", tx.Hash().Hex()), "receipt", receipt.Status)
	}

	return nil
}

func (c Coordinator) Execute(function string) error {
	c.logger.Debug(fmt.Sprintf("Executing %s", function))

	contractFunction, ok := c.spec.Functions[function]
	if !ok {
		return fmt.Errorf("not supported")
	}

	if contractFunction.Name == "" {
		if contractFunction.Message != "" {
			fmt.Println(contractFunction.Message)
		} else {
			fmt.Println("not supported")
		}

		return nil
	}

	c.logger.Debug(fmt.Sprintf("Loading ABI [Specification=%s, File=%s]", c.spec.Name, c.spec.ABI))
	abi, err := c.loadContractABI(c.spec.Name, c.spec.ABI)
	if err != nil {
		return err
	}

	c.logger.Debug(fmt.Sprintf("Binding contract [Address=%s]", c.spec.ContractAddress))
	contract := c.bind(c.spec.ContractAddress, abi)
	if contract == nil {
		return fmt.Errorf("failed to bind contract: %s", c.spec.ContractAddress)
	}

	tokens := strings.Split(contractFunction.Name, ".")
	if len(tokens) == 1 { // Direct function
		return c.transactAndWaitForMinted(
			c.spec.ContractAddress,
			contract,
			abi.Methods[contractFunction.Name].Inputs,
			contractFunction.Name,
			contractFunction.Parameters,
		)
	} else if len(tokens) == 2 { // Delegate function
		i := slices.IndexFunc(c.spec.Delegates, func(d Delegate) bool { return d.Name == tokens[0] })
		if i == -1 {
			return fmt.Errorf("invalid delegate: %s", tokens[0])
		}

		delegate := c.spec.Delegates[i]

		var delegateAddress string
		if delegate.ContractAddress != "" {
			delegateAddress = delegate.ContractAddress
		} else {
			address, err := c.callContractFunction(contract, c.spec.ContractAddress, tokens[0], nil)
			if err != nil {
				return err
			}

			delegateAddress = address[0].(common.Address).String()
		}

		delegateABI, err := c.loadContractABI(c.spec.Name, delegate.ABI)
		if err != nil {
			return err
		}

		delegateContract := c.bind(delegateAddress, delegateABI)
		if delegateContract == nil {
			return fmt.Errorf("failed to bind contract: %s", delegateAddress)
		}

		i = slices.IndexFunc(delegate.Functions, func(f Function) bool { return f.Name == tokens[1] })
		if i == -1 {
			return fmt.Errorf("invalid delegate function: %s", tokens[1])
		}

		delegateFunction := delegate.Functions[i]

		return c.transactAndWaitForMinted(
			delegateAddress,
			delegateContract,
			delegateABI.Methods[tokens[1]].Inputs,
			delegateFunction.Name,
			delegateFunction.Parameters)
	}

	return fmt.Errorf("invalid function name: %s", contractFunction.Name)
}

func (c Coordinator) loadContractABI(spec string, resource string) (*abi.ABI, error) {
	data, err := c.repo.LoadResource(spec, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to read abi: %s", resource)
	}

	abi, err := abi.JSON(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse abi: %s", resource)
	}

	return &abi, nil
}

func (c Coordinator) bind(address string, abi *abi.ABI) *bind.BoundContract {
	client, err := c.Client()
	if err != nil {
		return nil
	}

	return bind.NewBoundContract(common.HexToAddress(address), *abi, client, client, client)
}

func (c Coordinator) callContractFunction(
	contract *bind.BoundContract,
	address string,
	name string,
	params []interface{},
) ([]interface{}, error) {
	c.logger.Debug(
		fmt.Sprintf(
			"Calling contract function [Address=%s, Function=%s, Parameters=%+v]",
			address,
			name,
			params,
		),
	)

	var result []interface{}
	err := contract.Call(
		&bind.CallOpts{Context: context.Background()},
		&result,
		name,
		params...,
	)
	if err != nil {
		return nil, err
	}

	c.logger.Debug(fmt.Sprintf("Contract function invoked [Response=%+v]", result))
	return result, nil
}

func (c Coordinator) Call(function string, params *[]string) (interface{}, error) {
	c.logger.Debug(fmt.Sprintf("Calling %s", function))

	c.logger.Debug(fmt.Sprintf("Loading ABI [Specification=%s, File=%s]", c.spec.Name, c.spec.ABI))
	abi, err := c.loadContractABI(c.spec.Name, c.spec.ABI)
	if err != nil {
		return nil, err
	}

	c.logger.Debug(fmt.Sprintf("Binding contract [Address=%s]", c.spec.ContractAddress))
	contract := c.bind(c.spec.ContractAddress, abi)
	if contract == nil {
		return nil, fmt.Errorf("failed to bind contract: %s", c.spec.ContractAddress)
	}

	tokens := strings.Split(function, ".")
	if len(tokens) == 1 { // Direct function
		parameters, err := c.mapParameters(*params, abi.Methods[function].Inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to map parameters for contract function %s: %w", function, err)
		}

		output, err := c.callContractFunction(contract, c.spec.ContractAddress, function, parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to call function: %s", function)
		}

		c.logger.Debug(fmt.Sprintf("Invocation complete [Response=%+v]", output))
		return output[0], nil

	} else if len(tokens) == 2 { // Delegate function
		i := slices.IndexFunc(c.spec.Delegates, func(d Delegate) bool { return d.Name == tokens[0] })
		if i == -1 {
			return nil, fmt.Errorf("invalid delegate: %s", tokens[0])
		}

		delegate := c.spec.Delegates[i]
		delegateABI, err := c.loadContractABI(c.spec.Name, delegate.ABI)
		if err != nil {
			return nil, err
		}

		var delegateAddress string
		if delegate.ContractAddress != "" {
			delegateAddress = delegate.ContractAddress
		} else {
			address, err := c.callContractFunction(contract, c.spec.ContractAddress, tokens[0], nil)
			if err != nil {
				return nil, err
			}

			delegateAddress = address[0].(common.Address).String()
		}

		delegateContract := c.bind(delegateAddress, delegateABI)
		if delegateContract == nil {
			return nil, fmt.Errorf("failed to bind contract: %s", delegateAddress)
		}

		i = slices.IndexFunc(delegate.Functions, func(f Function) bool { return f.Name == tokens[1] })
		if i == -1 {
			return nil, fmt.Errorf("invalid delegate function: %s", function)
		}

		delegateFunction := delegate.Functions[i]
		parameters, err := c.mapParameters(delegateFunction.Parameters, delegateABI.Methods[tokens[1]].Inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to map parameters for contract function %s: %w", delegateFunction.Name, err)
		}

		output, err := c.callContractFunction(delegateContract, delegateAddress, tokens[1], parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to call delegate function: %s", function)
		}

		if delegateFunction.Transform != "" {
			return c.transform(output[0], delegateFunction.Transform)
		}

		return output[0], nil
	}

	return nil, fmt.Errorf("invalid function: %s", function)
}

func (c Coordinator) Type() string {
	return "contract"
}

func (c Coordinator) Register() error {
	return c.Execute("register")
}

func (c Coordinator) OptIn() error {
	return c.Execute("opt-in")
}

func (c Coordinator) OptOut() error {
	return c.Execute("opt-out")
}

func (c Coordinator) Deregister() error {
	return c.Execute("deregister")
}

func (c Coordinator) Status() (int, error) {
	function, ok := c.spec.Functions["status"]
	if !ok {
		return -1, fmt.Errorf("not supported")
	}

	status, err := c.Call(function.Name, &function.Parameters)
	if err != nil {
		return -1, err
	}

	result, err := convert(status, "int")
	if err != nil {
		return -1, err
	}

	return result.(int), nil
}
