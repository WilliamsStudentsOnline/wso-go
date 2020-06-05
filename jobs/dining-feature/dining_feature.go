package main

import(
	"fmt"
	// "strings"
	// "log"
	// "io"
	"io/ioutil"
	// "os"
	"net/http"
	
	// "golang.org/x/net/html"
)


func main(){
	parsedHTTP, err := getXML("https://dining.williams.edu/eats4ephs/")
	fmt.Errorf("Error: &v", err)
	fmt.Println(parsedHTTP)

	
}

func getXML(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", fmt.Errorf("GET error: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("Status error: %v", resp.StatusCode)
    }

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("Read body: %v", err)
    }

    return string(data), nil
}
