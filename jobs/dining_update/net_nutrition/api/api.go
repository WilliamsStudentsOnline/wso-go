package api

import (
	"errors"
	"net/http"
	"regexp"
	"fmt"
)

const WilliamsDiningBaseUrl = "https://nutrition.williams.edu/NetNutrition/1"

var RegexpGetId = regexp.MustCompile(`[\D]+(?P<id>\d+)[\D]+`)

type WilliamsDiningAPI struct {
	baseUrl string
	client  *http.Client
}

func CreateWilliamsDiningAPI() (*WilliamsDiningAPI, error) {
	sessionIdCookie, err := getSessionIdCookie()
	if err != nil {
		return nil, err
	}

	return &WilliamsDiningAPI{
		baseUrl: WilliamsDiningBaseUrl,
		client: &http.Client{
			Transport: NewNetNutTransport(sessionIdCookie),
		},
	}, nil
}

func getSessionIdCookie() (string, error) {
    	fmt.Println("getSessionIdCookie: Starting...")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
    			fmt.Println("getSessionIdCookie: Starting...")
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, WilliamsDiningBaseUrl, nil)
	if err != nil {
            	fmt.Println("getSessionIdCookie: Redirect intercepted, using last response")
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
        	fmt.Printf("getSessionIdCookie: Error creating request: %v\n", err)
		return "", err
	}
	
    	t.Logf("getSessionIdCookie: Found %d cookies\n", len(resp.Cookies()))
	for _, cookie := range resp.Cookies() {
        	fmt.Printf("  Cookie #%d: Name=%s, Value=%s\n", i+1, cookie.Name, cookie.Value)
		if cookie.Name == "ASP.NET_SessionId" {
            		fmt.Println("getSessionIdCookie: Found ASP.NET_SessionId cookie!")
			return cookie.Value, nil
		}
	}

	return "", errors.New("missing sessionId cookie")
}

type NetNutTransport struct {
	underlyingTransport http.RoundTripper
	sessionIDCookie     string
}

func NewNetNutTransport(sessionIDCookie string) *NetNutTransport {
	return &NetNutTransport{
		underlyingTransport: http.DefaultTransport,
		sessionIDCookie:     sessionIDCookie,
	}
}

func (t *NetNutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Cookie", "CBORD.netnutrition2=NNexternalID=1&Layout=; ASP.NET_SessionId="+t.sessionIDCookie)
	return t.underlyingTransport.RoundTrip(req)
}
