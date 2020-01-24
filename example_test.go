package lg_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/neilotoole/lg"
	"github.com/neilotoole/lg/loglg"
)

func Example_businessOperationV1() {
	log := loglg.NewWith(os.Stdout, false, true, false)

	receipt, err := BusinessOperationV1(log)
	if err != nil {
		log.Errorf(err.Error())
		fmt.Println("Failure:", err)
	} else {
		fmt.Println("Success:", receipt)
	}

	// Output:
	// Success: RECEIPT_ABC123
}

// BusinessOperationV1 performs a business operation against
// an external API. If the business operation fails, a non-nil
// error is returned. If the business operation succeeds,
// a non-empty transaction receipt is returned.
// BusinessOperationV1 closes dataSource via defer, but does ignores
// any error from Close.
func BusinessOperationV1(log lg.Log) (receipt string, err error) {
	dataSource, err := OpenBizData()
	if err != nil {
		return "", err
	}
	defer dataSource.Close() // Ignores any error from Close

	data, err := ioutil.ReadAll(dataSource)
	if err != nil {
		return "", err
	}

	return ExternalAPICall(data)
}

func Example_businessOperationV2() {
	log := loglg.NewWith(os.Stdout, false, true, false)

	receipt, err := BusinessOperationV2(log)
	if err != nil {
		log.Errorf(err.Error())
		fmt.Println("Failure:", err)
	} else {
		fmt.Println("Success:", receipt)
	}

	// Output:
	// WARN 	failed to close due to gremlins
	// Success: RECEIPT_ABC123
}

// BusinessOperationV2 closes dataSource in a defer, and logs at WARN level
// if an error results from Close.
func BusinessOperationV2(log lg.Log) (receipt string, err error) {
	dataSource, err := OpenBizData()
	if err != nil {
		return "", err
	}
	defer func() {
		err := dataSource.Close()
		if err != nil {
			log.Warnf(err.Error())
		}
	}()

	data, err := ioutil.ReadAll(dataSource)
	if err != nil {
		return "", err
	}

	return ExternalAPICall(data)
}

func Example_businessOperationV3() {
	log := loglg.NewWith(os.Stdout, false, true, false)

	receipt, err := BusinessOperationV3(log)
	if err != nil {
		log.Errorf(err.Error())
		fmt.Println("Failure:", err)
	} else {
		fmt.Println("Success:", receipt)
	}

	// Output:
	// WARN 	failed to close due to gremlins
	// Success: RECEIPT_ABC123
}

// BusinessOperationV3 uses WarnIfError to make the defer statement
// more succinct.
func BusinessOperationV3(log lg.Log) (receipt string, err error) {
	dataSource, err := OpenBizData()
	if err != nil {
		return "", err
	}
	defer func() {
		log.WarnIfError(dataSource.Close())
	}()

	data, err := ioutil.ReadAll(dataSource)
	if err != nil {
		return "", err
	}

	return ExternalAPICall(data)
}

func Example_businessOperationV4() {
	log := loglg.NewWith(os.Stdout, false, true, false)

	receipt, err := BusinessOperationV4(log)
	if err != nil {
		log.Errorf(err.Error())
		fmt.Println("Failure:", err)
	} else {
		fmt.Println("Success:", receipt)
	}

	// Output:
	// WARN 	failed to close due to gremlins
	// Success: RECEIPT_ABC123
}

// BusinessOperationV4 uses WarnIfFnError to make the defer statement
// yet more succinct.
func BusinessOperationV4(log lg.Log) (receipt string, err error) {
	dataSource, err := OpenBizData()
	if err != nil {
		return "", err
	}
	defer log.WarnIfFnError(dataSource.Close)

	data, err := ioutil.ReadAll(dataSource)
	if err != nil {
		return "", err
	}

	return ExternalAPICall(data)
}

// ExternalAPICall is a mock external API invocation. It always
// returns a non-empty receipt and nil err.
func ExternalAPICall(a []byte) (receipt string, err error) {
	return "RECEIPT_ABC123", nil
}

// OpenBizData returns a ReadCloser whose Close method
// always returns an error.
func OpenBizData() (*dataSource, error) {
	return &dataSource{bytes.NewReader([]byte("TOKEN_XYZ456"))}, nil
}

// dataSource is an io.ReadCloser whose Close
// method always returns an error.
type dataSource struct {
	io.Reader
}

func (d *dataSource) Close() error {
	return errors.New("failed to close due to gremlins")
}
